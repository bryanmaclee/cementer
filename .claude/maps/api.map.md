# api.map.md
# project: cementer
# updated: 2026-06-12T09:02:13-06:00  commit: ee446c3

All routes are registered on a single net/http ServeMux in cmd/cementer/main.go (no router library, no internal/api package). Method-prefixed patterns (Go 1.22+ ServeMux).

## HTTP Endpoints
GET /ws/live   [cmd/cementer/main.go:137 → serveWS]
  Handler: serveWS(h) — upgrades to WebSocket (gorilla/websocket), registers a hub.Subscriber(256), starts writePump + readPump.
  Auth: none (CheckOrigin returns true — LAN deployment; tighten when network posture changes)
  Direction: server→client only; clients send nothing meaningful (readPump just detects disconnect).
  Message shape (text frame): wsEnvelope { type: "reading", reading: Reading }
  Ping/pong: server pings every 54s (pingPeriod), pongWait 60s, writeWait 10s, ReadLimit 4096.

GET /debug/stats   [main.go:138 → serveStats]
  Handler: serveStats(st) — JSON snapshot from store.Stats()
  Auth: none
  Response: 200 application/json { rows: int64, latest_ts: RFC3339 } | 500 on error

GET /   (and all unmatched paths)   [main.go:139 → mountSPA / spaFallback]
  Handler: serves embedded web/dist via http.FileServerFS; unknown paths fall back to index.html (SPA routing, forward-looking).
  Auth: none
  Response: static assets (index.html, hashed JS/CSS)

## Parser contract (ASCII line → Reading)  [internal/parser/parser.go]
The wire-protocol boundary. parser.Parse(line []byte, ts time.Time) (model.Reading, ok bool):
  - Trim line; skip if empty or starts with "#" (comment) → ok=false
  - Split on cfg.Delimiter (default ","); for field index i map to cfg.Channels[i] (skip name=="" or i out of range)
  - ParseFloat each field; unparseable field is omitted (not fatal)
  - If zero usable values → ok=false; else increment seq and return Reading{seq, ts, values}
  DefaultConfig channels (current, dev): ["pressure","rate","density","volume"], delimiter ",".
  Note: ts is the server clock today (model.Reading.TS comment); a wire timestamp column is not yet used.

## Client WS contract  [web/src/ws.ts, web/src/types.ts]
connectLive(onReading, onStatus): opens ws(s)://<host>/ws/live, capped exponential backoff reconnect (1s→10s).
  Parses each frame as WSEnvelope; dispatches onReading only when env.type === "reading" && env.reading.
  Dev: Vite proxies /ws → ws://localhost:8080 so the client runs on Vite's dev server.

## Not present
No REST CRUD, no GraphQL, no gRPC. Job / pump-profile / recording endpoints are DESIGNED (data-model.md) but unbuilt. The "type" field in wsEnvelope reserves room for future message kinds (e.g. hello/profile, job updates).

## Tags
#cementer #map #api #websocket #http #parser-contract #servemux

## Links
- [primary.map.md](./primary.map.md)
- [events.map.md](./events.map.md)
- [schema.map.md](./schema.map.md)
- [state.map.md](./state.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
