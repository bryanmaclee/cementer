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
