# primary.map.md
# project: cementer
# updated: 2026-06-12T09:02:13-06:00  commit: ee446c3

## Project Fingerprint
Language:   Go 1.26.4 (module github.com/bryanmaclee/cementer) + TypeScript (vanilla, ES2022) client
Framework:  none — net/http stdlib server + gorilla/websocket; vanilla-TS + Vite client (no UI framework)
Runtime:    single static CGO-free binary; targets Raspberry Pi 4B (arm64), offline
Type:       embedded full-stack data-acquisition appliance (serial → durable store → live WebSocket readout)
Size:       ~51 git-tracked files (10 Go, 5 TS)

## Map Index
| Map                  | Status  | Contents                                          |
|----------------------|---------|---------------------------------------------------|
| structure.map.md     | present | directory layout, entry points, ownership         |
| dependencies.map.md  | present | 3 direct Go deps, 2 web build deps, module graph  |
| schema.map.md        | present | Go/TS contracts + 1 SQLite table (`samples`)      |
| config.map.md        | present | 8 flags, 1 env var (CEMENTER_DATA_DIR), no .env   |
| build.map.md         | present | Makefile + npm scripts, go:embed pipeline, no CI  |
| error.map.md         | present | no custom error types; stdlib + drop-policy       |
| test.map.md          | present | go testing, 1 test file (parser), 2 tests         |
| api.map.md           | present | 3 HTTP routes + parser line→Reading contract      |
| state.map.md         | present | ingest→store→broadcast pipeline; thin client state|
| events.map.md        | present | in-process hub fan-out, 1 WS topic ("reading")    |
| infra.map.md         | present | single binary + systemd on offline Pi             |
| style.map.md         | present | CSS custom-property theming (dark default)        |
| auth.map.md          | absent  | no auth in code (designed for Phase 3)            |
| domain.map.md        | absent  | concepts live in data-model.md (mostly unbuilt)   |
| i18n.map.md          | absent  | none                                              |
| migrations.map.md    | absent  | schema is inline CREATE IF NOT EXISTS, no tool    |
| jobs.map.md          | absent  | no scheduler/worker/cron                          |

## File Routing
types / contracts / SQLite schema       → schema.map.md
HTTP routes / WS + parser contract      → api.map.md
flags / env / config files              → config.map.md
test patterns / fixtures                → test.map.md
build commands / cross-compile / embed  → build.map.md
directory layout / entry points         → structure.map.md
external packages / module graph        → dependencies.map.md
ingest pipeline / store / client state  → state.map.md
hub fan-out / WS message topics         → events.map.md
error handling / drop policy            → error.map.md
deploy / systemd / Pi runtime layout    → infra.map.md
theming / CSS tokens / display inference → style.map.md

## Key Facts
- Entry point cmd/cementer/main.go wires the whole pipeline in run(): source (serial OR replay file) → rawlog (append-only, durability layer 1) → parser → store (SQLite WAL single-writer, durability layer 2) → hub → WebSocket. The web client is embedded via go:embed (assets.go) so the product ships as one file.
- The store has EXACTLY ONE table: `samples` (id, ts_us, channel, value) + a ts index. Jobs, recording_segments, pump_profiles, and DAQ-format mapping are DESIGNED in docs/design/data-model.md but have NO code and NO tables. Treat data-model.md as forward design, not current schema.
- Durability is layered and clients are off the write path: every byte hits the raw log before parse; the single SQLite writer batch-commits (default 250ms); only post-commit readings are broadcast. Slow/overloaded clients are DROPPED, never allowed to block ingestion.
- internal/parser is the ONLY protocol-specific code and is deliberately permissive (skips blank/comment/garbage, tolerates short/bad fields). DefaultConfig is the synthetic 4-channel layout (pressure,rate,density,volume); it is UNVERIFIED against the real 15-column Enbridge DAQ CSVs. Per project axiom, real-format adaptation is meant to be a no-code mapping/compute layer, not parser edits.
- No auth, no CI, no Docker. Deployment = one CGO-free arm64 binary + systemd on an offline Pi; data dir is one flag (`-data-dir`/`$CEMENTER_DATA_DIR`/`./data`), intended on an SSD in prod.
- The web client renders DYNAMIC channel cards (one per channel id seen in the stream) — no fixed channel set in the UI. web/src/chart/ and internal/api/ are empty placeholder dirs for unbuilt phases.
- esp32sketches/ and "pi4b & test db/" are a collaborator's Python→ESP32→InfluxDB→Grafana test bench, NOT the shipped Go product (deep-dive recommends keeping it only as a dev/diagnostic bench).

## Tags
#cementer #map #primary #go #vanilla-ts #single-binary #raspberry-pi #daq #sqlite #websocket

## Links
- [structure.map.md](./structure.map.md)
- [dependencies.map.md](./dependencies.map.md)
- [schema.map.md](./schema.map.md)
- [config.map.md](./config.map.md)
- [build.map.md](./build.map.md)
- [error.map.md](./error.map.md)
- [test.map.md](./test.map.md)
- [api.map.md](./api.map.md)
- [state.map.md](./state.map.md)
- [events.map.md](./events.map.md)
- [infra.map.md](./infra.map.md)
- [style.map.md](./style.map.md)
- [non-compliance.report.md](./non-compliance.report.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
