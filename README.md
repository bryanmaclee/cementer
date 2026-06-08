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
                                          ─► parser ─► store (SQLite WAL, layer 2)
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
| `internal/parser/` | ASCII line → `Reading` — **the only protocol-specific code** |
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
# development (no pump): replay a captured/synthetic file
./cementer -source testdata/sample-stream.txt -replay-interval 250ms

# production (real pump on the Pi): a STABLE serial path, data on an SSD
./cementer -serial /dev/serial/by-id/XXXX -baud 9600 -data-dir /mnt/ssd/cementer-data -addr :80
```

Key flags: `-serial` / `-source` (one required), `-baud`, `-replay-interval`,
`-replay-loop`, `-data-dir`, `-batch-interval`, `-addr`.

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

Phase 1 complete: durable ingest → WebSocket → dark-mode live value readout, with a
replay source so the whole pipeline runs without the pump. Next: job CRUD + auth, then
the uPlot charting centerpiece. See the build plan for the phased roadmap.
