---
status: current
last-reviewed: 2026-06-18
change-id: phase3-jobs-recording-profiles
phase: 3
depends-on: Phase 2 landed (internal/daqformat engine + Intellisense preset + IntellisenseChannels vocab, commit 83f036a)
---

# Phase 3 scope — Jobs + Recording segments + Pump Profiles + hello/profile message + scope-grouped display

> Authored Session 5 (2026-06-18) at the user's request ("scope all of Phase 3 first" — review
> before any build dispatch). No build has been dispatched. Auth is **deferred** (user decision,
> Session 5). This doc decomposes Phase 3 into ordered sub-arcs and fixes the concrete schemas /
> message shapes / API surface so each sub-arc is a clean dispatch.

## Goal

Make the Pi a **self-describing, job-aware island** (axiom #3): it knows *what pump it is* (Pump
Profile), *what it's recording* (Jobs + recording segments), and *tells the client* (hello/profile
message) so the client renders aligned to the actual pump — replacing today's id-inference stopgap in
`readout.ts`. All of it persisted on the Pi, all of it strictly preserving raw/live/recording
independence (axiom #1) and single-writer durability (axiom #4).

## Boundary — what Phase 3 IS and ISN'T

| In scope (Phase 3) | Out of scope (→ Phase 4 / later) |
|---|---|
| `pump_profiles` + `profile_channels` tables; seed the active Intellisense profile from `daqformat.IntellisenseChannels()` | uPlot charting (the printable multi-line job chart) |
| hello/**profile** WS message (Pi → client on connect) | the two chart-config scopes (live view + company print template) |
| **scope-grouped** live display (Unit 1 / Aggregate / Stage / Job), only *enabled* channels | per-job print overrides + company-default print template |
| `jobs` + `recording_segments` tables; record start/stop/adjust; active-job concept | full DaqFormat CRUD-in-UI (format stays the Phase-2 code preset, seeded — see D3) |
| Pump Profile + Job CRUD over HTTP (`internal/api`); minimal client job/record controls | **Auth** (deferred this session — Pi is a LAN island) |
| Retention/downsampling-as-code: **designed here, build deferred to 3c/Phase 4** (DD rider #3) | rich job-management UI (lives with the chart in Phase 4) |

## Current state (verified 2026-06-18 — what fits, what changes)

- **`internal/store`** — single `samples(id, ts_us, channel, value)` table + `idx_samples_ts`. Opens
  with `_pragma=foreign_keys(ON)` already → **ready for relational tables**. `SetMaxOpenConns(1)` =
  one serialized connection; sample writes funnel through a single `writeLoop` goroutine (batch +
  WAL). **This is the single DB owner and must stay so** (axiom #4, anti-patterns Part A). New tables
  + CRUD are added HERE, on the same single connection (see D2).
- **`internal/api/`** — empty placeholder. The CRUD/HTTP layer is built here.
- **`cmd/cementer/main.go`** — `wsEnvelope{Type, Reading}` is the only WS message; `serveWS` upgrades,
  registers a `hub.Subscriber`, starts read/write pumps. **Change locus:** extend `wsEnvelope` with a
  `Profile`, send the profile frame to each client on connect, mount the `internal/api` routes.
- **`internal/hub`** — fan-out of committed readings; transport-agnostic. The profile frame is a
  per-connection greeting, **not** a broadcast → send it in `serveWS` directly to the new conn (do
  NOT route it through `hub.Broadcast`).
- **`web/src/readout.ts`** — renders **flat** dynamic cards and **infers** label/uom/decimals from the
  channel id (`describeChannel`, ROLE_INFO). Its own comment: *"Until the pump profile (with real
  labels/units) arrives over the wire, infer..."* → Phase 3 **replaces the inference with the profile**
  and groups by scope. Currently it shows **all 13** Intellisense channels including this rig's
  flat-zero ones (`unit2.*`, `density.2`, `water.rate`) — the PumpProfile is what hides them.
- **`web/src/ws.ts`** — handles only `type==="reading"`. Add `type==="profile"`.
- **`web/src/types.ts`** — `WSEnvelope{type, reading?}`. Add `profile?` + `Profile`/`Channel` ifaces.
- **`web/src/chart/`** — empty (Phase 4).
- **Phase-2 payoff:** `daqformat.IntellisenseChannels()` already returns the 13-channel vocab
  (id/role/scope/unitIndex/uom/label/decimals) — the **seed** for the default profile.

## Data model (concrete)

### Pump Profile (axiom #3 — the pump self-describes)

`DaqFormat` (Phase-2 code) defines *all* channels a format CAN emit; the **PumpProfile** declares what
*this physical pump* actually has (units count, which channels are enabled, any label/uom overrides).
That split is exactly why the single-unit rig should hide `unit2.*` etc. (findings-doc rationale).

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
    channel_id  TEXT    NOT NULL,                 -- e.g. "unit1.pressure" (matches the DaqFormat field map)
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
```

- **D5 (channel storage):** normalized `profile_channels` (per-channel enable/relabel CRUD is the whole
  point) — chosen over a JSON-blob column on `pump_profiles`. *Alternative noted: JSON blob is simpler
  but makes per-channel edits a read-modify-write of the whole blob.*
- **Seed on first run:** if no active profile exists, create one named (e.g.) "Intellisense (this
  pump)" with `daq_format_id="intellisense"`, `units=1`, and one `profile_channels` row per
  `daqformat.IntellisenseChannels()` entry, all `enabled=1`. The operator then disables what this rig
  lacks. (Keep the seed honest: it mirrors the format's full vocab; the operator prunes.)

### Job

```sql
CREATE TABLE IF NOT EXISTS jobs (
    id            INTEGER PRIMARY KEY,
    name          TEXT    NOT NULL,               -- short display name
    company       TEXT    NOT NULL DEFAULT '',    -- operator/company
    well          TEXT    NOT NULL DEFAULT '',    -- well / location name
    casing_size   TEXT    NOT NULL DEFAULT '',    -- e.g. "9-5/8\""
    job_type      TEXT    NOT NULL DEFAULT '',    -- e.g. surface / intermediate / production / squeeze
    location      TEXT    NOT NULL DEFAULT '',    -- field / lease
    cementer      TEXT    NOT NULL DEFAULT '',    -- foreman / crew lead
    notes         TEXT    NOT NULL DEFAULT '',
    is_active     INTEGER NOT NULL DEFAULT 0,     -- the job recordings attach to
    created_at_us INTEGER NOT NULL,
    updated_at_us INTEGER NOT NULL
);
```

- **D8 (job fields) — RESOLVED (S5):** the set above (name, company, well, casing_size, job_type,
  location, cementer, notes + auto created_at). **Editable before 3b** — 3a doesn't use jobs. Optional
  first-class columns the user may still add before 3b: API#, rig/unit #, cement class/slurry, planned
  volume/sacks, hole/depth (else `notes` covers ad-hoc).

### Recording segment (markers over the continuous store — axioms #1 & #5)

```sql
CREATE TABLE IF NOT EXISTS recording_segments (
    id            INTEGER PRIMARY KEY,
    job_id        INTEGER NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    started_at_us INTEGER NOT NULL,               -- a point on the SAME timeline as samples.ts_us
    stopped_at_us INTEGER,                         -- NULL = open (recording in progress)
    created_at_us INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_segments_job ON recording_segments(job_id);
```

- A segment is a **time window over the continuous `samples` store** — nothing is gated or discarded.
- `stopped_at_us NULL` = the live open segment. Start = INSERT open segment under the active job; Stop
  = UPDATE its `stopped_at_us`. A job has **many** segments.
- **Adjustable after the fact** (axiom #5): editing a segment is an UPDATE of its endpoints; a
  forgotten "start" is recoverable because the samples were stored anyway.
- **Stages are orthogonal** (axiom #5): nothing here resets `vol.stage`; stage boundaries come from the
  DAQ, not the record button. Do NOT couple them.

### DaqFormat — stays code (D3)

No `daq_formats` table in Phase 3. The format remains the Phase-2 code preset, referenced by id from
`pump_profiles.daq_format_id`; `main.go -format` still resolves it. Full no-code format-CRUD-in-UI is
deferred. (data-model.md's no-code claim is already satisfied at the *recompile* level by Phase 2;
in-UI editing is the later increment.)

## The hello/profile WS message (Pi → client, on connect)

Extend the envelope (mirror in `types.ts`):

```go
type wsEnvelope struct {
    Type    string         `json:"type"`              // "reading" | "profile"
    Reading *model.Reading `json:"reading,omitempty"`
    Profile *Profile       `json:"profile,omitempty"` // sent ONCE to each client on connect
}

type Profile struct {
    Name     string    `json:"name"`
    Units    int       `json:"units"`
    FormatID string    `json:"formatId"`
    Channels []Channel `json:"channels"` // ENABLED channels only, in sort_order
}
type Channel struct {
    ID       string `json:"id"`
    Role     string `json:"role"`
    Scope    string `json:"scope"`     // unit|aggregate|stage|job|meta
    UnitIndex int   `json:"unitIndex"` // 0 when not unit-scoped
    Label    string `json:"label"`
    UoM      string `json:"uom"`
    Decimals int    `json:"decimals"`
}
```

- **Sent per-connection, not broadcast.** In `serveWS`, after the upgrade and before/at registration,
  write one `{type:"profile", profile:{…}}` frame to that conn (the store provides the active profile's
  enabled channels). Then proceed with the normal live-reading write pump.
- Package placement (engineer's choice; recommend `internal/api` or a small `internal/profile` owns the
  `Profile`/`Channel` wire structs; `wsEnvelope` embeds them). Do NOT put wire structs in
  `internal/model` if it would couple model→store. The JSON shape above is the contract `types.ts`
  mirrors — keep them in sync by hand (anti-patterns Part B: no codegen).

## API surface (`internal/api`, mounted on the existing mux)

REST-ish JSON over the same HTTP server. Handlers call **store methods only** (D2) — never touch
`*sql.DB` directly.

```
# Pump Profile (one active profile per Pi; light "CRUD")
GET    /api/profile                 -> active profile (+ all channels, enabled flag)
PUT    /api/profile                 -> update units / per-channel enabled,label,uom,decimals,sort_order
POST   /api/profile/reset           -> reseed channels from the format's vocab (operator escape hatch)

# Jobs
GET    /api/jobs                    -> list
POST   /api/jobs                    -> create
GET    /api/jobs/{id}               -> one
PUT    /api/jobs/{id}               -> update
GET    /api/job/active              -> active job (or null)
PUT    /api/job/active              -> set active job {id}

# Recording (segments over the active job)
GET    /api/recording/state         -> {recording: bool, openSegmentId?, jobId?}
POST   /api/recording/start         -> open a segment under the active job (server-stamps started_at_us)
POST   /api/recording/stop          -> close the open segment (server-stamps stopped_at_us)
GET    /api/recording/segments?job_id=  -> segments for a job
PUT    /api/recording/segments/{id} -> adjust started_at_us / stopped_at_us (axiom #5 after-the-fact)
```

- `/api/recording/start|stop` **must not** touch ingestion, the live readout, or stage volume — they
  only INSERT/UPDATE a marker row (axiom #1).

## Client (vanilla TS — anti-patterns Part B: NO framework)

- `ws.ts`: handle `type==="profile"` → a `ProfileHandler`; keep `type==="reading"` as-is.
- `types.ts`: add `Profile`/`Channel` ifaces + `profile?` on `WSEnvelope`.
- `readout.ts`: on profile, build **scope groups** — section per `Unit 1`, `Unit 2`, …, `Aggregate`,
  `Stage`, `Job` (meta hidden by default). Each group holds the cards for its enabled channels, using
  the profile's `label`/`uom`/`decimals` (DELETE the `describeChannel`/ROLE_INFO inference path; keep a
  minimal defensive fallback only for a streamed channel absent from the profile). A channel not
  `enabled` never gets a card even if it streams.
- 3b adds **minimal controls**: an active-job selector + a **Record start/stop** button bound to
  `/api/recording/*` + an open-segment indicator. Rich job management + the chart are Phase 4.

## DECISIONS

| # | Decision | Resolution |
|---|---|---|
| D1 | Persistence: SQLite tables vs JSON config | **SQLite tables** (existing DB, FKs on; one store). PA (R1). |
| D2 | CRUD write path vs single-writer axiom #4 | **RESOLVED (S5): adopted.** Store is the single DB owner. CRUD = synchronous store methods on the same `SetMaxOpenConns(1)` connection (serialized with the sample `writeLoop` by the 1-conn pool + WAL + busy_timeout). NO second `*sql.DB`, NO writes from handlers. Samples keep the async batch path. |
| D3 | DaqFormat CRUD now? | **No.** Format stays the Phase-2 code preset, referenced by id; in-UI format CRUD deferred. PA. |
| D4 | Auth | **Deferred** (USER, Session 5). LAN island; physical access ≈ authority. |
| D5 | Profile channel storage | **Normalized `profile_channels`** (per-channel CRUD). PA; JSON-blob alt noted. |
| D6 | Active profile/job | **`is_active` flag**, exactly one active each on the Pi. PA. |
| D7 | Segment time basis | **unix-micros over the sample timeline**; `stopped_at` NULL = open; adjustable via UPDATE; stages orthogonal. From data-model + axioms #1/#5. |
| D8 | Job header fields | **RESOLVED (S5):** name, company, well, casing_size, job_type, location, cementer, notes (+ auto created_at). Editable before 3b (3a doesn't use jobs). |
| D9 | Client job/record UI scope | **Minimal** controls in 3b; rich UI + chart = Phase 4. PA. |
| D10 | Retention/downsampling | **Designed here, build deferred** to 3c/Phase 4 (low urgency at ~7 rows/s). |

## Sub-arc decomposition (ordered; each a separate worktree dispatch)

### 3a — Self-describing pump backbone (recommended first build)
Schema: `pump_profiles` + `profile_channels` + seed. Store: profile read + CRUD methods (single-conn).
WS: `Profile` wire type + per-connection profile frame in `serveWS`. API: `GET/PUT /api/profile` +
`POST /api/profile/reset`. Client: `ws.ts`/`types.ts` profile handling + `readout.ts` scope-grouped
render (drop inference). **Verify:** connect → profile frame arrives → client shows channels grouped by
scope, **only enabled** ones; disable `unit2.*`/`density.2`/`water.rate` via `PUT /api/profile` →
client (on reconnect) hides them; live values still update.

### 3b — Jobs + recording segments
**Confirm D8 job fields first.** Schema: `jobs` + `recording_segments`. Store: job CRUD + active-job +
segment start/stop/adjust methods (single-conn). API: the `/api/jobs*` + `/api/recording/*` routes.
Client: active-job selector + Record start/stop + open-segment indicator. **Verify (axiom #1 is the
headline):** while recording is STOPPED, `/debug/stats` row count still climbs (ingestion ungated);
start → an open segment row appears with `started_at_us`; stop → `stopped_at_us` set; the live readout
never paused; `vol.stage` never reset by the button; adjust a segment endpoint via PUT and re-query.

### 3c — Retention/downsampling-as-code (deferred build; design sketch)
A scoped retention worker: downsample `samples` older than N days to 1 row per channel per M seconds
(configurable via flag/profile), pruning the originals. **Never touches the raw log** (layer-1 stays
the full-fidelity rebuild source). Runs off the single writer. Low urgency (~7 rows/s ⇒ ~600k rows/day;
SQLite handles millions). Build when volume or the user warrants it.

## Test + verify strategy (per pa.md §8 — do not claim done on units alone)

- **Unit:** store CRUD round-trips (create/update/active-flag uniqueness); segment open/close/adjust;
  seed idempotency (second boot doesn't duplicate the profile); profile→wire serialization.
- **E2E 3a:** run the binary against `captures/...-pressure.bin -format intellisense`; `curl
  /api/profile` shows the seeded 13 channels; `PUT` to disable `unit2.pressure`; open `/ws/live` and
  confirm the profile frame lists only enabled channels; load the page and eyeball scope groups.
- **E2E 3b (axiom #1 proof):** with recording stopped, confirm `samples` keeps growing; start/stop and
  confirm segment rows + that ingestion/live were never gated; UPDATE a segment endpoint and re-query.
- **Static:** gofmt/vet/test green; `make build` (CGO-free static binary, web embedded).

## Recommended dispatch (when builds are authorized)

- **Agent:** `cementer-go-engineer`, `model: opus`, `isolation: "worktree"` (per sub-arc).
- **Worktree prime:** Go is user-local (`export PATH=$HOME/.local/go/bin:$PATH`); `web/dist` is
  gitignored + `go:embed`-required → `make web` (Node 22 builds Vite 8) before `go build ./cmd/cementer`,
  or build `./internal/...` only for server-only steps.
- **MAPS — required first read:** regenerate first (they predate Phase 2 — stamp `ee446c3`, missing
  `internal/daqformat`). Then `schema`, `state`, `api`, `structure`, `events` (`.claude/maps/`).
- **Anti-patterns:** Part A (Go) for store/api; **Part B (vanilla-TS, NO framework)** mandatory for the
  `readout.ts`/`ws.ts` work.
- **Crash-recovery:** WIP-commit per work-breakdown unit; `docs/changes/phase3-jobs-recording-profiles/progress.md`.
- **Brief archival:** `docs/pa/briefs/phase3-<subarc>-<slug>.md`.

## Risks / unknowns

- **Axiom #4 (D2)** is the sharpest engineering risk: keep ALL DB access in the store on the one
  connection; a stray `sql.Open` or handler-side write is the failure mode to guard in review.
- **Axiom #1** must be provable, not asserted: the 3b verify explicitly checks ingestion is ungated by
  record state.
- **D8 job fields** is a real blocker for 3b — don't guess the operator's job header.
- **Profile reconnect semantics:** the profile is sent on connect; a profile edit shows after the
  client reconnects (acceptable Phase 3; a live profile-push could be a later nicety).
- **Multi-pump future:** schema supports multiple profiles (`is_active`), but the Pi serves one pump;
  don't over-build multi-profile UI now.
