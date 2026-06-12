# schema.map.md
# project: cementer
# updated: 2026-06-12T09:02:13-06:00  commit: ee446c3

## Go Types (data contracts)
### Reading  [internal/model/model.go:10]
The unit parsed from one ASCII line and broadcast live to clients.
seq: int64        — monotonic per-process sequence; gap detection; NOT persisted as a key
ts: time.Time     — ingest timestamp (server clock today)
values: map[string]float64 — channel name → numeric value for this frame
Method: Samples() []Sample — expands one Reading into one Sample per channel.

### Sample  [internal/model/model.go:31]
Long-form storage unit: one channel's value at one instant.
ts: time.Time
channel: string
value: float64

### parser.Config  [internal/parser/parser.go:18]
delimiter: string   — field separator (default ",")
channels: []string  — field index → channel name; "" means ignore that column
DefaultConfig(): {delimiter: ",", channels: ["pressure","rate","density","volume"]}

### serialreader.Config  [internal/serialreader/serialreader.go:16]
port: string, baudRate: int, dataBits: int, parity: serial.Parity, stopBits: serial.StopBits
DefaultConfig(port): 9600 8N1

### store.Stats  [internal/store/store.go:167]
rows: int64        — COUNT(*) of samples
latest_ts: time.Time — time of MAX(ts_us)

### wsEnvelope  [cmd/cementer/main.go:192]  (unexported; the WS wire shape)
type: string                 — message kind; currently only "reading"
reading: *model.Reading      — omitempty; the payload

## TypeScript Types  (web/src/types.ts — mirror of the Go contracts)
### Reading  [web/src/types.ts:3]
seq: number
ts: string   — RFC3339 timestamp
values: Record<string, number>

### WSEnvelope  [web/src/types.ts:9]
type: string
reading?: Reading

### ChannelSpec  [web/src/readout.ts:20]  (client display metadata, inferred from channel id)
label: string, uom: string, decimals: number, order: number

## Database Models (SQLite)  [internal/store/store.go:69 initSchema]
### samples  — the ONLY table that exists in code
id      INTEGER PRIMARY KEY
ts_us   INTEGER NOT NULL   — unix microseconds (r.TS.UnixMicro())
channel TEXT    NOT NULL   — channel id (e.g. "pressure")
value   REAL    NOT NULL
Index: idx_samples_ts ON samples(ts_us)
DSN pragmas: journal_mode=WAL, synchronous=NORMAL, busy_timeout=5000, foreign_keys=ON. MaxOpenConns=1 (single serialized writer).

## Designed-but-NOT-implemented models (docs/design/data-model.md — NO code, NO tables)
The following are DESIGN ONLY — they do not exist in internal/store or anywhere in source. Do not treat as current schema.
- PumpProfile { id, name, units, channels[] } and Channel { id, role, scope, unitIndex?, label, uom, decimals, source }
- DaqFormat { id, name, delimiter, hasHeader, timestamp?, fields[] }, FieldMap { column, channelId, transform? }, ComputedChannel { channelId, op, inputs }
- RecordingSegment { job_id, id, started_at, stopped_at? }
- Job concept (no `jobs` table)
See non-compliance.report.md "Design-ahead-of-code" for the full delta.

## Tags
#cementer #map #schema #sqlite #go-types #typescript #wire-contract

## Links
- [primary.map.md](./primary.map.md)
- [state.map.md](./state.map.md)
- [api.map.md](./api.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
