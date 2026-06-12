# dependencies.map.md
# project: cementer
# updated: 2026-06-12T09:02:13-06:00  commit: ee446c3

Module: github.com/bryanmaclee/cementer  (Go 1.26.4)

## Runtime Dependencies (Go, direct)
github.com/gorilla/websocket@v1.5.3 — WebSocket server (upgrade, ping/pong, message frames) for /ws/live
go.bug.st/serial@v1.7.1 — cross-platform serial port I/O; the production line source
modernc.org/sqlite@v1.52.0 — pure-Go SQLite driver (CGO-free → static ARM cross-compile); the durable store

## Indirect Dependencies (Go, // indirect)
github.com/dustin/go-humanize@v1.0.1 — pulled by modernc/sqlite
github.com/google/uuid@v1.6.0 — pulled by modernc/sqlite
github.com/mattn/go-isatty@v0.0.20 — TTY detection (transitive)
github.com/ncruces/go-strftime@v1.0.0 — strftime support for modernc/sqlite
github.com/remyoudompheng/bigfft@v0.0.0-20230129092748 — math support for modernc
golang.org/x/sys@v0.43.0 — low-level OS syscalls (serial / sqlite)
modernc.org/libc@v1.72.3 — pure-Go libc shim under modernc/sqlite
modernc.org/mathutil@v1.7.1 — math helpers for modernc
modernc.org/memory@v1.11.0 — memory allocator for modernc

## Web Dev / Build Dependencies (web/package.json)
Package: cementer-web (private, type: module). No runtime dependencies — zero framework.
typescript@^6.0.3 — type checking (`tsc`) before bundling
vite@^8.0.16 — dev server (with WS/debug proxy to :8080) and production bundler → web/dist

## Internal Module Graph
cmd/cementer/main.go → hub, model, parser, rawlog, serialreader, source, store (+ root `cementer` for WebDist)
internal/parser → internal/model
internal/store → internal/model
internal/serialreader → internal/source (implements LineSource; uses go.bug.st/serial)
internal/source → (stdlib only)
internal/rawlog → (stdlib only)
internal/hub → (stdlib `context` only)
internal/model → (stdlib `time` only)
web: main.ts → readout.ts, ws.ts ; readout.ts → theme.ts, types.ts ; ws.ts → types.ts

## Notes
- CGO is never used (`CGO_ENABLED=0` everywhere) — the pure-Go SQLite is what makes a single static ARM binary possible (`make pi`).
- internal/api is an empty package directory; nothing imports it.

## Tags
#cementer #map #dependencies #go-modules #sqlite #websocket #serial #vite

## Links
- [primary.map.md](./primary.map.md)
- [structure.map.md](./structure.map.md)
- [build.map.md](./build.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
