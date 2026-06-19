# schema.map.md
# project: cementer
# updated: 2026-06-19T23:05:55Z  commit: 1465bd9

## Go Types — model

### Reading  [internal/model/model.go:10]
One parsed frame from a single ASCII line; the unit broadcast live to clients.
| Field  | Type                  | Notes                                         |
|--------|-----------------------|-----------------------------------------------|
| Seq    | int64                 | monotonic per-process counter; not a DB key   |
| TS     | time.Time             | server-stamp (pipeline clock)                 |
| Values | map[string]float64    | channel id → numeric value for this frame     |
Method: `Samples() []Sample` — expands to one Sample per channel.

### Sample  [internal/model/model.go:32]
Long-form storage row: one channel's value at one instant.
| Field   | Type      |
|---------|-----------|
| TS      | time.Time |
| Channel | string    |
| Value   | float64   |

## Go Types — daqformat  [internal/daqformat/daqformat.go]

### DaqFormat
Describes one pump's serial wire format. Pure data; a preset is a DaqFormat literal.
| Field          | Type             | Notes                                                 |
|----------------|------------------|-------------------------------------------------------|
| ID             | string           | stable key, e.g. "intellisense"                      |
| Name           | string           | human label                                           |
| Delimiter      | string           | field separator; empty defaults to ","               |
| HasHeader      | bool             | first non-blank/comment line is a header to skip      |
| ExpectedFields | int              | exact token count for a valid steady-state line       |
| Timestamp      | TimestampSpec    | where/how the timestamp column is interpreted         |
| Fields         | []FieldMap       | raw column → channel mappings                        |
| Computed       | []ComputedChannel| derived channels (sum/mean over other channels)       |

### FieldMap
| Field     | Type       | Notes                                      |
|-----------|------------|--------------------------------------------|
| Column    | int        | zero-based field index in the split line   |
| ChannelID | string     | target channel id, e.g. "unit1.pressure"  |
| Transform | *Transform | optional linear rescale; nil = identity    |

### Transform
Linear rescale: `out = in * Scale + Offset`. A nil `*Transform` is the identity.
| Field  | Type    |
|--------|---------|
| Scale  | float64 |
| Offset | float64 |

### ComputedChannel
Derives a channel by reducing other channels' values (sum or mean).
| Field     | Type       | Notes                           |
|-----------|------------|---------------------------------|
| ChannelID | string     | target id, e.g. "agg.rate"     |
| Op        | string     | "sum" or "mean"                 |
| Inputs    | []string   | channel ids to reduce           |
| Transform | *Transform | optional rescale of result      |

### TimestampSpec
| Field  | Type          | Notes                                                       |
|--------|---------------|-------------------------------------------------------------|
| Column | int           | zero-based timestamp field index                            |
| Kind   | TimestampKind | ServerStamp (0) | HMSUptime (1) | ExcelSerial (2) | Unix (3) | RFC3339 (4) |
Only `ServerStamp` and `HMSUptime` are implemented; the rest fall back to the server clock.

### Channel  (bundled vocab metadata)
| Field     | Type   | Notes                                              |
|-----------|--------|----------------------------------------------------|
| ID        | string | stable key, e.g. "unit1.pressure"                  |
| Role      | string | pressure | rate | density | volume | meta        |
| Scope     | string | unit | aggregate | stage | job | ""           |
| UnitIndex | int    | 1-based when Scope=="unit"; 0 otherwise             |
| UoM       | string | e.g. "psi", "bbl/min", "ppg", "bbl"               |
| Label     | string | display label                                      |
| Decimals  | int    | display precision                                  |

## Go Types — store (wire + persistence)  [internal/store/]

### store.Channel  [profile.go:27]
Enabled-only display metadata; shape sent in the hello/profile WS frame.
Fields: ID, Role, Scope, UnitIndex, Label, UoM, Decimals (all exported; JSON tags camelCase).

### store.Profile  [profile.go:39]
Hello/profile WS frame body: enabled channels only, in sort_order.
Fields: Name string, Units int, FormatID string, Channels []Channel.

### store.EditorChannel  [profile.go:50]
Like Channel plus `Enabled bool` and `SortOrder int` — for GET /api/profile editor view.

### store.EditorProfile  [profile.go:59]
All channels (enabled + disabled) for the editor. Fields: Name, Units, FormatID, Channels []EditorChannel.

### store.SeedChannel  [profile.go:66]
Input shape supplied by main to seed a profile. Fields mirror Channel (no Enabled; all seeded enabled).

### store.ChannelUpdate  [profile.go:257]
Per-channel patch for PUT /api/profile. All fields except ChannelID are pointer (nil = leave unchanged).
Fields: ChannelID string, Enabled *bool, Label *string, UoM *string, Decimals *int, SortOrder *int.

### store.Job  [jobs.go:25]
| Field       | Type   | Notes                                         |
|-------------|--------|-----------------------------------------------|
| ID          | int64  | server-owned                                  |
| Name        | string | required                                      |
| Company     | string | operator/company                              |
| Well        | string | well / location name                          |
| CasingSize  | string | e.g. "9-5/8\""                                |
| JobType     | string | surface / intermediate / production / squeeze |
| Location    | string | field / lease                                 |
| Cementer    | string | foreman / crew lead                           |
| Notes       | string |                                               |
| IsActive    | bool   | exactly one row is active at a time           |
| CreatedAtUS | int64  | unix microseconds                             |
| UpdatedAtUS | int64  | unix microseconds                             |

### store.Segment  [recording.go:25]
Recording marker over the always-on samples store.
| Field       | Type   | Notes                                   |
|-------------|--------|-----------------------------------------|
| ID          | int64  |                                         |
| JobID       | int64  |                                         |
| StartedAtUS | int64  | unix microseconds (same timeline as ts_us) |
| StoppedAtUS | *int64 | nil = open (recording in progress)      |
| CreatedAtUS | int64  |                                         |

### store.SeriesPoint  [series.go:25]
Type alias `[2]float64` — `[ts_us_float64, value]`. uPlot-friendly 2-element array.

### store.Stats  [store.go:236]
| Field    | Type      |
|----------|-----------|
| Rows     | int64     |
| LatestTS | time.Time |

## Database Models (SQLite)  [internal/store/store.go:69 initSchema]

### samples
| Column  | Type    | Constraints      | Notes                        |
|---------|---------|------------------|------------------------------|
| id      | INTEGER | PRIMARY KEY      |                              |
| ts_us   | INTEGER | NOT NULL         | unix microseconds            |
| channel | TEXT    | NOT NULL         | channel id                   |
| value   | REAL    | NOT NULL         |                              |
Indexes: `idx_samples_ts` ON (ts_us); `idx_samples_channel_ts` ON (channel, ts_us) — used by series range scans.

### pump_profiles
| Column        | Type    | Constraints    | Notes                         |
|---------------|---------|----------------|-------------------------------|
| id            | INTEGER | PRIMARY KEY    |                               |
| name          | TEXT    | NOT NULL       |                               |
| units         | INTEGER | NOT NULL DEFAULT 1 | number of pumping units   |
| daq_format_id | TEXT    | NOT NULL       | references code preset        |
| is_active     | INTEGER | NOT NULL DEFAULT 0 | exactly one row = 1       |
| created_at_us | INTEGER | NOT NULL       | unix microseconds             |
| updated_at_us | INTEGER | NOT NULL       | unix microseconds             |

### profile_channels
| Column     | Type    | Constraints                          | Notes                         |
|------------|---------|--------------------------------------|-------------------------------|
| id         | INTEGER | PRIMARY KEY                          |                               |
| profile_id | INTEGER | NOT NULL REFERENCES pump_profiles(id) ON DELETE CASCADE |       |
| channel_id | TEXT    | NOT NULL                             | e.g. "unit1.pressure"         |
| role       | TEXT    | NOT NULL                             | pressure\|rate\|density\|...  |
| scope      | TEXT    | NOT NULL                             | unit\|aggregate\|stage\|job   |
| unit_index | INTEGER | NOT NULL DEFAULT 0                   | 1-based when scope=unit       |
| label      | TEXT    | NOT NULL                             |                               |
| uom        | TEXT    | NOT NULL DEFAULT ''                  |                               |
| decimals   | INTEGER | NOT NULL DEFAULT 2                   |                               |
| enabled    | INTEGER | NOT NULL DEFAULT 1                   |                               |
| sort_order | INTEGER | NOT NULL DEFAULT 0                   |                               |
UNIQUE(profile_id, channel_id). Index: `idx_profile_channels_profile` ON (profile_id).

### jobs
| Column        | Type    | Constraints    | Notes                         |
|---------------|---------|----------------|-------------------------------|
| id            | INTEGER | PRIMARY KEY    |                               |
| name          | TEXT    | NOT NULL       | required; rest default ''    |
| company       | TEXT    | NOT NULL DEFAULT '' |                          |
| well          | TEXT    | NOT NULL DEFAULT '' |                          |
| casing_size   | TEXT    | NOT NULL DEFAULT '' |                          |
| job_type      | TEXT    | NOT NULL DEFAULT '' |                          |
| location      | TEXT    | NOT NULL DEFAULT '' |                          |
| cementer      | TEXT    | NOT NULL DEFAULT '' |                          |
| notes         | TEXT    | NOT NULL DEFAULT '' |                          |
| is_active     | INTEGER | NOT NULL DEFAULT 0 | exactly one row = 1      |
| created_at_us | INTEGER | NOT NULL       | unix microseconds             |
| updated_at_us | INTEGER | NOT NULL       | unix microseconds             |

### recording_segments
| Column        | Type    | Constraints                        | Notes                               |
|---------------|---------|------------------------------------|-------------------------------------|
| id            | INTEGER | PRIMARY KEY                        |                                     |
| job_id        | INTEGER | NOT NULL REFERENCES jobs(id) ON DELETE CASCADE |                       |
| started_at_us | INTEGER | NOT NULL                           | same timeline as samples.ts_us      |
| stopped_at_us | INTEGER | NULLable                           | NULL = open (recording in progress) |
| created_at_us | INTEGER | NOT NULL                           |                                     |
Index: `idx_segments_job` ON (job_id).

DSN pragmas: journal_mode=WAL, synchronous=NORMAL, busy_timeout=5000, foreign_keys=ON. MaxOpenConns=1.

## TypeScript Types  [web/src/types.ts — hand-mirrored from Go, no codegen]

| Type             | Fields                                                                              |
|------------------|-------------------------------------------------------------------------------------|
| Reading          | seq: number, ts: string (RFC3339), values: Record\<string, number\>               |
| Channel          | id, role, scope, unitIndex, label, uom, decimals                                   |
| Profile          | name, units, formatId, channels: Channel[]                                         |
| WSEnvelope       | type: string, reading?: Reading, profile?: Profile                                 |
| Job              | id, name, company, well, casingSize, jobType, location, cementer, notes, isActive, createdAtUs, updatedAtUs |
| JobInput         | name (required), company?, well?, casingSize?, jobType?, location?, cementer?, notes? |
| Segment          | id, jobId, startedAtUs, stoppedAtUs: number\|null, createdAtUs                    |
| RecordingState   | recording: boolean, openSegmentId?: number, jobId?: number                         |

## Tags
#cementer #map #schema #sqlite #go-types #typescript #wire-contract #daqformat #pump-profile #recording-segments

## Links
- [primary.map.md](./primary.map.md)
- [api.map.md](./api.map.md)
- [state.map.md](./state.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
