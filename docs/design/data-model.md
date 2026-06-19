# cementer — configuration-driven data model

This captures the architecture decided after the Phase-1 readout, once the real
shape of cement pumps was clarified. It supersedes the fixed 4-channel assumption.

## Roles & topology

- **The cementer** is the privileged user — the crew foreman (what we earlier called
  "admin"). He carries a laptop and moves between pumps during/across jobs.
- **The Pi is pump-mounted** and **self-describing**. Each Pi is a **standalone
  island** — there is no central server. A pump's spec and DAQ format are configured
  **once on the Pi** and persist there; whichever cementer links gets the right
  channels and interpretation automatically.
- **The laptop is a browser client.** It carries only *personal* preferences (theme,
  personal live-chart view). It does not hold pump definitions.

```
cementer's laptop (browser)  ──HTTP/WS──►  pump-mounted Pi (cementer binary)
  - theme (local)                            - pump profile      (persisted here)
  - personal live-chart view (local)         - DAQ format        (persisted here)
                                             - raw log + SQLite  (durable)
                                             - company print default (bundled in deploy)
                                             - per-job print overrides (with the job)
```

On link, the Pi sends the client a **hello/profile** message describing its channels;
the client renders fields aligned to *that* pump. Stream values are keyed by channel id.

## Pump Profile — *what sensors this pump has*

Configured per pump, stored on the Pi. Pumps vary: 1 or 2 pumping units; variable
counts of pressure transducers, densitometers, rate counters.

```
PumpProfile {
  id, name
  units: int                      // number of pumping units (1, 2, ...)
  channels: Channel[]
}

Channel {
  id            // stable key, e.g. "unit1.pressure", "agg.rate", "vol.stage", "density.1"
  role          // pressure | rate | density | volume | temperature | ... (extensible)
  scope         // unit:N | aggregate | stage | job
  unitIndex?    // set when scope = unit:N
  label         // display label, e.g. "Unit 1 Pressure"
  uom           // unit of measure, e.g. "psi", "bbl/min", "ppg", "bbl"
  decimals      // display precision
  source        // how this channel is produced — see DAQ Format
}
```

### Scope model (multi-unit pumps)

- With multiple pumping units, **psi / rate / volume are almost always per-unit**
  (`scope = unit:N`), plus there are **aggregate** fields (`scope = aggregate`).
- **Volume always has at least two job-level counters:** `vol.stage` (stage volume)
  and `vol.job` (job volume), both `scope = stage|job`. Per-unit volume may also exist.
- Densitometers / rate counters / pressure transducers may be multiple → they are just
  multiple channels in the list (`density.1`, `density.2`, …).

### Realized contract — Phase 3a (built; this is the living spec)

The Pump Profile is persisted on the Pi in the single SQLite store (the store is the
sole DB owner; CRUD is synchronous store methods on the one `SetMaxOpenConns(1)`
connection — never a second `*sql.DB`, never a write from an HTTP handler). Exactly one
profile row is active (`is_active=1`). On first run `main` seeds it from the active
DaqFormat's channel vocab (`daqformat.IntellisenseChannels()` / `SyntheticChannels()`,
all enabled); the seed is idempotent (guarded by `HasActiveProfile`). The store stays
format-agnostic — `main` converts the format vocab to the store's neutral
`SeedChannel`, so `internal/store` does not import `internal/daqformat`.

```sql
CREATE TABLE IF NOT EXISTS pump_profiles (
    id            INTEGER PRIMARY KEY,
    name          TEXT    NOT NULL,
    units         INTEGER NOT NULL DEFAULT 1,     -- number of pumping units
    daq_format_id TEXT    NOT NULL,               -- references the code preset, e.g. "intellisense"
    is_active     INTEGER NOT NULL DEFAULT 0,     -- exactly one row = 1 (the pump this Pi is)
    created_at_us INTEGER NOT NULL,
    updated_at_us INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS profile_channels (
    id          INTEGER PRIMARY KEY,
    profile_id  INTEGER NOT NULL REFERENCES pump_profiles(id) ON DELETE CASCADE,
    channel_id  TEXT    NOT NULL,                 -- e.g. "unit1.pressure"
    role        TEXT    NOT NULL,                 -- pressure|rate|density|volume|meta|...
    scope       TEXT    NOT NULL,                 -- unit|aggregate|stage|job|meta
    unit_index  INTEGER NOT NULL DEFAULT 0,       -- 1-based when scope=unit; 0 otherwise
    label       TEXT    NOT NULL,
    uom         TEXT    NOT NULL DEFAULT '',
    decimals    INTEGER NOT NULL DEFAULT 2,
    enabled     INTEGER NOT NULL DEFAULT 1,       -- the pump physically has this channel
    sort_order  INTEGER NOT NULL DEFAULT 0,
    UNIQUE(profile_id, channel_id)
);
CREATE INDEX IF NOT EXISTS idx_profile_channels_profile ON profile_channels(profile_id);
```

**hello/profile WS message** (Pi → client, sent ONCE per connection in `serveWS`,
directly to the new conn — a greeting, never routed through `hub.Broadcast`; it never
gates the live readout). The frame lists ENABLED channels only, in `sort_order`. The
`wsEnvelope.type` discriminates `"reading"` vs `"profile"`. The JSON shape (mirrored by
hand in `web/src/types.ts`, no codegen):

```jsonc
// {type:"profile", profile: Profile}
{
  "name": "Intellisense (this pump)",
  "units": 1,
  "formatId": "intellisense",
  "channels": [                       // ENABLED only, in sort_order
    { "id": "unit1.pressure", "role": "pressure", "scope": "unit",
      "unitIndex": 1, "label": "Unit 1 Pressure", "uom": "psi", "decimals": 0 }
    // ...
  ]
}
```

The client (vanilla TS, no framework) renders cards grouped by scope —
`Unit 1`, `Unit 2`, … (by `unitIndex`), then `Aggregate`, `Stage`, `Job`; `meta`-scoped
channels are hidden by default. A streamed channel absent from the enabled profile gets
a minimal defensive card in a trailing "Other" group; a disabled channel never gets a
card. The old id-inference stopgap in `readout.ts` is removed.

**Profile HTTP API** (`internal/api`, mounted on the same mux; handlers call store
methods only):

```
GET    /api/profile          -> active profile incl. ALL channels (each with its `enabled` flag + sortOrder)
PUT    /api/profile          -> update units + per-channel enabled/label/uom/decimals/sortOrder (omitted fields unchanged)
POST   /api/profile/reset    -> reseed channels from the active format's vocab (operator escape hatch)
```

Note the asymmetry: the **GET** returns *all* channels (the editor must see disabled
ones to re-enable them) while the **WS profile frame** sends *enabled only*. A profile
edit takes effect for a client on its next (re)connect (the profile is a per-connect
greeting; a live profile push is a later nicety).

## DAQ Format — *how the wire maps to those channels* (no-code)

The serial format is a property of the pump (the Pi is pump-mounted, so it stays put
until the Pi moves). Two presets are common in the wild — **Intellisense** and
**MD Totco** — with a growing tail of boutique one-offs. Adapting to a new format must
be **manual configuration, strictly no code.**

```
DaqFormat {
  id, name                 // "Intellisense", "MD Totco", or a custom name
  delimiter                // e.g. ","   (ASCII text lines — confirmed)
  hasHeader: bool
  timestamp?               // field index/name + parse hint, or "server stamps"
  fields: FieldMap[]       // raw column -> channel
}

FieldMap {
  column        // index or header name in the raw line
  channelId     // which Channel this column feeds
  transform?    // optional, no-code: scale, offset
}

// Derived channels (aggregates a pump does NOT emit) are computed, not mapped:
ComputedChannel {
  channelId     // e.g. "agg.rate"
  op            // sum | mean | ...
  inputs        // [ "unit1.rate", "unit2.rate" ]
  // plus scale/offset
}
```

- **Field mapping** covers pumps that already emit a value (incl. aggregates they
  provide).
- **Compute layer** (sum / scale / offset, configured in the UI — still no-code) covers
  pumps that *don't* emit aggregates: `agg.rate = sum(unit1.rate, unit2.rate)`, etc.
- Presets (Intellisense, MD Totco) ship as starting templates the cementer can clone and
  adjust. **The Intellisense preset is now defined from real wire data** (captured 2026-06-16
  off a live unit — 19200 8N1, 14 columns, no header, `HH:MM:SS`-uptime timestamp; column map
  empirically confirmed): see `docs/changes/phase2-intellisense-daqformat/intellisense-wire-capture-2026-06-16.md`.
  The **MD Totco** preset is still undefined (unit not yet accessible) — stay format-agnostic for it.

**Durability is unaffected by format changes:** the Pi appends every raw line to the raw
log *before* any mapping. A wrong or edited mapping only re-interprets data; it never
loses it. The structured store is keyed by channel id and can be rebuilt from the raw log.

## Recording, live, and raw — three separate concerns

Easy to conflate, deliberately independent:

- **Raw capture is always on.** Every byte is appended to the raw log the moment it
  arrives, regardless of everything below. Pure durability; never gated.
- **The live readout is always live.** While the pump is on, the cementer always sees
  current values. The live view never pauses and is unaffected by recording state.
- **Recording is the cementer's start/stop** and bounds *what becomes the job* — what the
  chart plots and what the job log covers — so idle / warmup / cleanup periods don't
  obfuscate the relevant data.

**Recording model — always store; start/stop are markers.** Structured samples are stored
continuously (not gated). A recording segment is a marker over that continuous store:

```
RecordingSegment { job_id, id, started_at, stopped_at? }   // stopped_at null = open
```

The chart and job log default to showing only data inside segments. Because nothing is
discarded, boundaries are **adjustable after the fact** (nudge a start earlier, trim a
stop) and a forgotten "start" is fully recoverable. A job may have **multiple segments**.

**Stages are a separate concept.** Stage volume (`vol.stage`) resets per stage, job volume
(`vol.job`) is cumulative — but stages are driven independently (by the pump/DAQ or a
separate stage marker), **not** by the record start/stop button. Do not reset stage volume
on record-start; recording segments and stages are orthogonal and may not line up.

### Realized contract — Phase 3b (built; this is the living spec)

Jobs and recording segments are persisted on the Pi in the single SQLite store (same
single-owner discipline as 3a: synchronous store methods on the one `SetMaxOpenConns(1)`
connection — never a second `*sql.DB`, never a write from an HTTP handler). Exactly one
job is active (`is_active=1`): the job new recording segments open under.

```sql
CREATE TABLE IF NOT EXISTS jobs (
    id            INTEGER PRIMARY KEY,
    name          TEXT    NOT NULL,
    company       TEXT    NOT NULL DEFAULT '',    -- operator/company
    well          TEXT    NOT NULL DEFAULT '',    -- well / location name
    casing_size   TEXT    NOT NULL DEFAULT '',    -- e.g. "9-5/8\""
    job_type      TEXT    NOT NULL DEFAULT '',    -- surface / intermediate / production / squeeze
    location      TEXT    NOT NULL DEFAULT '',    -- field / lease
    cementer      TEXT    NOT NULL DEFAULT '',    -- foreman / crew lead
    notes         TEXT    NOT NULL DEFAULT '',
    is_active     INTEGER NOT NULL DEFAULT 0,     -- exactly one row = 1 (the active job)
    created_at_us INTEGER NOT NULL,
    updated_at_us INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS recording_segments (
    id            INTEGER PRIMARY KEY,
    job_id        INTEGER NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    started_at_us INTEGER NOT NULL,               -- a point on the SAME timeline as samples.ts_us
    stopped_at_us INTEGER,                          -- NULL = open (recording in progress)
    created_at_us INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_segments_job ON recording_segments(job_id);
```

**Recording is marker-only (AXIOM #1, proven not asserted).** `StartRecording` inserts an
open segment (`stopped_at_us NULL`) under the active job; `StopRecording` updates its
`stopped_at_us`. These touch *only* `recording_segments` — they never gate ingestion, the
live readout, or stage volume. The sample `writeLoop` is untouched: samples are stored
continuously whether or not a segment is open (E2E-verified — `samples` row count climbs
identically while stopped, while recording, and after stop). Segment timestamps are
`time.Now().UnixMicro()`, the **same clock/scale as `samples.ts_us`**, so a Phase-4 chart
can filter samples to `[started_at_us, stopped_at_us)`. A job has **many** segments; a
forgotten start is recoverable because the samples were stored anyway, and endpoints are
**adjustable after the fact** (`AdjustSegment`, axiom #5). Switching the active job while a
segment is open is **refused** (the open segment stays bound to one job — stop first).

**Job + recording HTTP API** (`internal/api`, mounted on the same mux; handlers call store
methods only). JSON shapes mirrored by hand in `web/src/types.ts` (`Job`, `JobInput`,
`Segment`, `RecordingState`):

```
GET    /api/jobs                       -> list (newest first)
POST   /api/jobs                       -> create (name required; 201 + the new job; not auto-active)
GET    /api/jobs/{id}                  -> one (404 if absent)
PUT    /api/jobs/{id}                  -> update descriptive fields (DisallowUnknownFields)
GET    /api/job/active                 -> the active job, or {"active":null}
PUT    /api/job/active                 -> {"id":N} set active (404 unknown; 409 if a DIFFERENT job is recording)

GET    /api/recording/state            -> {recording, openSegmentId?, jobId?}
POST   /api/recording/start            -> open a segment under the active job
                                          (400 no active job; 409 + the open segment if already recording)
POST   /api/recording/stop             -> close the open segment (409 if not recording)
GET    /api/recording/segments?job_id=N -> a job's segments (chronological)
PUT    /api/recording/segments/{id}    -> nudge started_at_us / stopped_at_us
                                          (404 unknown id; 400 bad ordering — started must be <= stopped)
```

The client (vanilla TS, no framework) adds a minimal control strip between the readout
header and the live values: an active-job `<select>` (with an inline "+ New job…" form for
the D8 fields), a Record/Stop button showing the open segment's elapsed time, and a state
line. It polls `GET /api/recording/state` every ~3 s (and refreshes after each action) so
multiple clients converge on the record state — there is no WS recording-state push yet (a
deferred Phase-4 nicety). Rich job management + the printable chart are Phase 4.

## Two chart-config scopes

1. **Cementer's live view** (personal) — what *he* sees in his client: which channels/
   lines, time window, colors. Stored client-side (per laptop). Does not affect others.
2. **Company printed-chart standard** — what every job's printout looks like. A **company
   default** template (bundled with the deploy, since Pis are islands) that the cementer
   can **tweak per job**; per-job overrides are stored with the job on the Pi. The company
   default is change-controlled (updated via deploy/config), not casually editable.

## Client customization

- **Theme:** dark (default) / light, plus room for small niceties. A per-client local
  preference (localStorage); not synced.

## Storage location — trivially flippable

- `-data-dir` flag (and `CEMENTER_DATA_DIR` env) selects where the SQLite DB + raw logs
  live. **Dev:** the Pi's built-in storage (e.g. `./data`). **Prod:** an external SSD
  (e.g. `/mnt/ssd/cementer-data`). One value to change; nothing else moves.

## Build order implied by this model

1. (now, agnostic) Client renders **dynamic channels** from the stream instead of a fixed
   set; theme toggle; storage env. None of this depends on the real format.
2. (after the Intellisense CSV) Define the Intellisense `DaqFormat` preset; build the
   format mechanism (mapping + compute) on the Pi.
3. Job CRUD (company/casing/…) + **recording segments (start/stop markers)**; Pump Profile
   CRUD + the hello/profile message + scope-grouped display.
4. Charting (uPlot) with the two config scopes (chart shows only recorded segments by
   default); printing with the company default + per-job overrides.
