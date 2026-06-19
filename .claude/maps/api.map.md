# api.map.md
# project: cementer
# updated: 2026-06-19T23:05:55Z  commit: 1465bd9

All REST routes are registered by `api.API.Register(mux)` [internal/api/api.go:34]; WS and debug routes are registered directly in `cmd/cementer/main.go`. Go 1.22+ method-pattern ServeMux. No router library.

## WebSocket Endpoint
| Route       | Handler    | Notes                                                                              |
|-------------|------------|------------------------------------------------------------------------------------|
| GET /ws/live| serveWS()  | Upgrades to WebSocket; sends hello/profile frame (enabled channels only) on connect BEFORE entering live read loop; registers hub.Subscriber(256). Ping every 54s, pongWait 60s, writeWait 10s, ReadLimit 4096. Auth: none (LAN). |

WebSocket frames are JSON `wsEnvelope { type, reading?, profile? }`:
- `type: "profile"` — sent once per connection (and after reconnect): `{ type: "profile", profile: Profile }`.
- `type: "reading"` — each committed batch reading: `{ type: "reading", reading: Reading }`.

## REST Endpoints — Profile  [internal/api/api.go]
| Method | Path                | Handler         | Notes                                                                     |
|--------|---------------------|-----------------|---------------------------------------------------------------------------|
| GET    | /api/profile        | getProfile      | Returns ALL channels (enabled + disabled) as EditorProfile. 404 if no active profile. |
| PUT    | /api/profile        | putProfile      | Body: `{ units?: int, channels: [{id, enabled?, label?, uom?, decimals?, sortOrder?}] }`. Patches only sent fields. 400 on unknown channel id or out-of-range decimals. Returns refreshed EditorProfile. |
| POST   | /api/profile/reset  | resetProfile    | Replaces active profile's channels with the active format's default vocab (all enabled, sort_order = preset order). Returns refreshed EditorProfile. |

## REST Endpoints — Jobs  [internal/api/jobs.go]
| Method | Path                | Handler         | Notes                                                                     |
|--------|---------------------|-----------------|---------------------------------------------------------------------------|
| GET    | /api/jobs           | listJobs        | Returns all jobs, newest first. Always 200 (empty array if none).         |
| POST   | /api/jobs           | createJob       | Body: `jobDTO` (name required; rest optional). 201 + created job. 400 if name empty. |
| GET    | /api/jobs/{id}      | getJob          | 200 + job or 404.                                                         |
| PUT    | /api/jobs/{id}      | updateJob       | Body: `jobDTO`. 200 + updated job. 404 / 400.                             |
| GET    | /api/job/active     | getActiveJob    | 200 + active job or `{"active": null}` when none.                         |
| PUT    | /api/job/active     | setActiveJob    | Body: `{ id: int64 }`. Makes job active. 409 if a DIFFERENT job is currently recording (stop recording first). |

## REST Endpoints — Recording  [internal/api/jobs.go]
| Method | Path                          | Handler          | Notes                                                              |
|--------|-------------------------------|------------------|--------------------------------------------------------------------|
| GET    | /api/recording/state          | recordingState   | Returns `{ recording, openSegmentId?, jobId? }`.                   |
| POST   | /api/recording/start          | startRecording   | Opens segment under active job. 201 + Segment. 400 if no active job. 409 (+ open segment body) if already recording. |
| POST   | /api/recording/stop           | stopRecording    | Closes open segment. 200 + Segment. 409 if not recording.          |
| GET    | /api/recording/segments       | listSegments     | Query: `?job_id=N` (required). Returns all segments for that job, oldest first. |
| PUT    | /api/recording/segments/{id}  | adjustSegment    | Body: `{ startedAtUs?: int64, stoppedAtUs?: int64 }`. Nudges endpoints. 404 / 400 (bad range). Returns refreshed Segment. |

## REST Endpoints — Series  [internal/api/series.go]
| Method | Path                      | Handler      | Notes                                                                                  |
|--------|---------------------------|--------------|----------------------------------------------------------------------------------------|
| GET    | /api/samples              | getSamples   | Query: `from=<us>&to=<us>[&channels=a,b,c][&max=N]`. from/to in unix-microseconds. Returns `{ series: { channelId: [[ts_us, value], ...] } }`. Default max=4000/channel, cap=20000. |
| GET    | /api/jobs/{id}/series     | getJobSeries | Query: `[channels=...][&max=N]`. Returns `{ segments: Segment[], series: {...} }`. 404 if job not found. Only in-segment samples returned (gaps between segments stay gaps). |

## Debug Endpoint  [cmd/cementer/main.go]
| Method | Path            | Notes                                            |
|--------|-----------------|--------------------------------------------------|
| GET    | /debug/stats    | JSON `{ rows: int64, latest_ts: string }`. 500 on error. |

## SPA Fallback  [cmd/cementer/main.go]
| Method | Path  | Notes                                                    |
|--------|-------|----------------------------------------------------------|
| GET    | /     | Serves embedded `web/dist` via `http.FileServerFS`. Unknown paths fall back to `index.html`. |

## Auth
None. All endpoints are unauthenticated (LAN deployment; CheckOrigin returns true for WS).

## Request discipline
All POST/PUT handlers call `json.NewDecoder.DisallowUnknownFields()` — unknown fields are rejected with 400.

## Tags
#cementer #map #api #websocket #http #servemux #profile #jobs #recording #series #rest

## Links
- [primary.map.md](./primary.map.md)
- [schema.map.md](./schema.map.md)
- [events.map.md](./events.map.md)
- [state.map.md](./state.map.md)
- [error.map.md](./error.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
