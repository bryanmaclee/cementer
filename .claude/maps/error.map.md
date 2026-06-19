# error.map.md
# project: cementer
# updated: 2026-06-19T23:05:55Z  commit: 1465bd9

## Custom Error Types (store sentinel errors)
| Error                 | File                          | Line | When raised                                                |
|-----------------------|-------------------------------|------|------------------------------------------------------------|
| ErrJobNameRequired    | internal/store/jobs.go:41     | ~41  | CreateJob or UpdateJob called with empty Name              |
| ErrNoSuchJob          | internal/store/jobs.go:44     | ~44  | GetJob / UpdateJob / SetActiveJob for unknown id           |
| ErrNoActiveJob        | internal/store/recording.go:35| ~35  | StartRecording when no job has is_active=1                 |
| ErrRecording          | internal/store/recording.go:39| ~39  | StartRecording when already open; SetActiveJob mid-record  |
| ErrNotRecording       | internal/store/recording.go:43| ~43  | StopRecording when no segment is open                      |
| ErrNoSuchSegment      | internal/store/recording.go:46| ~46  | AdjustSegment for unknown segment id                       |
| ErrBadSegmentRange    | internal/store/recording.go:49| ~49  | AdjustSegment where started_at > stopped_at                |

All are `errors.New(...)` sentinel values; handlers use `errors.Is(err, store.ErrXxx)` to map them to HTTP status codes.

## HTTP status code mapping  [internal/api/]
| Error sentinel        | HTTP status |
|-----------------------|-------------|
| ErrJobNameRequired    | 400         |
| ErrNoSuchJob          | 404         |
| ErrNoSuchSegment      | 404         |
| ErrBadSegmentRange    | 400         |
| ErrNoActiveJob        | 400         |
| ErrRecording          | 409         |
| ErrNotRecording       | 409         |
| anything else         | 500         |

## Error Handling Patterns
- **Startup fatal:** `main()` → `run()` returns error → `log.Fatalf("cementer: %v")` — any wiring error exits the process.
- **Per-line tolerance:** `daqformat.Engine.Apply()` returns `(Reading, ok=false)` for blank/comment/torn/header lines; bad individual fields are skipped, not fatal.
- **Best-effort logging (keep running):** rawlog append failure → `log.Printf`, continue; store commit failure → `fmt.Printf` + keep ingesting (raw log already has the bytes).
- **API handlers:** decode errors → 400; store sentinel errors → mapped status; unexpected errors → 500 with logged detail.
- **WebSocket pumps:** any read/write error exits the pump goroutine; `readPump`'s deferred `Unregister` cleans up. Upgrade failure returns silently.

## Global Error Boundaries
| Boundary              | File                        | Scope                                      |
|-----------------------|-----------------------------|--------------------------------------------|
| HTTP 500 on Stats()   | cmd/cementer/main.go        | GET /debug/stats only                      |
| API handler 500       | internal/api/ all handlers  | each handler writes its own error response |
| WS malformed frame    | web/src/ws.ts               | try/catch; bad frame is silently dropped   |
| Ordered shutdown      | cmd/cementer/main.go        | 5s context; server → source → store → rawlog |

## Resilience / Drop Policy (by design)
- **`hub.Broadcast`:** non-blocking; if hub overloaded the live message is DROPPED (data already durable in SQLite).
- **`hub.Run` subscriber send:** a subscriber whose buffered Send is full is DROPPED and closed — ingestion is never blocked by a slow client.
- **`store.Submit`:** blocks under backpressure (never drops while open); silently discards only after `Close` (shutdown path).

## Unhandled Error Risks
- Several `Close()` / `Sync()` calls on the shutdown path are deliberately ignored (`_ =`) — acceptable for a process about to exit, but a sync error before a crash would be silent.
- Store commit failure leaves the batch uncommitted to SQLite (recoverable from raw log by manual re-import; no automatic retry).

## Tags
#cementer #map #error #go-errors #durability #drop-policy #graceful-shutdown #sentinel-errors

## Links
- [primary.map.md](./primary.map.md)
- [api.map.md](./api.map.md)
- [state.map.md](./state.map.md)
- [events.map.md](./events.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
