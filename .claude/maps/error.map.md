# error.map.md
# project: cementer
# updated: 2026-06-12T09:02:13-06:00  commit: ee446c3

No custom error types are defined. Errors are handled with stdlib `errors.New`, `fmt.Errorf` wrapping (`%w`), and Go's standard `if err != nil` pattern.

## Notable Errors Raised
errors.New("provide -serial <device> or -source <replay-file>")  — cmd/cementer/main.go:58 — neither source flag given at startup
fmt.Errorf("data dir: %w", err)        — main.go:71 — cannot create data directory
fmt.Errorf("open store: %w", err)      — main.go:94 — SQLite open/schema-init failed
fmt.Errorf("open raw log: %w", err)    — main.go:101 — raw log file open failed
fmt.Errorf("open source: %w", err)     — main.go:120 — serial/replay source open failed
fmt.Errorf("web assets: %w", err)      — main.go:283 — embedded web/dist unavailable
os.ErrClosed (returned)                — internal/rawlog/rawlog.go:52 — Append after Close

## Error Handling Patterns
Startup fatal: main() → run() returns error → log.Fatalf("cementer: %v") — any wiring error exits the process. [main.go:39]
Per-line tolerance: parser.Parse never errors — returns (Reading, ok=false) for blank/comment/garbage; bad fields are silently omitted. [internal/parser/parser.go]
Best-effort logging (keep running): rawlog append failure → log.Printf, continue; store commit failure → fmt.Printf("store: commit failed..."), data still safe in raw log, keep ingesting. [main.go:111, store.go:105]
WebSocket pumps: any read/write error returns from the pump goroutine → readPump's deferred Unregister cleans up. Upgrade failure returns silently (response already written). [main.go:210-266]
Source goroutine: src.Run error logged only if ctx not cancelled (clean shutdown is not an error). [main.go:128]

## Resilience / Drop Policy (by design, not errors)
hub.Broadcast: non-blocking send; if hub overloaded the live message is DROPPED (data already durable). [internal/hub/hub.go:91]
hub.Run broadcast: a subscriber whose buffered Send is full is DROPPED and closed — ingestion is never blocked by a slow client. [hub.go:62]
store.Submit: blocks under backpressure but NEVER drops while open (store is source of truth); silently drops only after Close. [store.go:85]

## Global Error Boundaries
HTTP: /debug/stats returns 500 via http.Error on Stats() failure. [main.go:271] No global HTTP error middleware (3-route mux).
Client: web/src/ws.ts swallows malformed WS frames in a try/catch; no React-style ErrorBoundary (vanilla TS). [web/src/ws.ts:32]
Shutdown: ordered teardown with a 5s timeout context — server first, then source, then store, then rawlog. [main.go:158]

## Unhandled Error Risks
Several Close()/Sync() calls are deliberately ignored (`_ =`) on the shutdown path — acceptable for a process about to exit, but a Sync error before a crash would be silent. [store.go, rawlog.go, main.go]
store commit failure leaves rows uncommitted to SQLite (recoverable from raw log only by manual re-import; no automatic retry). [store.go:103]

## Tags
#cementer #map #error #go-errors #durability #drop-policy #graceful-shutdown

## Links
- [primary.map.md](./primary.map.md)
- [state.map.md](./state.map.md)
- [events.map.md](./events.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
