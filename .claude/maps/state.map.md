# state.map.md
# project: cementer
# updated: 2026-06-12T09:02:13-06:00  commit: ee446c3

This is a data-pipeline, not a UI-state app. "State" here = the server-side ingest→store→broadcast pipeline (the source of truth) plus the thin client-side display state. There is no Redux/Zustand-style store.

## Pipeline (the canonical data flow)  [cmd/cementer/main.go run()]
source (serial | replay) --emit(line)--> handleLine:
  1. rawlog.Append(line)            durability layer 1 (append-only file, periodic fsync)  [internal/rawlog]
  2. parser.Parse(line, now)        line → Reading (or skip)                                [internal/parser]
  3. store.Submit(reading)          queue for the single SQLite writer                      [internal/store]
store writeLoop (single goroutine):
  4. batch readings (maxBatch 512 or every batchInterval, default 250ms)
  5. commit batch in one tx → INSERT each channel value into `samples`  (THE durability point)
  6. onCommit(reading) per committed reading → json.Marshal(wsEnvelope) → hub.Broadcast
hub.Run (single goroutine):
  7. fan-out broadcast to all Subscriber.Send channels; drop+close any whose buffer is full
WS writePump: drains Subscriber.Send → socket; client renders.

Invariant: clients only ever see readings that are already durably committed (broadcast happens in onCommit, post-commit).

## Server "store shape" (ownership of mutable state)
store.Store          — owns *sql.DB, the `in chan model.Reading` (buf 4096), batchInterval, onCommit, single writer goroutine. [internal/store/store.go]
hub.Hub              — owns subs map[*Subscriber]struct{}, register/unregister/broadcast channels; single Run goroutine mutates the map. [internal/hub/hub.go]
parser.Parser        — owns cfg + monotonic seq counter; safe for one ingest goroutine. [internal/parser/parser.go]
rawlog.Writer        — owns file + buffered writer behind a mutex + background fsync loop. [internal/rawlog/rawlog.go]

Concurrency model: each subsystem confines its mutable state to one goroutine and communicates over channels (CSP). No shared locks across subsystems except rawlog's internal mutex.

## Client display state  [web/src/readout.ts — class Readout]
cards: Map<channelId, Card>  — one DOM card per channel id seen in the stream (DYNAMIC; created on first sighting, reordered by role priority)
connected: boolean           — WS connection status
lastReadingAt: number        — Date.now() of last reading; drives live/stalled/offline status (STALE_MS = 3000)
seq display + "updated Ns ago" footer; status recomputed every 1s.
Theme state: localStorage "cementer.theme" (dark default / light) via web/src/theme.ts — per-laptop, not synced.

## Access Pattern
Server: pure channel hand-off; no global state. Client: connectLive(onReading, onStatus) callbacks drive readout.update() / readout.setConnected(); no state-management library.

## Designed-but-unbuilt state
Jobs, recording segments (start/stop markers over the continuous store), pump profiles, DAQ-format mapping/compute — all DESIGN ONLY (data-model.md). The store currently keeps a continuous, ungated `samples` table with no job/segment scoping.

## Tags
#cementer #map #state #pipeline #csp #single-writer #durability #dynamic-channels

## Links
- [primary.map.md](./primary.map.md)
- [events.map.md](./events.map.md)
- [schema.map.md](./schema.map.md)
- [error.map.md](./error.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
