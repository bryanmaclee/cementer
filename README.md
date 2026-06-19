# cementer

Data-acquisition module for an oilfield, high-pressure, down-hole cement pump.

Data arrives as RS-232 → USB serial on a Raspberry Pi. A single Go binary captures
every line durably, parses it, stores it in SQLite, and fans the readings out to
browser clients over WebSocket. The browser shows a dark-mode live readout (Phase 1),
and — in later phases — the customizable, printable multi-line job chart that is the
centerpiece of the project.

## Architecture (one binary)

```
pump ──RS232──►[USB adapter]──► Pi: cementer
                                   source ─► rawlog (append-only, durability layer 1)
                                          ─► daqformat engine ─► store (SQLite WAL, layer 2)
                                                        └─ after commit ─► hub ─► WebSocket clients
                                   + serves the embedded dark-mode web client
```

**Reliability rule:** ingestion is decoupled from clients. Every byte is appended to a
raw log first, the single SQLite writer batch-commits, and only committed readings are
broadcast. A slow or crashed client is dropped, never blocking ingestion — so nothing is
lost on a multi-hour job. See `docs`/the plan for detail.

## Layout

| Path | Role |
|---|---|
| `cmd/cementer/` | entrypoint: wiring, WebSocket, embedded SPA, flags |
| `internal/source/` | `LineSource` interface + replay (dev) source |
| `internal/serialreader/` | production serial source (`go.bug.st/serial`) |
| `internal/rawlog/` | append-only raw capture (durability layer 1) |
| `internal/daqformat/` | generic, **config-driven** mapping + compute engine — a new pump format is a `DaqFormat` value (data), not code. Ships the Intellisense + synthetic presets |
| `internal/parser/` | Phase-1 positional ASCII→`Reading` parser (superseded by `daqformat`; off the main path, kept for its tests) |
| `internal/store/` | SQLite (modernc, pure-Go) single-writer (durability layer 2) |
| `internal/hub/` | WebSocket fan-out (drops slow clients) |
| `web/` | vanilla TS + Vite client (dark mode); built into `web/dist`, embedded |
| `deploy/cementer.service` | systemd unit for the Pi |
| `testdata/sample-stream.txt` | synthetic stream for development without a pump |

## Build & run

Requires Go 1.22+ and Node. The web client is built first and embedded into the binary.

```sh
make build                              # builds web/dist then the cementer binary
make run                                # runs against the synthetic replay stream
# then open http://localhost:8080
```

Run manually:

```sh
# development (no pump): replay the synthetic 4-channel file
./cementer -source testdata/sample-stream.txt -format synthetic -replay-interval 250ms

# replay a real Intellisense wire capture (14-col, 19200 8N1)
./cementer -source captures/capture-2026-06-16T161347-19200-8N1-pressure.bin -format intellisense

# production (real pump on the Pi): a STABLE serial path, data on an SSD
./cementer -serial /dev/serial/by-id/XXXX -baud 19200 -format intellisense -data-dir /mnt/ssd/cementer-data -addr :80
```

Key flags: `-serial` / `-source` (one required), `-baud`, `-format` (`intellisense` (default)
| `synthetic`), `-replay-interval`, `-replay-loop`, `-data-dir`, `-batch-interval`, `-addr`.

**Storage location is trivially flippable** between the Pi's built-in storage (dev)
and an SSD (prod): `-data-dir`, else `$CEMENTER_DATA_DIR`, else `./data`. One value.

## Design

The pump-specific, configuration-driven model (pump profiles, no-code DAQ formats,
per-unit vs aggregate channels, the two chart-config scopes) is described in
[`docs/design/data-model.md`](docs/design/data-model.md). The privileged user is **the
cementer** (crew foreman). Each Pi is a standalone island; the pump self-describes.

## Deploy to a Raspberry Pi (single binary, no C toolchain)

```sh
make pi                                 # cross-compiles cementer-arm64 (CGO disabled)
# copy cementer-arm64 + deploy/cementer.service to the Pi, edit the unit, then:
#   sudo systemctl enable --now cementer
```

## Status

- **Phase 1** complete: durable ingest → WebSocket → dark-mode live value readout, with a
  replay source so the whole pipeline runs without the pump.
- **Phase 2** complete: the config-driven `internal/daqformat` engine + the Intellisense
  preset (characterized from a live wire capture) + the `-format` flag.
- **Phase 3a** complete: the **self-describing pump backbone** — `pump_profiles` /
  `profile_channels` tables (seeded on first run from the active format's vocab), the
  per-connection **hello/profile** WS message, the `GET/PUT /api/profile` +
  `POST /api/profile/reset` HTTP API (`internal/api`, store is the sole DB owner), and a
  **scope-grouped** live readout that renders only the channels this rig actually has.
- **Phase 3b** complete: **jobs + recording segments** — `jobs` / `recording_segments`
  tables, the `/api/jobs*` + `/api/recording/*` HTTP API (job CRUD, an active-job concept,
  record start/stop/adjust), and a minimal client control strip (active-job selector,
  Record/Stop with elapsed timer, inline new-job form). Recording is a **pure marker over
  the always-on store** — start/stop/adjust insert/update segment rows only; they never
  gate ingestion or the live readout, and never reset stage volume (axioms #1 & #5).

Next: the uPlot charting centerpiece (Phase 4 — the printable per-job chart that defaults
to recorded segments). The phased plan lives in
[`docs/changes/phase3-jobs-recording-profiles/scope.md`](docs/changes/phase3-jobs-recording-profiles/scope.md).
