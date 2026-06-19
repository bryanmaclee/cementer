# Phase 3a — Self-describing pump backbone — progress

Append-only. Timestamped done / next / blockers.

## 2026-06-18 — startup
- DONE: startup-verification gate passed (worktree root != integration root; clean tree; Go 1.26.4).
- DONE: read maps (structure/state/schema/api/events), anti-patterns A+B, scope.md (phase3), data-model.md.
- DONE: read current source (store.go, presets.go, daqformat.go, main.go, model.go, web/src/*).
- DONE: baseline `go test ./internal/...` green (daqformat + parser pass; others have no tests).
- NEXT: WBS-1 schema DDL (pump_profiles + profile_channels) in store initSchema.
- BLOCKERS: scope.md exists only on main (not the worktree branch yet) — read from main path for reference (allowed: read main, write worktree).

## 2026-06-18 — store backbone
- DONE: WBS-1 schema DDL (pump_profiles + profile_channels + index) in store.initSchema.
- DONE: WBS-2/3 internal/store/profile.go — Profile/Channel + EditorProfile/EditorChannel wire types (json tags = the client contract); SeedChannel input vocab; HasActiveProfile, SeedActiveProfile (demotes prior active), ActiveProfile (enabled-only, sort_order), ActiveEditorProfile (all channels w/ enabled flag), UpdateActiveProfile (units + per-channel patch), ResetActiveProfileChannels. All single-conn synchronous methods on s.db (axiom #4 / D2). Store does NOT import daqformat (axiom #2).
- DONE: profile_test.go — round-trip, seed idempotency (HasActiveProfile guard + edit-survives-reboot), set-active uniqueness, channel enable/label/uom/decimals/sort update, unknown-channel-errors-atomic, reset, empty-store. All green.
- NEXT: WBS-4 main.go — idempotent seed from daqformat vocab; pass profile provider into serveWS; per-connection profile frame; mount api routes.
- BLOCKERS: none.

## 2026-06-18 — api + main wiring
- DONE: WBS-5 internal/api/api.go — API{store, resetVocab}; GET /api/profile (all channels via ActiveEditorProfile), PUT /api/profile (units + per-channel patch DTO w/ DisallowUnknownFields + validation), POST /api/profile/reset. Handlers call store methods only; clear JSON errors + status codes (400 bad body/unknown channel, 404 no profile, 500 server). httptest tests all green.
- DONE: WBS-4 main.go — wsEnvelope += Profile *store.Profile; idempotent seed (HasActiveProfile guard) from seedVocab(format) (main converts daqformat.Channel -> store.SeedChannel, so store never imports daqformat); serveWS now takes activeProfile provider and writes ONE {type:profile} frame to the new conn before the live write pump (greeting, not broadcast); api.New(...).Register(mux) mounted; reset vocab provider injected.
- DONE: web/dist primed (npm install + build); full `go build ./...` + `go vet ./...` + gofmt clean.
- NEXT: WBS-6 client types.ts + ws.ts profile handling; WBS-7 readout.ts scope-grouped render (drop inference) + main.ts + styles.css.
- BLOCKERS: none.

## 2026-06-18 — client + E2E
- DONE: WBS-6 types.ts (Profile/Channel ifaces, profile? on WSEnvelope) + ws.ts (ProfileHandler param, type==="profile" dispatch, kept reading).
- DONE: WBS-7 readout.ts rewritten — applyProfile(p) builds scope groups (Unit N by unitIndex, then Aggregate/Stage/Job; meta hidden), cards use profile label/uom/decimals; describeChannel/ROLE_INFO/PART_LABEL inference DELETED; defensive "Other" group only for a streamed id absent from the enabled profile; status/seq/stale footer + theme preserved. main.ts wires onProfile. styles.css: .content + .group/.group-title/.group-grid (was .grid). tsc strict clean; vite build clean; make build clean.
- DONE: full gate — gofmt -l empty, go vet ./... clean, go test ./... green (api+store new, daqformat+parser existing), make build clean.
- DONE: E2E vs captures/...-pressure.bin -format intellisense: seed logged 13 channels; GET /api/profile shows 13 w/ enabled flags; PUT disable unit2.pressure/unit2.rate/water.rate/density.2 -> 200, GET confirms; WS first frame = {type:profile} with EXACTLY the 9 enabled (4 disabled absent); restart same data-dir -> NO second seed log line, one profile, the 4 disable edits persisted (idempotent); /debug/stats rows climbed 9724->10543->11193 across profile GET+PUT calls (axiom #1 ungated); reading frames carry sane real values (vol.stage 2.1); / serves embedded SPA referencing built dist hashes. Throwaway _wscheck dir created+removed (git stayed clean).
- NEXT: WBS-9 fold realized contract into docs/design/data-model.md; README check.
- BLOCKERS: none.

# Phase 3b — Jobs + recording segments — progress

## 2026-06-18 — 3b startup
- DONE: startup-verification gate passed (worktree root != integration root cementer; clean tree; Go 1.26.4; tip cd71beb = Phase 3a landed).
- DONE: read scope.md (3b sections + D2/D7/D8 decisions), anti-patterns A+B, data-model.md recording section.
- DONE: read current post-3a source: store.go (initSchema has samples+pump_profiles+profile_channels), profile.go (the single-conn CRUD pattern to mirror, queryer helper, is_active demote-in-tx), api.go (API{st,resetVocab}, helpers, DisallowUnknownFields), main.go (wsEnvelope, api.New().Register), web/src/* (readout/ws/types/main/styles).
- DONE: baseline `go test ./internal/...` green (api, daqformat, parser, store ok; no-test pkgs noted).
- NEXT: WBS-1 schema DDL (jobs + recording_segments + index) in initSchema.
- BLOCKERS: none.

## 2026-06-18 — 3b store backbone
- DONE: WBS-1 schema DDL (jobs + recording_segments + idx_segments_job) in store.initSchema, after profile_channels. FK job_id REFERENCES jobs(id) ON DELETE CASCADE; stopped_at_us NULL = open.
- DONE: WBS-2 internal/store/jobs.go — Job struct (json tags = client contract; id/isActive/timestamps server-owned); CreateJob (name required, NOT auto-active), ListJobs (newest first), GetJob (ok bool), UpdateJob (descriptive fields + updated_at; ErrNoSuchJob), ActiveJob, SetActiveJob (demote-in-tx is_active invariant; REFUSES with ErrRecording if a DIFFERENT job has an open segment — axiom-consistent bind). All single-conn synchronous on s.db (axiom #4/D2).
- DONE: WBS-3 internal/store/recording.go — Segment struct (StoppedAtUS *int64 = nullable JSON null); typed conditions ErrNoActiveJob/ErrRecording/ErrNotRecording/ErrNoSuchSegment/ErrBadSegmentRange; StartRecording (requires active job; double-start returns the OPEN segment + ErrRecording, no second open; marker-only insert), StopRecording (closes open; ErrNotRecording), RecordingState, ListSegments (chronological), AdjustSegment (optional started/stopped; validates ordering; ErrNoSuchSegment). Times = time.Now().UnixMicro() = SAME scale as samples.ts_us. NO ingestion/live/stage touch — markers only (axioms #1 & #5).
- DONE: jobs_test.go + recording_test.go — round-trip, name-required, missing-id, list-newest, set-active uniqueness, set-active-rejected-while-recording (+ same-job allowed), start-requires-active-job, full start/stop/state lifecycle, double-start-returns-open (one open row), stop-not-recording, multiple-segments-per-job, adjust-moves-start, adjust-bad-ordering, adjust-missing. `go test ./internal/store/` green.
- NEXT: WBS-4 api job routes + handler tests; WBS-5 api recording routes + handler tests.
- BLOCKERS: none.

## 2026-06-18 — 3b api layer
- DONE: WBS-4/5 internal/api/api.go Register extended with 6 job routes + 5 recording routes (Go-1.22 method patterns, {id} path params); internal/api/jobs.go all handlers (call store methods only — axiom #4/D2). decodeStrict (DisallowUnknownFields) + pathID helpers. Status mapping: 201 create, 200 ok, 400 (bad body/missing name/bad id/no-active-job/bad ordering/missing job_id), 404 (unknown job/segment), 409 (double-start returns open segment, stop-not-recording, set-active-while-recording). startRecording/stopRecording = marker insert/update only (axiom #1; no ingestion/live/stage touch — axiom #5).
- DONE: added store.GetSegment(id) (single-conn read) so adjust handler echoes the refreshed row without handler-side DB.
- DONE: internal/api/jobs_test.go (httptest) — create+get, name-required-400, unknown-field-400, get-404, bad-id-400, update, update-missing-404, active-null, set-active+get, set-active-missing-404, set-active-while-recording-409, full recording lifecycle (state/start/stop/segments), start-no-active-job-400, double-start-409-returns-open, stop-not-recording-409, segments-missing-job_id-400, adjust-moves-start+re-GET, adjust-bad-ordering-400, adjust-missing-404, segment-timeline-in-micros sanity. `go test ./internal/...` green.
- NEXT: WBS-6 client controls.ts + main.ts + types.ts + styles.css.
- BLOCKERS: none.

## 2026-06-18 — 3b client + E2E + docs
- DONE: WBS-6 types.ts (Job/JobInput/Segment/RecordingState ifaces mirroring store, by hand). web/src/controls.ts (vanilla TS, no framework) — active-job <select> (+ "+ New job…" inline D8 form), Record/Stop button with open-segment elapsed timer (server-clock-skew calibrated), state line; REST client (getJSON/sendJSON); polls /api/recording/state every 3s + refreshes after actions; 409 on job-switch-while-recording surfaces "Stop recording before switching jobs"; 400 on start-no-job surfaces "Select a job before recording". readout.ts adds a controls-host element between header and content + controlsHost() accessor (readout owns layout, controls owns strip; recording never gates readout — axiom #1). main.ts wires `new Controls(readout.controlsHost())`. styles.css: .controls / .record-btn(.recording pulse) / .newjob-form etc. tsc strict + vite build clean (fixed 2 strict errors: hidden===true, JobInput->Record cast via unknown). make build clean.
- DONE: WBS-7 E2E vs captures/...-pressure.bin -format intellisense -replay-loop -addr :8090. AXIOM #1 PROVEN: samples rows climbed 4537->4953 STOPPED, 10881->11284 RECORDING, 11284->11609 after STOP. Job flow: POST job -> id 1, PUT active -> isActive true, GET active confirms. Recording: state not-recording -> start (open seg, stopped null, startedAtUs 1.78e15 = unix-micros) -> state recording+ids -> stop (stopped set) -> segments list -> PUT adjust started -30s -> re-GET confirms persisted. Errors: start-no-active-job 400 (fresh dir), double-start 409 returning open segment, set-active-different-job-while-recording 409, stop-not-recording 409. WS: profile frame + reading frames sane (vol.stage 2.1, agg.rate 1.28). SPA serves embedded build w/ controls. DB check: all 5 tables present; 2 segments persisted (adjusted start kept). Throwaway _wscheck/_dbcheck in /tmp (git stayed clean).
- DONE: AXIOM #5 PROVEN: grep of 3b diff (store jobs/recording, api/jobs, controls) shows only COMMENTS about not touching stage — zero stage-reset/zeroing logic, zero writes to samples/vol.stage. Recording methods only INSERT/UPDATE recording_segments.
- DONE: WBS-8 docs — data-model.md "Realized contract — Phase 3b" (jobs + recording_segments schema, marker-only/axiom-#1 note, segment timeline = samples.ts_us scale, full /api/jobs* + /api/recording/* contract, client strip). README Status += Phase 3b complete.
- DONE: full gate — gofmt -l empty, go vet ./... clean, go build ./... ok, go test ./... green (api+store new + daqformat/parser existing), make build clean.
- STATUS: complete. git status clean in worktree.
