# Anti-pattern briefing — author-in-language

Read before writing any cementer code; re-read before each feature. Counters training-data bias toward
other languages/frameworks. Two parts: **Go** (server) and **vanilla-TS** (web client).

---

## Part A — idiomatic Go (the `internal/*`, `cmd/*` server)

Counter the Java/Python/JS reflexes:

- **Don't add interfaces speculatively.** Accept interfaces, return structs. Define an interface where
  it's *consumed*, not next to the implementation. `internal/source.LineSource` is the right shape (one
  consumer-side seam); don't grow a parallel interface for every package.
- **Handle every error; don't swallow.** No bare `_ =` on errors that matter. The pipeline logs and
  continues on a per-line parse failure (correct — durability first), but a store/open failure is fatal.
  Match that discipline.
- **No getters/setters for plain data.** Exported fields on `model.Reading` are fine. Don't write
  `GetValue()`.
- **No goroutine leaks.** Every goroutine has a clear exit (context cancel, channel close). The hub's
  writePump/readPump and the store's writeLoop are the pattern — a new goroutine must show its exit.
- **Single-writer discipline for the store is sacred.** All DB writes funnel through the one writer
  goroutine (WAL, `SetMaxOpenConns(1)`). Never open a second writer or write from a handler.
- **Keep protocol-specific code in `internal/parser` only.** Per project axiom #2, format adaptation is
  config, not code — don't scatter format assumptions into store/hub/web.
- **`gofmt` + `go vet` clean, always.** Errors wrapped with `%w` and context (`fmt.Errorf("open store:
  %w", err)`), as `main.go` already does.
- **Prefer the standard library + the three deps** (`gorilla/websocket`, `go.bug.st/serial`,
  `modernc.org/sqlite` — pure-Go, no CGO). Don't add a framework; don't reach for an ORM.

## Part B — vanilla-TS web client (`web/src/*`)

**This client uses NO framework.** Counter the React/Vue/Svelte/Angular reflex hard:

- **No framework, no JSX, no virtual DOM, no component library.** It's plain TypeScript modules + Vite
  + direct DOM. `main.ts`, `readout.ts`, `theme.ts`, `ws.ts`, `types.ts` are plain modules.
- **No state-management library** (Redux/Zustand/Pinia). State is plain module variables + DOM.
- **Render dynamically from the stream/profile**, not from a hard-coded channel set (project axiom #1 +
  the config-driven model). The client builds fields from the channels the Pi describes.
- **localStorage for personal prefs only** (theme, live-chart view). Pump definitions live on the Pi,
  never in the client (project axiom #3).
- **Keep it embeddable.** Output goes to `web/dist` and is `go:embed`'d; don't add a runtime CDN
  dependency or anything that breaks the single-binary, offline-on-the-Pi guarantee.
- **TypeScript strict.** Type the WS envelope (`types.ts`) to match `cmd/cementer/main.go`'s
  `wsEnvelope` — keep the two in sync by hand (no codegen).
