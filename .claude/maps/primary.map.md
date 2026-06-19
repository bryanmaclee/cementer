# primary.map.md
# project: cementer
# updated: 2026-06-19T23:05:55Z  commit: 1465bd9

## Project Fingerprint
Language:   Go 1.26.4 (module github.com/bryanmaclee/cementer) + TypeScript (vanilla, ES2022) client
Framework:  none — `net/http` stdlib + `gorilla/websocket`; vanilla-TS + Vite + uPlot client (no UI framework)
Runtime:    single static CGO-free binary; targets Raspberry Pi 4B (arm64), fully offline
Type:       embedded full-stack data-acquisition appliance (USB-serial → durable SQLite store → live WS readout + REST API + historical chart)
Size:       ~27 Go source files, ~12 TypeScript files; ~80+ test functions

## Map Index
| Map                  | Status  | Contents                                                           |
|----------------------|---------|--------------------------------------------------------------------|
| structure.map.md     | present | directory layout, entry points, package ownership                  |
| dependencies.map.md  | present | 3 direct Go deps, 1 web runtime (uPlot), 2 web build; module graphs |
| schema.map.md        | present | 5 SQLite tables, 12+ Go types, 8 TypeScript interfaces             |
| config.map.md        | present | 9 CLI flags, 1 env var, 2 localStorage keys, no .env               |
| build.map.md         | present | Makefile targets, npm scripts, go:embed pipeline, no CI            |
| error.map.md         | present | 7 custom error sentinels, HTTP status mapping, drop policy         |
| test.map.md          | present | go testing, 10 test files, ~79 test functions                      |
| api.map.md           | present | 15 REST routes + WS endpoint + debug                               |
| state.map.md         | present | ingest→store→broadcast pipeline; client display + control state    |
| events.map.md        | present | in-process hub, 2 WS message types (profile, reading)              |
| infra.map.md         | present | single binary + systemd on offline Pi, 5-table SQLite              |
| style.map.md         | present | CSS custom-property theming, uPlot chart colors, class vocabulary  |
| auth.map.md          | absent  | no auth in code (deferred to Phase 5+)                             |
| domain.map.md        | absent  | domain concepts live in docs/design/data-model.md                  |
| i18n.map.md          | absent  | none                                                               |
| migrations.map.md    | absent  | schema is inline `CREATE TABLE IF NOT EXISTS`, no migration tool   |
| jobs.map.md          | absent  | no scheduler/worker/cron (jobs here = cement jobs, not task queues) |

## File Routing
types / interfaces / SQLite schema          → schema.map.md
HTTP REST routes / WS endpoint              → api.map.md
flags / env vars / localStorage keys        → config.map.md
test patterns / fixtures / test counts      → test.map.md
build commands / cross-compile / go:embed   → build.map.md
directory layout / entry points             → structure.map.md
external packages / module graph            → dependencies.map.md
ingest pipeline / store / client state      → state.map.md
hub fan-out / WS message types              → events.map.md
error sentinels / HTTP status mapping       → error.map.md
deploy / systemd / Pi runtime layout        → infra.map.md
theming / CSS tokens / chart colors         → style.map.md

## Key Facts
- **Entry point** `cmd/cementer/main.go` wires: source (serial OR replay) → rawlog (append-only, durability layer 1) → `daqformat.Engine.Apply()` (config-driven field mapping + computed channels) → `store.Submit()` (SQLite WAL single-writer, durability layer 2) → `hub.Broadcast()` → WebSocket. The web client is embedded via `go:embed` (assets.go); ships as one file.
- **DAQ format layer** (`internal/daqformat/`): the `Engine` is generic; a format is a plain `DaqFormat` value. Two presets ship: `Intellisense()` (14-col, 13 channels, live-wire-characterized at 19200 8N1) and `Synthetic()` (4-col dev format). `-format` flag selects the preset at startup. Adding a new pump format requires only a new preset value, not a code edit.
- **Store** (`internal/store/`): five tables — `samples`, `pump_profiles`, `profile_channels`, `jobs`, `recording_segments`. Single writer goroutine; batch commit (default 250ms). Reads (API, series) share the one `*sql.DB` connection serialized by the 1-conn pool + WAL + busy_timeout.
- **Profile seed:** on first run, `main` seeds an active pump profile from the active format's channel vocab (`IntellisenseChannels()` / `SyntheticChannels()`). The store is format-agnostic (never imports `internal/daqformat`); main bridges via `[]store.SeedChannel`.
- **API** (`internal/api/`): 15 REST routes — profile CRUD (GET/PUT/POST reset), jobs CRUD + active-job, recording start/stop/state/segments/adjust, sample series + job series. All handlers call store methods only (never touch `*sql.DB` directly).
- **Client** (`web/src/`): no framework; uPlot for charts. `Readout` shell (Live tab = rolling chart, Job History tab = recorded chart per job). `Controls` strip (job selector + Record/Stop + elapsed timer). `ws.ts` auto-reconnects with exponential backoff. Receives hello/profile frame on connect then live reading frames.
- **Recording is a marker** (axiom #1): recording start/stop inserts/updates rows in `recording_segments` ONLY — it never gates ingestion, the live stream, or stage volume. Samples are stored continuously regardless of record state.
- **Deployment**: static CGO-free arm64 binary + systemd on offline Pi. No Docker, no CI, no cloud. Data dir defaults to `./data`; production points at an external SSD via `-data-dir`.
- **`internal/parser`** is OFF the main path (main.go does NOT import it); it remains for its tests only. The live ingest path uses `internal/daqformat` exclusively.

## Tags
#cementer #map #primary #go #vanilla-ts #single-binary #raspberry-pi #daq #sqlite #websocket #daqformat #pump-profile #jobs #recording #uplot #charting

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
