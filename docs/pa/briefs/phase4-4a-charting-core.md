---
change-id: phase4-charting-printing
slug: 4a-charting-core
agent: cementer-go-engineer
model: opus
isolation: worktree
agent-id: ac1820ce112d2db48
dispatched: 2026-06-19 (Session 5)
dispatched-from-tip: cf46ab3
status: dispatched
---

# Dispatch brief (verbatim) — Phase 4a: charting core (series API + uPlot live + historical chart)

> Archived per pa.md §5. Exact prompt sent to `cementer-go-engineer` (background, worktree, opus).
> Paths/SHAs as-at-dispatch.

---

You are the canonical cementer Go dev-agent. Build **Phase 4a — the charting core**: a samples series/range query API + the **uPlot** chart, with the **live view becoming a rolling real-time chart** (replacing the value readout) and a **historical per-job chart** over recorded segments. This is the project centerpiece; build it clean. Builds on Phase 3 (profiles + jobs + recording_segments, commit `cf46ab3`).

═══════════════════════════════════════════════════════════════════════
0) STARTUP-VERIFICATION GATE (FIRST)
═══════════════════════════════════════════════════════════════════════
- cwd == your worktree under `/home/bryan-maclee/...`; `git rev-parse --show-toplevel` == worktree (NOT `/home/bryan-maclee/cementer`); clean tree. Else STOP + report.
- Go user-local: `export PATH=$HOME/.local/go/bin:$PATH`; `go version` (1.26.x).
- This arc adds a web dep (uPlot) + changes the client, so: `cd web && npm install` (adds uplot), `npm run build` (Vite 8 / Node 22). Binary + E2E need `make build`. `web/dist` is gitignored + `go:embed`-required.

═══════════════════════════════════════════════════════════════════════
1) MAPS — stale; use normative docs + source as current truth
═══════════════════════════════════════════════════════════════════════
`.claude/maps/` is stamped `ee446c3` (pre-Phase-2/3) — does NOT reflect daqformat, profiles, jobs, recording, or the api package. Read `structure`/`state`/`api` for the stable pipeline shape only; **current truth = `docs/design/data-model.md` (realized 3a/3b contracts), `docs/changes/phase4-charting-printing/scope.md`, and the live code in `internal/store/`, `internal/api/`, `cmd/cementer/main.go`, `web/src/`.** (Full map regen at wrap.) Report whether maps were load-bearing.

═══════════════════════════════════════════════════════════════════════
2) ANTI-PATTERNS — READ BEFORE CODE (BOTH parts)
═══════════════════════════════════════════════════════════════════════
`docs/pa/anti-patterns.md`. Part A (Go) for store/api. **Part B (vanilla-TS) is critical here:** the client is NO framework — plain TS modules + Vite + direct DOM. **uPlot is a focused charting LIBRARY (ratified in the storage+viz deep-dive), NOT a framework — it is allowed.** Bundle it via npm/Vite so it's embedded and works OFFLINE on the Pi (NO CDN). localStorage is for personal prefs only (axiom #3).

═══════════════════════════════════════════════════════════════════════
3) NORMATIVE SOURCES + RESOLVED DECISIONS
═══════════════════════════════════════════════════════════════════════
- `docs/changes/phase4-charting-printing/scope.md` — **AUTHORITATIVE.** Read the goal, the user domain decisions, "Current state", sub-arc **4a**, and the Decisions table. (4b print/PDF is OUT of this arc.)
- `docs/design/data-model.md` — the two chart-config scopes; "the chart defaults to showing only recorded segments."

User domain decisions baked into 4a:
- **X-axis = TIME.**
- **Traces = ALL enabled profile channels, auto-grouped by ROLE** — one uPlot scale per role/uom (pressure→psi, rate→bbl/min, density→ppg, volume→bbl), axes auto-assigned, a distinct color per channel.
- **Live view REPLACES the value readout** with a rolling real-time chart — BUT keep current values glanceable in the chart **legend** (the readout's value-at-a-glance utility must survive).
- (PDF/printing is 4b — do NOT build it here.)

Axioms: **#1** the chart is READ-ONLY over the store — it NEVER gates or touches ingestion / the live stream / recording; the new series query is a read on the single store connection (the store stays the sole DB owner, **#4/D2** — no second `*sql.DB`, no handler-side DB). **#3** personal live-view config is per-laptop localStorage; pump/job data stays on the Pi.

═══════════════════════════════════════════════════════════════════════
4) CURRENT CODE (verified 2026-06-19)
═══════════════════════════════════════════════════════════════════════
- `internal/store/store.go` — `samples(ts_us, channel, value)` + `idx_samples_ts`; `SetMaxOpenConns(1)`; sample `writeLoop`. NO historical query. Add a `(channel, ts_us)` composite index + a series read here.
- `internal/store/profile.go`, `jobs.go`, `recording.go` — the single-conn method patterns to MIRROR; `RecordingState`/`ListSegments`/`ActiveJob`/`Segment`/`Job` exist (use them for the job chart).
- `internal/api/api.go` + `jobs.go` — `API{st,...}`, `Register(mux)`, helpers `writeJSON`/`writeJSONError`/`writeError`, `decodeStrict`, `pathID`; Go-1.22 routing. EXTEND with the series routes.
- `cmd/cementer/main.go` — `wsEnvelope{Type, Reading, Profile}`; `serveWS` sends the profile greeting then streams readings; `api.New(...).Register(mux)` mounted; SPA mounted.
- `web/src/` — `readout.ts` currently owns the screen (header/status/theme + a profile-driven scope-grouped VALUE GRID + footer + a 3b controls host); `ws.ts` dispatches `profile` + `reading`; `types.ts` has `Profile`/`Channel`/`Reading`/`Job`/`Segment`/`RecordingState`; `controls.ts` (3b job/record strip); `main.ts` wires Readout + Controls + connectLive; `styles.css`; `theme.ts`. `web/src/chart/` is EMPTY.

═══════════════════════════════════════════════════════════════════════
5) TARGET DESIGN — Phase 4a
═══════════════════════════════════════════════════════════════════════
**Server:**
- Composite index: `CREATE INDEX IF NOT EXISTS idx_samples_channel_ts ON samples(channel, ts_us)` in `initSchema`.
- `store` series read (single-conn): `Series(fromUS, toUS int64, channels []string, maxPerChannel int) (map[string][][2]float64, error)` — for each requested channel, the [ts_us, value] points in [from,to], **decimated** when a channel would exceed `maxPerChannel`: bucket the range and emit **min AND max per bucket** (preserve spikes — pressure spikes matter), so the cap is ~2–4k points/channel. Empty channels → empty slices. If `channels` is empty, use all distinct channels in range.
- API (extend `Register`; handlers call store methods only):
  - `GET /api/samples?from=<us>&to=<us>&channels=a,b,c` → `{ "series": { "<channel>": [[tsUs,val],...] } }` (raw range, decimated by the cap).
  - `GET /api/jobs/{id}/series?channels=` → `{ "segments": [{"id","startedAtUs","stoppedAtUs"}], "series": { "<channel>": [[tsUs,val],...] } }` covering the job's segments (union span; samples WITHIN segments — the recorded data; gaps between segments left as gaps). 404 if no such job. Decimated by the cap.
  - Validate from<=to, sane caps; clear JSON errors.
**Client (vanilla TS + uPlot, `web/src/chart/`):**
- Add `uplot` (npm). Import its CSS too (bundled, offline).
- **LiveChart** (`web/src/chart/livechart.ts`): a rolling uPlot. `applyProfile(Profile)` builds series from ENABLED channels, **one uPlot scale per role/uom**, axes auto-assigned (sensible left/right; don't crowd — extra roles can share or use a compact axis), distinct per-channel colors, and a **legend that shows each channel's latest value** (uPlot legend, live-updated). `push(Reading)` appends to a rolling ring buffer (default window e.g. 5 min) and updates. X-axis = time.
- **JobChart** (`web/src/chart/jobchart.ts`): given a job id, fetch `/api/jobs/{id}/series`, render vs time with **segment shading** (uPlot hooks/plugin to shade [startedAt,stoppedAt] bands), same role-grouped axes. Pan/zoom is fine to enable.
- **View integration:** the live rolling chart REPLACES `readout.ts`'s value grid as the default view. Restructure so `readout.ts` (or a renamed shell) keeps the header/status/theme/footer + the 3b controls host, and hosts a **view area** with a simple **Live | Job History** toggle: Live shows LiveChart; Job History shows JobChart for the active/selected job. Keep the connection/stale status + theme. `main.ts` wires `onProfile`→liveChart.applyProfile, `onReading`→liveChart.push (mirroring today's readout wiring).
- **Personal live-view config (scope #1, localStorage):** per-channel line on/off, rolling-window length, (optionally) colors. Persist per-laptop; reflect in the LiveChart. Keep it minimal but real.
- Keep `theme.ts` working (chart colors should respect dark/light or at least be legible in both).

═══════════════════════════════════════════════════════════════════════
6) WORK BREAKDOWN — WIP-commit per unit; progress.md appended each step
═══════════════════════════════════════════════════════════════════════
(create/append `docs/changes/phase4-charting-printing/progress.md`)
1. Composite index + `store.Series` (decimation: min/max-per-bucket) + unit tests (range boundaries, channel filter, empty range, decimation cap correctness incl. spike preservation).
2. API `/api/samples` + `/api/jobs/{id}/series` + handler tests (httptest; 404; validation).
3. uPlot dep + `chart/livechart.ts` (profile→series, role-scales, legend latest values, rolling push).
4. View restructure: Live|Job toggle in the shell; LiveChart replaces the value grid; main.ts wiring; preserve controls + status + theme.
5. `chart/jobchart.ts` (fetch job series + segment shading).
6. Personal live-view config (localStorage) + styles.css.
7. E2E verification (next section).
8. **Docs (landing discipline):** fold the realized series API (`/api/samples`, `/api/jobs/{id}/series`) into `docs/design/data-model.md`; update README (uPlot dep, the chart view).

═══════════════════════════════════════════════════════════════════════
7) VERIFY-BEFORE-CLAIM (not "tests pass")
═══════════════════════════════════════════════════════════════════════
`make build`, run against the real capture (loop):
`./cementer -source captures/capture-2026-06-16T161347-19200-8N1-pressure.bin -format intellisense -replay-interval 40ms -replay-loop=true -data-dir /tmp/ce-4a -addr :8090`
- `GET /api/samples?from=0&to=<bignum>&channels=agg.pressure,unit1.pressure` → series arrays with plausible values (agg.pressure peaks ~1306); confirm decimation caps a large range.
- Create+activate a job, record a short segment (POST start, wait, POST stop), then `GET /api/jobs/{id}/series` → segments[] + series within them.
- **Axiom #1 (still holds):** `/debug/stats` rows climb throughout regardless of any /api/samples or chart calls (the chart is read-only).
- Load `http://localhost:8090/`: the LIVE view is now a rolling chart updating in real time, all enabled channels grouped by role, legend showing live values; switch to Job History and load the recorded job → segment-shaded historical chart. (Eyeball; describe what you saw + paste any console errors.)
- Confirm OFFLINE bundling: uPlot + its CSS are in `web/dist` (no CDN/network ref in the built HTML/JS).
- Static: `gofmt -l` empty; `go vet ./...`; `go test ./...` all green; `make build` clean (CGO-free).

═══════════════════════════════════════════════════════════════════════
8) INVARIANTS
═══════════════════════════════════════════════════════════════════════
- Chart is READ-ONLY: ingestion/live/recording never gated or touched (axiom #1); the series query is a read on the single store connection (axiom #4/D2 — no second *sql.DB, no handler-side DB).
- uPlot bundled offline (no CDN); single static CGO-free binary still builds; gofmt/vet/test green.
- All enabled channels, role-grouped axes, X=time; live legend shows current values (readout utility preserved).
- Phase 1/2/3 paths intact (profile frame, controls, jobs/recording still work).

═══════════════════════════════════════════════════════════════════════
9) REPORT BACK (final message = data for the PA)
═══════════════════════════════════════════════════════════════════════
(a) worktree path + branch + final SHA; (b) files-touched; (c) verify commands + output (the /api/samples + /api/jobs/{id}/series JSON shape, decimation evidence, axiom-#1 rows-climb, the offline-bundle check, gofmt/vet/test); (d) what the live + job charts looked like (you loaded the page) + any console errors; (e) maps load-bearing?; (f) deferred items (4b print/PDF is deferred by design); (g) anything contradicting the brief/scope/design; (h) confirm data-model.md updated with the series API. `git status` clean in the worktree before DONE.
