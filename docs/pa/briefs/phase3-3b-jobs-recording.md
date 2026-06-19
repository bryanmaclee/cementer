---
change-id: phase3-jobs-recording-profiles
slug: 3b-jobs-recording
agent: cementer-go-engineer
model: opus
isolation: worktree
agent-id: ab2e3e7d2b4467f7b
dispatched: 2026-06-18 (Session 5)
dispatched-from-tip: cd71beb
status: dispatched
---

# Dispatch brief (verbatim) — Phase 3b: Jobs + recording segments

> Archived per pa.md §5. Exact prompt sent to `cementer-go-engineer` (background, worktree, opus).
> Paths/SHAs as-at-dispatch.

---

You are the canonical cementer Go dev-agent. Build **Phase 3b — Jobs + recording segments**: the `jobs` + `recording_segments` tables, job CRUD + an active-job concept, recording start/stop/adjust over the continuous store, and minimal client controls. This builds directly on Phase 3a (profiles/api, just landed at commit `cd71beb`). **The headline invariant is axiom #1: recording start/stop must NEVER gate ingestion or the live readout — segments are pure markers over the always-on store.**

═══════════════════════════════════════════════════════════════════════
0) STARTUP-VERIFICATION GATE (FIRST, before any write)
═══════════════════════════════════════════════════════════════════════
- Confirm cwd == your assigned worktree under `/home/bryan-maclee/...`; `git rev-parse --show-toplevel` == that worktree (NOT the integration root `/home/bryan-maclee/cementer`); clean tree. If any fails, STOP and report.
- Go user-local: `export PATH=$HOME/.local/go/bin:$PATH`; verify `go version` (1.26.x).
- `web/dist` is gitignored + `go:embed`-required AND this arc touches the client, so build web: `cd web && npm install && npm run build` (Node 22 + Vite 8 work). Server-only steps can `go build ./internal/...`; binary + E2E need `make build`.

═══════════════════════════════════════════════════════════════════════
1) MAPS — stale; use the normative docs + source as current truth
═══════════════════════════════════════════════════════════════════════
`.claude/maps/` is stamped `ee446c3` (pre-Phase-2) — it does NOT reflect Phase 2 (`internal/daqformat`) or Phase 3a (`pump_profiles`/`profile_channels`, the `internal/api` package, the WS profile frame). Read `structure`/`state`/`api` for the stable pipeline shape, but treat them as a verify-against-source hypothesis. **The CURRENT-TRUTH sources are:** `docs/design/data-model.md` (now carries the realized 3a contract), `docs/changes/phase3-jobs-recording-profiles/scope.md`, and the actual code in `internal/store/`, `internal/api/`, `cmd/cementer/main.go`. Verify against those. (A full map regen happens at the session wrap.) Report whether maps were load-bearing.

═══════════════════════════════════════════════════════════════════════
2) ANTI-PATTERNS — READ BEFORE CODE (BOTH parts)
═══════════════════════════════════════════════════════════════════════
`docs/pa/anti-patterns.md` — Part A (Go) for store/api; **Part B (vanilla-TS, NO framework)** for the client controls (plain TS modules + Vite + direct DOM; no React/Vue/Svelte; localStorage for personal prefs only; keep `types.ts` in sync with Go by hand).

═══════════════════════════════════════════════════════════════════════
3) NORMATIVE SOURCES + RESOLVED DECISIONS
═══════════════════════════════════════════════════════════════════════
- `docs/changes/phase3-jobs-recording-profiles/scope.md` — **AUTHORITATIVE.** Read the Job + Recording-segment data-model sections, the API surface (`/api/jobs*`, `/api/recording/*`), sub-arc **3b**, and the test/verify strategy. Decisions resolved: **D2** (store sole DB owner, single-conn synchronous CRUD — same as 3a), **D7** (segment times = unix-micros over the sample timeline; `stopped_at` NULL = open; adjustable; stages orthogonal), **D8** (job fields, below). Auth deferred. Retention (3c) out of scope.
- `docs/design/data-model.md` — recording model (always store; start/stop are markers; multiple segments per job; adjustable; **stages orthogonal to recording**).

Project axioms (MUST honor):
- **#1 (THE headline for 3b):** raw/live/recording strictly independent. Recording start/stop/adjust ONLY insert/update marker rows — they do NOT touch ingestion, the live readout, or sample storage. Samples store continuously whether recording or not.
- **#5:** stages orthogonal to recording — NEVER reset `vol.stage` (or any stage state) on record start/stop. Introduce zero stage-reset logic.
- **#4 / D2:** store is the SOLE DB owner; new CRUD = synchronous methods on the same `SetMaxOpenConns(1)` connection; NO second `*sql.DB`, NO handler-side DB access; sample `writeLoop` untouched.

═══════════════════════════════════════════════════════════════════════
4) CURRENT CODE (post-3a, verified)
═══════════════════════════════════════════════════════════════════════
- `internal/store/store.go` — `initSchema` now has `samples` + `pump_profiles` + `profile_channels` DDL. Add the `jobs` + `recording_segments` DDL here. `s.db` is the one connection.
- `internal/store/profile.go` — the 3a pattern to MIRROR: synchronous `*Store` methods, transactions for multi-step ops, `queryer` helper, `is_active` invariant via demote-in-tx. Put job/recording methods in NEW files `internal/store/jobs.go` + `internal/store/recording.go` (same single-conn discipline).
- `internal/api/api.go` — `API{st, resetVocab}`, `New(...)`, `Register(mux)` with Go-1.22 pattern routes; helpers `writeJSON`/`writeJSONError`/`writeError`; PUT uses `DisallowUnknownFields`. EXTEND `Register` with the job + recording routes (new handlers; reuse the helpers). Handlers call store methods only.
- `cmd/cementer/main.go` — `api.New(st, ...).Register(mux)` already mounted; `wsEnvelope{Type, Reading, Profile}`. No structural change needed beyond what the routes need (the api already has the store).
- `web/src/` — `readout.ts` (scope-grouped live values), `ws.ts`/`types.ts` (profile + reading frames), `main.ts` (wires Readout + connectLive), `styles.css`, `theme.ts`. Add the job/record controls as a NEW module (`web/src/controls.ts`) wired in `main.ts`; keep `readout.ts` focused on live values.

═══════════════════════════════════════════════════════════════════════
5) TARGET DESIGN — Phase 3b
═══════════════════════════════════════════════════════════════════════
**Schema (store `initSchema`)** — exactly per scope.md, with the D8 job fields:
- `jobs(id, name, company, well, casing_size, job_type, location, cementer, notes, is_active, created_at_us, updated_at_us)` — TEXT cols default ''; `is_active` INTEGER (one active at a time).
- `recording_segments(id, job_id INTEGER NOT NULL REFERENCES jobs(id) ON DELETE CASCADE, started_at_us INTEGER NOT NULL, stopped_at_us INTEGER /*NULL=open*/, created_at_us INTEGER NOT NULL)` + `CREATE INDEX idx_segments_job ON recording_segments(job_id)`.

**Store methods (single-conn; jobs.go + recording.go):**
- Jobs: `CreateJob(Job) (id, error)`, `ListJobs() ([]Job, error)`, `GetJob(id) (Job, ok, error)`, `UpdateJob(id, Job) error`, `ActiveJob() (Job, ok, error)`, `SetActiveJob(id) error` (demote others in a tx, like the profile is_active pattern). Define a `Job` struct with json tags (the wire + DB shape).
- Recording: `StartRecording() (Segment, error)` — requires an active job; if a segment is already open, return it (or a typed "already recording" condition) — do NOT open a second; `StopRecording() (Segment, error)` — close the open segment (set stopped_at_us=now); `RecordingState() (recording bool, openSegmentID int64, jobID int64, error)`; `ListSegments(jobID) ([]Segment, error)`; `AdjustSegment(id, startedAtUS *int64, stoppedAtUS *int64) error` (optional fields; validate started<=stopped when both known). Times = `time.Now().UnixMicro()` — the SAME clock/scale as `samples.ts_us` (so Phase-4 charts can filter samples to segment ranges). `Segment{id, jobID, startedAtUS, stoppedAtUS *int64, createdAtUS}` json-tagged.
- **Guard (axiom-consistent):** changing the active job while a segment is open should be rejected ("stop recording first") — keep the open segment bound to one job.

**API (extend `internal/api`):**
```
GET  /api/jobs                      -> list
POST /api/jobs                      -> create (JSON body: the D8 fields; name required)
GET  /api/jobs/{id}                 -> one (404 if absent)
PUT  /api/jobs/{id}                 -> update (DisallowUnknownFields)
GET  /api/job/active                -> active job or {"active":null}
PUT  /api/job/active                -> {"id": N} set active (400 if N has an open segment conflict? no — 409 if a DIFFERENT job is currently recording)
GET  /api/recording/state           -> {recording, openSegmentId, jobId}
POST /api/recording/start           -> open a segment under the active job (400 if no active job; 409 if already recording, returning the open segment)
POST /api/recording/stop            -> close the open segment (409 if not recording)
GET  /api/recording/segments?job_id=N -> segments for a job
PUT  /api/recording/segments/{id}   -> adjust started_at_us / stopped_at_us (400 on bad ordering / unknown id)
```
Use `r.PathValue("id")` for path params (Go 1.22 routing). Handlers call store methods only; clear JSON errors + status codes; `%w`-wrap server-side.

**Client (vanilla TS — minimal controls, NEW `web/src/controls.ts`):**
- A control strip (above or in the readout header): an **active-job `<select>`** (lists `/api/jobs`, plus a "+ New job…" option that reveals a small inline form with the D8 fields — name required; POST `/api/jobs` then set active); a **Record button** that reads "● Record" when stopped / "■ Stop (mm:ss)" when recording (red indicator + elapsed from the open segment's started_at); a small state line.
- Wire via REST: GET `/api/jobs`, GET `/api/job/active`, GET `/api/recording/state` on load; POST start/stop on click; PUT `/api/job/active` on select. **Poll `/api/recording/state` every ~3 s** (and refresh after actions) so multiple clients reflect record state. No new WS message type (a live WS recording-state push is a deferred Phase-4 nicety).
- Keep it minimal + clean; rich job management + the chart are Phase 4. No framework, no CDN.

═══════════════════════════════════════════════════════════════════════
6) WORK BREAKDOWN — WIP-commit per unit; progress.md appended each step
═══════════════════════════════════════════════════════════════════════
(append to `docs/changes/phase3-jobs-recording-profiles/progress.md`)
1. Schema DDL (jobs + recording_segments + index) in `initSchema`.
2. `store/jobs.go`: Job struct + CRUD + active-job (tx demote); unit tests (round-trip, set-active uniqueness, list).
3. `store/recording.go`: segment start/stop/state/list/adjust; unit tests (start requires active job; double-start returns the open one; stop closes; adjust validates ordering; active-job-change-while-open rejected).
4. `api`: job routes + handler tests (httptest).
5. `api`: recording routes + handler tests (incl. start-without-active-job 400, double-start 409, stop-not-recording 409).
6. Client `controls.ts` + `main.ts` wiring + `types.ts` (Job/Segment ifaces) + `styles.css`.
7. E2E verification (next section).
8. **Docs (landing discipline):** fold the realized `jobs`/`recording_segments` schema + the `/api/jobs*` + `/api/recording/*` contract into `docs/design/data-model.md`. Update README if flags/layout changed.

═══════════════════════════════════════════════════════════════════════
7) VERIFY-BEFORE-CLAIM (do NOT claim done on units alone) — axiom #1 is the headline
═══════════════════════════════════════════════════════════════════════
`make build`, run against the real capture (loop so rows keep flowing):
`./cementer -source captures/capture-2026-06-16T161347-19200-8N1-pressure.bin -format intellisense -replay-interval 40ms -replay-loop=true -data-dir /tmp/ce-3b -addr :8090`
- **AXIOM #1 PROOF (the one that matters):** with recording STOPPED, poll `/debug/stats` (or the DB) twice over ~1s and show `samples` rows CLIMB — ingestion is NOT gated by record state. Then POST `/api/recording/start`, confirm rows keep climbing; POST stop, confirm rows STILL climb. Recording only adds/edits segment rows.
- Job flow: POST `/api/jobs` (a job), `PUT /api/job/active`, GET `/api/job/active` confirms.
- Recording flow: `/api/recording/state` → not recording; start → open segment (stopped_at null) under the active job; state → recording + ids; stop → stopped_at set; `GET /api/recording/segments?job_id=` lists it; `PUT .../segments/{id}` nudges started_at earlier and re-GET confirms (axiom #5 adjustability).
- **Axiom #5:** confirm NO stage reset — there is no code path that writes/zeroes `vol.stage` on start/stop (grep your own diff; state it).
- Error paths: start with no active job → 400; double-start → 409; stop when not recording → 409.
- Client: load the page, create+activate a job, start/stop recording, see the indicator + elapsed (eyeball; note it).
- Static: `gofmt -l` empty, `go vet ./...`, `go test ./...` all green; `make build` clean.
- (If you need a DB query, `sqlite3` is NOT installed — use `/debug/stats`, the API, or a throwaway `_`-prefixed Go helper you create then remove so git stays clean.)

═══════════════════════════════════════════════════════════════════════
8) INVARIANTS
═══════════════════════════════════════════════════════════════════════
- **Axiom #1:** ingestion + live readout never gated by record state (PROVEN, not asserted).
- **Axiom #5:** zero stage-reset logic introduced.
- **Axiom #4/D2:** store sole DB owner; one connection; no handler-side DB; sample writeLoop untouched.
- Segment times share the samples timeline (UnixMicro). gofmt/vet/test green; CGO-free build. Phase-1/2/3a paths still work (synthetic + intellisense replay; profile frame + scope-grouped display intact).

═══════════════════════════════════════════════════════════════════════
9) REPORT BACK (final message = data for the PA)
═══════════════════════════════════════════════════════════════════════
(a) worktree path + branch + final tip SHA; (b) files-touched; (c) exact verify commands + output — ESPECIALLY the axiom-#1 proof (rows climbing while stopped AND while recording), the segment lifecycle, and the error-path codes; (d) maps load-bearing?; (e) deferred items; (f) anything contradicting this brief/scope/design; (g) confirm `data-model.md` updated with the realized jobs/recording contract; (h) confirm axiom #5 (no stage-reset) by pointing at the absence in your diff. `git status` clean in the worktree before reporting DONE.
