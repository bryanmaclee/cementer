# events.map.md
# project: cementer
# updated: 2026-06-12T09:02:13-06:00  commit: ee446c3

The only "event bus" is the in-process hub that fans committed readings out to WebSocket clients. There is no Kafka/RabbitMQ/Redis/EventEmitter. Transport to clients is WebSocket (gorilla/websocket).

## Bus Type
In-process channel-based fan-out hub.  [internal/hub/hub.go]
Hub channels: register (unbuffered), unregister (buf 16), broadcast (buf 256). One Run goroutine owns the subscriber set.
Subscriber: a buffered Send chan []byte (buffer 256 as created in main.go serveWS).

## Message Topics (WebSocket frames)
"reading" — the single message kind today. Payload: wsEnvelope { type: "reading", reading: Reading }. JSON text frame.
  (wsEnvelope.type reserves room for future kinds — hello/profile, job updates, log entries — none implemented.)

## Emitters
store.writeLoop → onCommit(reading)   [internal/store/store.go:108] — fires once per reading AFTER its batch commits.
onCommit closure (main.go:83) → json.Marshal(wsEnvelope) → hub.Broadcast(bytes)  [cmd/cementer/main.go:83-88]
hub.Broadcast → broadcast channel; non-blocking (drops the message if the hub is overloaded, since data is already durable). [hub.go:91]

## Listeners
hub.Run → for each Subscriber, non-blocking send to Subscriber.Send; if full, DROP + close (slow-client policy). [hub.go:62]
writePump (main.go:225) drains Subscriber.Send → conn.WriteMessage(TextMessage); plus periodic Ping.
Client: ws.onmessage → JSON.parse → if type==="reading" → onReading(reading) → Readout.update. [web/src/ws.ts:28]

## Lifecycle / cleanup
Register: serveWS creates NewSubscriber(256), h.Register(sub). [main.go:216]
Unregister: readPump's deferred h.Unregister(sub) on any read error/disconnect → Run closes Send → writePump exits. [main.go:255]
Shutdown: ctx cancel → hub.Run closes every Subscriber.Send and returns. [hub.go:49]

## Reliability rule
Ingestion is never blocked by a client. Both Broadcast (hub overload) and per-subscriber send (slow client) DROP rather than back up. Dropped clients reconnect and backfill from the durable store (backfill mechanism itself: designed, not yet built — currently a reconnected client just resumes live).

## Tags
#cementer #map #events #websocket #fan-out #pub-sub #drop-policy #hub

## Links
- [primary.map.md](./primary.map.md)
- [api.map.md](./api.map.md)
- [state.map.md](./state.map.md)
- [error.map.md](./error.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
