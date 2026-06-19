# events.map.md
# project: cementer
# updated: 2026-06-19T23:05:55Z  commit: 1465bd9

The only "event bus" is the in-process hub that fans committed readings to WebSocket clients. No Kafka, RabbitMQ, Redis pub-sub, or EventEmitter. Client polling of the REST API (recording state, active job, jobs list) is not event-driven — it is interval-based pull.

## Bus Type
In-process channel-based fan-out hub.  [internal/hub/hub.go]
Hub channels: `register` (unbuffered), `unregister` (buf 16), `broadcast` (buf 256). One `Run` goroutine owns the subscriber map.
Subscriber: a buffered `Send chan []byte` (buffer 256 as created in serveWS).

## WebSocket Message Topics
| Type      | Payload                           | When sent                                       |
|-----------|-----------------------------------|-------------------------------------------------|
| "profile" | `{ type, profile: Profile }`      | Once per WS connection open (hello frame), written directly to the conn BEFORE the hub loop |
| "reading" | `{ type, reading: Reading }`      | After each SQLite batch commit; via hub.Broadcast → all Subscriber.Send channels |

(The `wsEnvelope` type field reserves room for future kinds — none currently besides these two.)

## Emitters
| Emitter                            | Topic     | Notes                                                      |
|------------------------------------|-----------|------------------------------------------------------------|
| serveWS → conn.WriteMessage        | "profile" | Direct write to this connection only; NOT routed through hub |
| store.writeLoop → onCommit closure | "reading" | fires once per reading post-commit; marshals + calls hub.Broadcast |
| hub.Broadcast                      | —         | non-blocking enqueue to broadcast chan; drops if hub overloaded |

## Listeners
| Listener                       | Listens for | Notes                                                           |
|--------------------------------|-------------|-----------------------------------------------------------------|
| hub.Run → Subscriber.Send      | broadcast   | non-blocking send per subscriber; DROPS + closes slow subscriber |
| WS writePump                   | Subscriber.Send | drains to conn.WriteMessage; pings every 54s              |
| Client ws.onmessage            | both types  | parses env; routes to onProfile or onReading callback          |

## REST poll events (client-side, not hub)
`Controls` polls `/api/recording/state` every 3s and `/api/jobs` + `/api/job/active` on init and after actions. No server push for job/recording state changes (polling only).

## Lifecycle / Cleanup
- Register: `serveWS` creates `NewSubscriber(256)`, calls `h.Register(sub)`.
- Unregister: `readPump`'s deferred `h.Unregister(sub)` on any read error/disconnect → `Run` closes `Send` → `writePump` exits.
- Shutdown: ctx cancel → `hub.Run` closes every `Subscriber.Send` and returns.

## Reliability Rule
Ingestion is NEVER blocked by a client. Both `hub.Broadcast` (hub overload) and per-subscriber send (slow client) DROP rather than back up. Dropped clients reconnect (auto-reconnect with exponential backoff 1s–10s) and will miss the dropped readings — those are already in SQLite; a future backfill mechanism could serve them.

## Tags
#cementer #map #events #websocket #fan-out #pub-sub #drop-policy #hub #profile-frame

## Links
- [primary.map.md](./primary.map.md)
- [api.map.md](./api.map.md)
- [state.map.md](./state.map.md)
- [error.map.md](./error.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
