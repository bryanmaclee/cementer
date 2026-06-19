# dependencies.map.md
# project: cementer
# updated: 2026-06-19T23:05:55Z  commit: 1465bd9

Module: github.com/bryanmaclee/cementer  (Go 1.26.4)

## Runtime Dependencies (Go, direct)
`github.com/gorilla/websocket@v1.5.3` — WebSocket upgrade, ping/pong, write deadlines for `/ws/live`
`go.bug.st/serial@v1.7.1` — cross-platform USB-serial port I/O; production `LineSource`
`modernc.org/sqlite@v1.52.0` — pure-Go SQLite driver (CGO-free); the durable store; enables static arm64 cross-compile

## Indirect Dependencies (Go)
`github.com/dustin/go-humanize@v1.0.1` — pulled by modernc/sqlite
`github.com/google/uuid@v1.6.0` — pulled by modernc/sqlite
`github.com/mattn/go-isatty@v0.0.20` — TTY detection (transitive)
`github.com/ncruces/go-strftime@v1.0.0` — strftime for modernc/sqlite
`github.com/remyoudompheng/bigfft@v0.0.0-20230129092748` — math for modernc
`golang.org/x/sys@v0.43.0` — OS syscalls (serial / sqlite)
`modernc.org/libc@v1.72.3` — pure-Go libc shim for modernc/sqlite
`modernc.org/mathutil@v1.7.1` — math helpers
`modernc.org/memory@v1.11.0` — memory allocator for modernc

## Web Runtime Dependencies (web/package.json)
`uplot@^1.6.32` — lightweight charting library; live-rolling + per-job historical charts; bundled offline (no CDN)

## Web Dev / Build Dependencies
`typescript@^6.0.3` — type checking (`tsc`) before bundling
`vite@^8.0.16` — dev server + production bundler → `web/dist`

## Internal Module Graph (Go)
```
cmd/cementer/main.go
  → cementer (root package, WebDist)
  → internal/api
  → internal/daqformat
  → internal/hub
  → internal/model
  → internal/rawlog
  → internal/serialreader
  → internal/source
  → internal/store

internal/api      → internal/store
internal/daqformat → internal/model
internal/store    → internal/model
internal/serialreader → (go.bug.st/serial + stdlib)
internal/source   → (stdlib only)
internal/rawlog   → (stdlib only)
internal/hub      → (context stdlib only)
internal/model    → (time stdlib only)
internal/parser   → internal/model  [OFF main path; test-only]
```

## Internal Module Graph (TypeScript)
```
web/src/main.ts → readout.ts, controls.ts, ws.ts
web/src/readout.ts → chart/livechart.ts, chart/jobchart.ts, chart/config.ts, theme.ts, types.ts
web/src/controls.ts → types.ts
web/src/ws.ts → types.ts
web/src/chart/livechart.ts → chart/roles.ts, chart/config.ts, types.ts
web/src/chart/jobchart.ts → chart/roles.ts, types.ts
web/src/chart/roles.ts → types.ts
web/src/chart/config.ts → (localStorage + stdlib only)
```

## Notes
- `CGO_ENABLED=0` everywhere — the pure-Go SQLite is what makes a static arm64 binary possible (`make pi`).
- `internal/parser` is decoupled from the main pipeline (off-path); nothing in main.go imports it. Only its test imports it.
- `uPlot` is bundled into `web/dist` at build time; the Pi runs fully offline with no CDN dependency.

## Tags
#cementer #map #dependencies #go-modules #sqlite #websocket #serial #vite #uplot

## Links
- [primary.map.md](./primary.map.md)
- [structure.map.md](./structure.map.md)
- [build.map.md](./build.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
