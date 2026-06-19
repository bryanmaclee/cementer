// Command cementer is the single Pi-side binary: it reads ASCII lines from the
// pump (a USB-serial port or, in development, a replay file), captures every line
// to an append-only raw log, parses and durably stores readings in SQLite, and
// fans the committed readings out to browser clients over WebSocket. It also
// serves the embedded dark-mode web client.
//
// Pipeline (see docs/design/data-model.md): source -> rawlog -> daqformat engine
// -> store -> (post-commit) hub.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	cementer "github.com/bryanmaclee/cementer"
	"github.com/bryanmaclee/cementer/internal/api"
	"github.com/bryanmaclee/cementer/internal/daqformat"
	"github.com/bryanmaclee/cementer/internal/hub"
	"github.com/bryanmaclee/cementer/internal/model"
	"github.com/bryanmaclee/cementer/internal/rawlog"
	"github.com/bryanmaclee/cementer/internal/serialreader"
	"github.com/bryanmaclee/cementer/internal/source"
	"github.com/bryanmaclee/cementer/internal/store"

	"github.com/gorilla/websocket"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("cementer: %v", err)
	}
}

func run() error {
	var (
		addr           = flag.String("addr", ":8080", "HTTP listen address")
		serialPort     = flag.String("serial", "", "serial device path (e.g. /dev/serial/by-id/...); leave empty to use -source")
		baud           = flag.Int("baud", 9600, "serial baud rate")
		replayPath     = flag.String("source", "", "replay file of ASCII lines (dev source); overridden by -serial")
		replayInterval = flag.Duration("replay-interval", 250*time.Millisecond, "delay between replayed lines")
		replayLoop     = flag.Bool("replay-loop", true, "loop the replay file when exhausted")
		dataDir        = flag.String("data-dir", "", "directory for the SQLite DB and raw logs (default $CEMENTER_DATA_DIR or ./data)")
		batchInterval  = flag.Duration("batch-interval", 250*time.Millisecond, "SQLite commit / live-broadcast cadence")
		formatID       = flag.String("format", "intellisense", "DAQ format preset: intellisense | synthetic")
	)
	flag.Parse()

	if *serialPort == "" && *replayPath == "" {
		return errors.New("provide -serial <device> or -source <replay-file>")
	}

	format, err := resolveFormat(*formatID)
	if err != nil {
		return err
	}

	// Storage location is trivially flippable: -data-dir, else $CEMENTER_DATA_DIR,
	// else ./data. Dev uses the Pi's built-in storage; prod points at an SSD.
	dir := *dataDir
	if dir == "" {
		dir = os.Getenv("CEMENTER_DATA_DIR")
	}
	if dir == "" {
		dir = "./data"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("data dir: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// --- hub (fan-out) ---
	h := hub.New()
	go h.Run(ctx)

	// onCommit runs after each batch is durably committed: serialize the reading
	// and broadcast it. Clients only ever see what is already stored.
	onCommit := func(r model.Reading) {
		b, err := json.Marshal(wsEnvelope{Type: "reading", Reading: &r})
		if err != nil {
			return
		}
		h.Broadcast(b)
	}

	// --- store (durability layer 2) ---
	dbPath := filepath.Join(dir, "cementer.db")
	st, err := store.Open(dbPath, *batchInterval, onCommit)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}

	// Seed the active Pump Profile on first run (axiom #3 — the pump self-describes).
	// The store is format-agnostic (axiom #2): main supplies the seed vocabulary from
	// the active DaqFormat. The seed is idempotent — a reboot against the same data
	// dir must NOT duplicate the profile or clobber operator edits.
	vocab := seedVocab(format)
	if has, err := st.HasActiveProfile(); err != nil {
		return fmt.Errorf("check active profile: %w", err)
	} else if !has {
		name := format.Name + " (this pump)"
		if err := st.SeedActiveProfile(name, defaultUnits(format), format.ID, vocab); err != nil {
			return fmt.Errorf("seed profile: %w", err)
		}
		log.Printf("seeded active pump profile %q (%d channels) from format %q", name, len(vocab), format.ID)
	}

	// --- raw log (durability layer 1) ---
	rawPath := filepath.Join(dir, "raw-"+time.Now().Format("20060102-150405")+".log")
	rl, err := rawlog.Open(rawPath, time.Second)
	if err != nil {
		return fmt.Errorf("open raw log: %w", err)
	}

	eng := daqformat.New(format)

	// handleLine is the head of the pipeline: capture raw bytes first (so nothing
	// is ever lost), then map the line through the (config-driven) format engine
	// and submit for durable storage. Raw capture is never gated on the parse.
	handleLine := func(line []byte) {
		if err := rl.Append(line); err != nil {
			log.Printf("rawlog append: %v", err)
		}
		if r, ok := eng.Apply(line, time.Now()); ok {
			st.Submit(r)
		}
	}

	// --- source (serial in production, replay in dev) ---
	src, srcDesc, err := buildSource(*serialPort, *baud, *replayPath, *replayInterval, *replayLoop)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}

	var srcWG sync.WaitGroup
	srcWG.Add(1)
	go func() {
		defer srcWG.Done()
		if err := src.Run(ctx, handleLine); err != nil && ctx.Err() == nil {
			log.Printf("source stopped: %v", err)
		} else {
			log.Printf("source finished")
		}
	}()

	// --- HTTP: WebSocket, debug stats, profile API, embedded SPA ---
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws/live", serveWS(h, st.ActiveProfile))
	mux.HandleFunc("GET /debug/stats", serveStats(st))
	// Profile CRUD over HTTP. Handlers call store methods only (axiom #4 / D2). The
	// reset escape hatch reseeds from the active format's vocab, which main supplies
	// so the api package stays format-agnostic.
	api.New(st, func() []store.SeedChannel { return seedVocab(format) }).Register(mux)
	if err := mountSPA(mux); err != nil {
		return err
	}

	srv := &http.Server{Addr: *addr, Handler: mux}
	go func() {
		log.Printf("cementer listening on %s  (format: %s, source: %s, db: %s, raw: %s)", *addr, format.ID, srcDesc, dbPath, rawPath)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("http server: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Printf("shutting down...")

	// Ordered shutdown: stop accepting clients, stop the source, then flush the
	// durable layers. The source must stop before the store closes so no Submit
	// races the store shutdown.
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutCtx)
	srcWG.Wait()
	_ = src.Close()
	if err := st.Close(); err != nil {
		log.Printf("store close: %v", err)
	}
	if err := rl.Close(); err != nil {
		log.Printf("rawlog close: %v", err)
	}
	log.Printf("bye")
	return nil
}

// resolveFormat maps a -format preset id to its DaqFormat. Adapting to a new pump
// format is configuration (a new preset value), not a code edit elsewhere.
func resolveFormat(id string) (daqformat.DaqFormat, error) {
	switch id {
	case "intellisense":
		return daqformat.Intellisense(), nil
	case "synthetic":
		return daqformat.Synthetic(), nil
	default:
		return daqformat.DaqFormat{}, fmt.Errorf("unknown -format %q (want: intellisense | synthetic)", id)
	}
}

// seedVocab returns the default channel vocabulary for a format, converted to the
// store's format-agnostic SeedChannel shape. main owns this mapping so internal/store
// never imports internal/daqformat (axiom #2). An unknown format yields an empty
// vocab (resolveFormat already rejects unknown ids before this is reached).
func seedVocab(format daqformat.DaqFormat) []store.SeedChannel {
	var chans []daqformat.Channel
	switch format.ID {
	case "intellisense":
		chans = daqformat.IntellisenseChannels()
	case "synthetic":
		chans = daqformat.SyntheticChannels()
	}
	out := make([]store.SeedChannel, 0, len(chans))
	for _, c := range chans {
		out = append(out, store.SeedChannel{
			ID:        c.ID,
			Role:      c.Role,
			Scope:     c.Scope,
			UnitIndex: c.UnitIndex,
			Label:     c.Label,
			UoM:       c.UoM,
			Decimals:  c.Decimals,
		})
	}
	return out
}

// defaultUnits is the seed value for a profile's pumping-unit count. The rig this Pi
// describes is single-unit for both shipped formats; the operator bumps it for a
// two-unit pump via PUT /api/profile.
func defaultUnits(format daqformat.DaqFormat) int {
	switch format.ID {
	case "intellisense", "synthetic":
		return 1
	default:
		return 1
	}
}

func buildSource(serialPort string, baud int, replayPath string, interval time.Duration, loop bool) (source.LineSource, string, error) {
	if serialPort != "" {
		cfg := serialreader.DefaultConfig(serialPort)
		cfg.BaudRate = baud
		r, err := serialreader.Open(cfg)
		if err != nil {
			return nil, "", err
		}
		return r, fmt.Sprintf("serial %s @ %d", serialPort, baud), nil
	}
	r, err := source.NewReplayFile(replayPath, interval, loop)
	if err != nil {
		return nil, "", err
	}
	return r, fmt.Sprintf("replay %s every %s (loop=%v)", replayPath, interval, loop), nil
}

// wsEnvelope is the message shape sent to clients. The Type field discriminates the
// kinds: "reading" (the live sample frame) and "profile" (the hello/profile greeting
// sent ONCE per connection). It leaves room for future kinds (job/recording updates).
// The TypeScript mirror lives in web/src/types.ts — keep them in sync by hand.
type wsEnvelope struct {
	Type    string         `json:"type"`
	Reading *model.Reading `json:"reading,omitempty"`
	Profile *store.Profile `json:"profile,omitempty"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	// LAN deployment: accept any origin. Tighten when the network posture changes.
	CheckOrigin: func(*http.Request) bool { return true },
}

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

// serveWS upgrades to WebSocket and, on connect, sends the per-connection
// hello/profile greeting (the active profile's ENABLED channels) BEFORE starting the
// live-reading write pump. The profile is a greeting, not a broadcast, so it is
// written directly to this conn — never routed through hub.Broadcast. activeProfile
// is the store accessor; it returns (profile, ok, err).
func serveWS(h *hub.Hub, activeProfile func() (store.Profile, bool, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return // Upgrade already wrote an error response.
		}

		// hello/profile greeting: write one {type:"profile"} frame to THIS conn.
		// If there is no active profile (shouldn't happen post-seed) we send nothing
		// and log; this never gates the live readout (axiom #1).
		if p, ok, err := activeProfile(); err != nil {
			log.Printf("ws: load active profile: %v", err)
		} else if ok {
			if b, err := json.Marshal(wsEnvelope{Type: "profile", Profile: &p}); err == nil {
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
					_ = conn.Close()
					return
				}
			}
		} else {
			log.Printf("ws: no active profile to greet client with")
		}

		sub := hub.NewSubscriber(256)
		h.Register(sub)
		go writePump(conn, sub)
		go readPump(conn, h, sub)
	}
}

// writePump drains the subscriber's Send channel to the socket and sends periodic
// pings. It exits when the hub closes Send (client dropped/unregistered).
func writePump(conn *websocket.Conn, sub *hub.Subscriber) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = conn.Close()
	}()
	for {
		select {
		case msg, ok := <-sub.Send:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump detects client disconnects (and drains any client->server frames). On
// any read error it unregisters the subscriber, which closes its Send channel and
// stops writePump.
func readPump(conn *websocket.Conn, h *hub.Hub, sub *hub.Subscriber) {
	defer h.Unregister(sub)
	conn.SetReadLimit(4096)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func serveStats(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := st.Stats()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(s)
	}
}

func mountSPA(mux *http.ServeMux) error {
	dist, err := cementer.WebDist()
	if err != nil {
		return fmt.Errorf("web assets: %w", err)
	}
	fileServer := http.FileServerFS(dist)
	mux.Handle("GET /", spaFallback(dist, fileServer))
	return nil
}

// spaFallback serves static files, falling back to index.html for unknown paths so
// client-side routes resolve (not needed in Phase 1, but harmless and forward-looking).
func spaFallback(dist fs.FS, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			if _, err := fs.Stat(dist, filepath.ToSlash(r.URL.Path[1:])); err != nil {
				r.URL.Path = "/"
			}
		}
		next.ServeHTTP(w, r)
	})
}
