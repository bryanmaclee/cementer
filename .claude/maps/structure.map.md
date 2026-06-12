# structure.map.md
# project: cementer
# updated: 2026-06-12T09:02:13-06:00  commit: ee446c3

## Entry Points
cmd/cementer/main.go: the single binary's `main`/`run` — parses flags, wires the pipeline (source → rawlog → parser → store → hub), mounts HTTP (`/ws/live`, `/debug/stats`, `/`), and serves the embedded web client. The whole product is this one process.
assets.go (package `cementer`, repo root): `go:embed all:web/dist` + `WebDist()` — embeds the built web client into the binary.
web/src/main.ts: browser client entry — instantiates `Readout`, opens the live WebSocket via `connectLive`.

## Directory Ownership
cmd/cementer/        — binary entrypoint: flag parsing, pipeline wiring, WebSocket pumps, embedded-SPA mounting, debug stats.
internal/source/     — `LineSource` interface + `Replay` (dev file replay, with loop). Decouples ingest from where bytes come from.
internal/serialreader/ — production `LineSource`: reads newline ASCII off a USB-serial port via `go.bug.st/serial`.
internal/rawlog/     — durability layer 1: append-only raw-line file with periodic fsync. Captures every byte before parse/store.
internal/parser/     — the ONLY protocol-specific code: ASCII line → `model.Reading`. Permissive (skips blank/comment/garbage). Has the only Go test.
internal/model/      — shared data contracts: `Reading` (broadcast unit) and `Sample` (long-form storage row).
internal/store/      — durability layer 2: SQLite (modernc pure-Go) WAL single-writer; batch commit; `onCommit` post-commit hook; `Stats`.
internal/hub/        — WebSocket fan-out: `Hub` + `Subscriber`; drops slow clients rather than blocking ingestion. Transport-agnostic.
internal/api/        — EMPTY directory (no files). No HTTP handler package; routes live in cmd/cementer/main.go.
web/                 — vanilla-TS + Vite client (no framework). `index.html`, configs; built to `web/dist` (git-ignored, embedded at build).
web/src/             — client modules: main, readout (live value cards), ws (reconnecting socket), theme (dark/light), types, styles.css.
web/src/chart/       — EMPTY directory (no files). Reserved for the future uPlot charting centerpiece (Phase 4, not built).
deploy/              — `cementer.service` systemd unit template for the Pi.
testdata/            — `sample-stream.txt`: synthetic comma-separated stream for `make run` (no pump needed).
docs/design/         — `data-model.md`: normative configuration-driven design (pump profiles, DAQ formats, recording segments) — largely DESIGNED, not yet built.
docs/deep-dives/     — `storage-and-viz-architecture-2026-06-12.md`: architecture decision research (Go/SQLite vs Influx/Grafana). See non-compliance report.
docs/pa/             — project-management source-of-truth docs: status, changelog, hand-off, anti-patterns, design-insights, user-voice.
docs/changes/        — change-log artifact dir (currently only `.gitkeep`).
esp32sketches/       — collaborator test-rig: ESP32 `.ino` sketches, real Enbridge CSVs, `send_csv.py`. A dev/diagnostic bench, NOT shipped product source.
pi4b & test db/      — collaborator bench README (Influx/Grafana stack notes + plaintext test creds). NOT product source. See non-compliance report.

## Ignored / Generated Paths
.git, web/node_modules, web/dist (go:embed-built), /data/, *.db / *.db-wal / *.db-shm, raw-*.log, /cementer binary, .claude

## Notes
- ~10 Go files, 5 TypeScript files, 51 git-tracked files total. Not a monorepo (single Go module + one co-located web client).
- Root-level `dumb_file` is a 16-byte scratch file with no role in the build.

## Tags
#cementer #map #structure #go #vanilla-ts #single-binary #raspberry-pi #daq

## Links
- [primary.map.md](./primary.map.md)
- [dependencies.map.md](./dependencies.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
