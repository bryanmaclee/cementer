// Command cementer is the single Pi-side binary: it reads ASCII lines from the
// pump (a USB-serial port or, in development, a replay file), captures every line
// to an append-only raw log, parses and durably stores readings in SQLite, and
// fans the committed readings out to browser clients over WebSocket. It also
// serves the embedded dark-mode web client.
//
// Pipeline (see docs/plan): source -> rawlog -> parser -> store -> (post-commit) hub.
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
	"github.com/bryanmaclee/cementer/internal/hub"
	"github.com/bryanmaclee/cementer/internal/model"
	"github.com/bryanmaclee/cementer/internal/parser"
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
	)
	flag.Parse()

	if *serialPort == "" && *replayPath == "" {
		return errors.New("provide -serial <device> or -source <replay-file>")
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

	// --- raw log (durability layer 1) ---
	rawPath := filepath.Join(dir, "raw-"+time.Now().Format("20060102-150405")+".log")
	rl, err := rawlog.Open(rawPath, time.Second)
	if err != nil {
		return fmt.Errorf("open raw log: %w", err)
	}

	p := parser.New(parser.DefaultConfig())

	// handleLine is the head of the pipeline: capture raw bytes first (so nothing
	// is ever lost), then parse and submit for durable storage.
	handleLine := func(line []byte) {
		if err := rl.Append(line); err != nil {
			log.Printf("rawlog append: %v", err)
		}
		if r, ok := p.Parse(line, time.Now()); ok {
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

	// --- HTTP: WebSocket, debug stats, embedded SPA ---
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws/live", serveWS(h))
	mux.HandleFunc("GET /debug/stats", serveStats(st))
	if err := mountSPA(mux); err != nil {
		return err
	}

	srv := &http.Server{Addr: *addr, Handler: mux}
	go func() {
		log.Printf("cementer listening on %s  (source: %s, db: %s, raw: %s)", *addr, srcDesc, dbPath, rawPath)
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

// wsEnvelope is the message shape sent to clients. The Type field leaves room for
// future message kinds (job updates, log entries, ...).
type wsEnvelope struct {
	Type    string         `json:"type"`
	Reading *model.Reading `json:"reading,omitempty"`
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

func serveWS(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return // Upgrade already wrote an error response.
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
