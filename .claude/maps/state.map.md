# state.map.md
# project: cementer
# updated: 2026-06-19T23:05:55Z  commit: 1465bd9

This is a data-pipeline appliance, not a UI-state app. "State" = the server-side ingest→store→broadcast pipeline (durable source of truth) plus the thin client-side display/control state. No Redux/Zustand/etc.

## Server Pipeline (canonical data flow)  [cmd/cementer/main.go run()]
```
source (serialreader | Replay file)
  --emit(line)--> handleLine:
    1. rawlog.Append(line)               — durability layer 1 (append-only file, periodic fsync)
    2. daqformat.Engine.Apply(line, now) — line → Reading (or skip: blank/torn/header/zero-values)
    3. store.Submit(reading)             — queue for the single SQLite writer

store.writeLoop (single goroutine):
    4. batch up to 512 readings OR every 250ms (batchInterval)
    5. BEGIN; INSERT each channel value into `samples`; COMMIT  ← THE DURABILITY POINT
    6. onCommit(reading) per committed reading → json.Marshal(wsEnvelope) → hub.Broadcast

hub.Run (single goroutine):
    7. fan-out to all Subscriber.Send channels; DROP+CLOSE any whose buffer is full

WS writePump: drains Subscriber.Send → conn.WriteMessage(TextMessage)
Client: readout.update(reading) → liveChart.push(reading)
```

**Invariant:** clients only ever see readings that are already durably committed. Broadcast is post-commit.

## Subsystem State Ownership
| Subsystem             | Owner goroutine           | Mutable state                                       |
|-----------------------|---------------------------|-----------------------------------------------------|
| store.Store           | single writeLoop goroutine| `in chan model.Reading` (buf 4096), *sql.DB, batch  |
| hub.Hub               | single Run goroutine      | `subs map[*Subscriber]struct{}`                     |
| daqformat.Engine      | ingest goroutine (single) | seq counter, headerSkipped bool                     |
| rawlog.Writer         | multiple (mutex)          | file + bufio.Writer behind sync.Mutex               |
| serialreader / Replay | dedicated goroutine       | serial.Port / io.ReadCloser + bufio.Scanner         |

Concurrency model: each subsystem confines mutable state to one goroutine; communication is over channels (CSP). No shared locks across subsystems except rawlog's internal mutex.

## Server-side persistent state (SQLite)
Five tables: `samples`, `pump_profiles`, `profile_channels`, `jobs`, `recording_segments`. See schema.map.md for full DDL.
- Exactly one pump_profiles row has `is_active=1`.
- Exactly one jobs row has `is_active=1` (or zero if no jobs created yet).
- Exactly one recording_segments row has `stopped_at_us IS NULL` (the open segment) — or zero if not recording.

## Client Display State  [web/src/readout.ts — class Readout]
| State field     | Type               | Notes                                               |
|-----------------|--------------------|-----------------------------------------------------|
| liveChart       | LiveChart          | rolling ring buffer (xs: number[], ys: number[][]), in-memory |
| jobChart        | JobChart           | fetches from /api/jobs/{id}/series on load          |
| connected       | boolean            | WS connection status                                |
| lastReadingAt   | number             | Date.now() of last reading; drives live/stalled/offline (STALE_MS=3000) |
| activeJobId     | number\|null       | set by Controls, drives Job History view auto-load  |
| view            | "live"\|"job"      | which tab is shown                                  |

## Client Control State  [web/src/controls.ts — class Controls]
| State field     | Type               | Notes                                               |
|-----------------|--------------------|-----------------------------------------------------|
| jobs            | Job[]              | refreshed from /api/jobs                            |
| activeJobId     | number\|null       | tracks the active job; notified via onActiveJob callback |
| recording       | boolean            | from /api/recording/state poll (every 3s)           |
| openStartedAtUs | number\|null       | started_at of the open segment (for elapsed timer)  |
| clockSkewUs     | number             | server unix-micros minus client Date.now()*1000     |

## Personal Config State  [web/src/chart/config.ts — localStorage]
Key `cementer.liveview`: `{ hidden?: Record<string, boolean>, colors?: Record<string, string>, windowSec?: number }`. Read on init; written on toggle/change. Pi-scoped state (profile/jobs/recordings) is NEVER in localStorage (axiom #3).

## Access Pattern
Server: pure channel hand-off; no global Go state. Client: `connectLive(onReading, onStatus, onProfile)` callbacks drive `readout.update()`/`readout.setConnected()`/`readout.applyProfile()`; `Controls` polls the REST API and calls `onActiveJob` callback on change.

## Tags
#cementer #map #state #pipeline #csp #single-writer #durability #dynamic-channels #ring-buffer #recording

## Links
- [primary.map.md](./primary.map.md)
- [schema.map.md](./schema.map.md)
- [events.map.md](./events.map.md)
- [error.map.md](./error.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
