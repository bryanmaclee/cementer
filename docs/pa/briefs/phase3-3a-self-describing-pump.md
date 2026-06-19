---
change-id: phase3-jobs-recording-profiles
slug: 3a-self-describing-pump
agent: cementer-go-engineer
model: opus
isolation: worktree
agent-id: a4aabe8fced032743
dispatched: 2026-06-18 (Session 5)
dispatched-from-tip: 83f036a
status: dispatched
---

# Dispatch brief (verbatim) — Phase 3a: self-describing pump backbone

> Archived per pa.md §5. Exact prompt sent to `cementer-go-engineer` (background, worktree, opus).
> Paths/SHAs as-at-dispatch.

---

You are the canonical cementer Go dev-agent. Build **Phase 3a — the self-describing pump backbone**: PumpProfile persistence + the hello/profile WS message + scope-grouped client display. This touches BOTH the Go server and the vanilla-TS client. Make the Pi tell the client what pump it is, so the client renders only the channels this rig actually has, grouped by scope.

═══════════════════════════════════════════════════════════════════════
0) STARTUP-VERIFICATION GATE (FIRST, before any write)
═══════════════════════════════════════════════════════════════════════
- Confirm cwd == your assigned worktree under `/home/bryan-maclee/...` and `git rev-parse --show-toplevel` == that worktree (MUST NOT equal the integration root `/home/bryan-maclee/cementer`). Confirm clean tree. If any check fails, STOP and report.
- Go is user-local: `export PATH=$HOME/.local/go/bin:$PATH`. Verify `go version` (1.26.x).
- `web/dist` is gitignored + `go:embed`-required (assets.go), AND this arc changes the web client, so you WILL build web: `cd web && npm install && npm run build` (Node 22 + Vite 8 work here). For server-only intermediate steps, `go build ./internal/...` is fine; the binary build + E2E need `make build`.

═══════════════════════════════════════════════════════════════════════
1) MAPS — REQUIRED FIRST READ (with a currency caveat)
═══════════════════════════════════════════════════════════════════════
Read `.claude/maps/`: `structure`, `state`, `schema`, `api`, `events`. **Currency caveat:** the maps are stamped `ee446c3` (pre-Phase-2). Since then ONE source change landed (commit `83f036a`): the new `internal/daqformat` package + the `main.go` parser→daqformat wire-in + the `-format` flag. Treat the maps as accurate for store/hub/model/web/main STRUCTURE, but factor in daqformat and verify against source where it matters. Report whether the maps were load-bearing.

═══════════════════════════════════════════════════════════════════════
2) ANTI-PATTERNS — READ BEFORE CODE (BOTH parts — this is the first web change)
═══════════════════════════════════════════════════════════════════════
Read `docs/pa/anti-patterns.md` — **Part A (Go)** for store/api/main AND **Part B (vanilla-TS)** for web/src. Part B is mandatory and easy to violate: the client is **NO framework, no JSX, no virtual DOM, no state library** — plain TS modules + Vite + direct DOM. Build the UI from the profile the Pi sends, not a hard-coded set. localStorage is for personal prefs only. Keep the WS envelope type in `types.ts` in sync with Go BY HAND (no codegen).

═══════════════════════════════════════════════════════════════════════
3) NORMATIVE SOURCES
═══════════════════════════════════════════════════════════════════════
- Design intent: `docs/design/data-model.md` (Pump Profile / Channel / scope model; raw-live-recording independence; the hello/profile message).
- **AUTHORITATIVE BUILD SPEC for this arc:** `docs/changes/phase3-jobs-recording-profiles/scope.md` — read the "Data model (Pump Profile)", "hello/profile WS message", "API surface", "Client", and "Sub-arc 3a" sections. Decisions **D2 and D5 are RESOLVED** there; honor them. (Jobs/recording = 3b, NOT this arc. Auth = deferred. DaqFormat CRUD = deferred; format stays the code preset.)

Project axioms (pa.md §"Project axioms") you MUST honor:
- **#4 single-writer durability (the sharp one for this arc — D2 RESOLVED):** the **store is the single DB owner.** New CRUD = synchronous store methods on the SAME `SetMaxOpenConns(1)` connection (database/sql serializes them against the sample `writeLoop`; WAL + busy_timeout handle contention). **NO second `*sql.DB`, NO `sql.Open` anywhere else, NO DB access from HTTP handlers** — handlers call store methods only. Samples keep their async batch goroutine untouched.
- **#3 self-describing island:** the pump's identity lives on the Pi (the profile); the client is a thin renderer of what the Pi sends.
- **#1 raw/live/recording independent:** the profile/display path NEVER gates ingestion or the live readout.
- **#2 format = config:** don't scatter format assumptions; the profile is seeded from the format's channel vocab but the store stays format-agnostic (see seed note below).

═══════════════════════════════════════════════════════════════════════
4) CURRENT CODE (verified 2026-06-18)
═══════════════════════════════════════════════════════════════════════
- `internal/store/store.go` — one `samples` table; opens with `_pragma=foreign_keys(ON)`; `SetMaxOpenConns(1)`; sample writes via `writeLoop` goroutine (batch+WAL). Add the new table + CRUD HERE on the same connection. `initSchema` is where DDL goes.
- `internal/api/` — EMPTY. Build the HTTP/JSON layer here.
- `cmd/cementer/main.go` — `wsEnvelope{Type, Reading}` is the only WS message; `serveWS(h *hub.Hub)` upgrades + registers a `hub.Subscriber` + starts read/write pumps; routes mounted on a `http.ServeMux` with Go-1.22 pattern routing (`mux.HandleFunc("GET /ws/live", ...)`). main already imports `store` and `daqformat`.
- `internal/hub` — fan-out of committed readings; transport-agnostic. The profile frame is a per-connection greeting — send it directly to the new conn in `serveWS`; do NOT route it through `hub.Broadcast`.
- `web/src/readout.ts` — renders FLAT cards; `describeChannel()`/`ROLE_INFO` INFER label/uom/decimals from the id (its comment says this is the stopgap "until the pump profile arrives"). REPLACE that with the profile; group by scope.
- `web/src/ws.ts` — handles only `type==="reading"`. Add `type==="profile"`.
- `web/src/types.ts` — `WSEnvelope{type, reading?}`. Add `profile?` + `Profile`/`Channel` ifaces.
- `internal/daqformat/presets.go` — `IntellisenseChannels()` returns the 13-channel vocab (id/role/scope/unitIndex/uom/label/decimals); `SyntheticChannels()` the 4-channel one. **These seed the default profile.**

═══════════════════════════════════════════════════════════════════════
5) TARGET DESIGN — Phase 3a
═══════════════════════════════════════════════════════════════════════
**Schema (in store `initSchema`)** — exactly the DDL in scope.md:
- `pump_profiles(id, name, units, daq_format_id, is_active, created_at_us, updated_at_us)`
- `profile_channels(id, profile_id FK ON DELETE CASCADE, channel_id, role, scope, unit_index, label, uom, decimals, enabled, sort_order, UNIQUE(profile_id, channel_id))`

**Store methods (single-conn, store is sole DB owner):**
- `ActiveProfile() (Profile, bool, error)` — the active profile + its channels (ordered by sort_order).
- profile CRUD: create a profile (+ channels), update units, update a channel's `enabled/label/uom/decimals/sort_order`, set-active, and a "reset channels from a supplied vocab" op.
- Keep the sample path untouched; these are ordinary methods on `*Store` using `s.db` (the one connection). They are infrequent — synchronous is correct.

**Profile wire type** (define where it's cleanest — recommend `internal/store` owns `Profile`/`Channel` structs WITH json tags so api + ws serialize them directly; engineer's discretion, but keep `internal/model` from depending on store). JSON shape (this is the contract `types.ts` mirrors):
```
Profile { name string, units int, formatId string, channels []Channel }   // channels = ENABLED only, in sort_order
Channel { id, role, scope string, unitIndex int, label, uom string, decimals int }
```

**Seed on first run (keep the store format-agnostic — main wires it):** after `store.Open`, if there is no active profile, `main.go` creates one from the active format's vocab — `daqformat.IntellisenseChannels()` for `-format intellisense`, `SyntheticChannels()` for `synthetic` — all channels `enabled=1`, `daq_format_id` = the format id, `is_active=1`, `units` = 1 (intellisense) / 1 (synthetic). The store exposes generic create-profile/seed methods; **main** supplies the daqformat vocab (so `internal/store` does NOT import `internal/daqformat`). Seed must be **idempotent** (a second boot doesn't duplicate).

**hello/profile WS message:** extend `wsEnvelope` with `Profile *<wiretype> json:"profile,omitempty"`. `serveWS` must be able to fetch the active profile (pass the store, or a `func() (Profile,bool)`, into `serveWS`). On connect, AFTER upgrade and BEFORE/at registration, write ONE `{type:"profile", profile:{…}}` text frame directly to that conn; then run the normal live-reading write pump. If no active profile (shouldn't happen post-seed), send nothing and log.

**API (`internal/api`, mounted on the mux):** a constructor taking `*store.Store`, registering Go-1.22 pattern routes; handlers call store methods only (no `*sql.DB`):
- `GET /api/profile` -> active profile incl. ALL channels with their `enabled` flag (the editor needs to see disabled ones too — note: the GET for the editor returns all channels; the WS profile frame sends ENABLED only).
- `PUT /api/profile` -> update units + per-channel enabled/label/uom/decimals/sort_order (accept a JSON body; validate).
- `POST /api/profile/reset` -> reseed channels from the format vocab (escape hatch). main injects the vocab provider.
Return clear JSON errors + proper status codes. `%w`-wrap server-side errors.

**Client (vanilla TS — Part B):**
- `types.ts`: add `Profile`/`Channel` ifaces + `profile?` on `WSEnvelope`.
- `ws.ts`: on `env.type==="profile" && env.profile` call a new `onProfile` handler (add a param); keep reading handling.
- `readout.ts`: on profile, build **scope groups** — a section per `Unit 1`, `Unit 2`, … (by unitIndex), then `Aggregate`, `Stage`, `Job` (meta hidden). Each group renders cards for its channels using the profile's `label`/`uom`/`decimals`. DELETE the `describeChannel`/ROLE_INFO/PART_LABEL inference; keep only a minimal defensive fallback for a streamed channel id absent from the profile (render it ungrouped, last). A channel that isn't in the profile's ENABLED set never gets a card. Preserve the existing status/seq/stale footer + theme toggle.
- `main.ts`: wire the `onProfile` handler to the Readout.
- Style: extend `styles.css` for group sections/headers; stay consistent with the existing dark/light theme + grid; no framework, no CDN.

═══════════════════════════════════════════════════════════════════════
6) WORK BREAKDOWN — one WIP-commit per unit; progress.md after each
═══════════════════════════════════════════════════════════════════════
Append-only `docs/changes/phase3-jobs-recording-profiles/progress.md` (timestamped done/next/blockers).
1. Schema DDL in store `initSchema` (pump_profiles + profile_channels).
2. Store profile types + CRUD methods (single-conn) + `ActiveProfile` + reset; unit tests (round-trip, set-active uniqueness, channel update, **seed idempotency**).
3. Profile wire serialization (enabled-only, sort_order); test.
4. `main.go`: idempotent seed from the daqformat vocab on first run; pass the store/profile-provider into `serveWS`; send the per-connection profile frame; mount the api routes.
5. `internal/api`: GET/PUT /api/profile + POST /api/profile/reset; handler tests (httptest) hitting store methods.
6. Client: types.ts + ws.ts profile handling.
7. Client: readout.ts scope-grouped render (drop inference) + main.ts wiring + styles.css.
8. E2E verification (next section).
9. **Docs (landing discipline — adopted S5, do NOT skip):** fold the REALIZED contract into `docs/design/data-model.md` — add the concrete `pump_profiles`/`profile_channels` schema + the hello/profile message JSON shape + the `/api/profile*` routes into the normative doc (so it stays the living spec). Update README if the layout/flags changed.

═══════════════════════════════════════════════════════════════════════
7) VERIFY-BEFORE-CLAIM (do NOT mark done on unit tests alone)
═══════════════════════════════════════════════════════════════════════
`make build`, then run against a real capture:
`./cementer -source captures/capture-2026-06-16T161347-19200-8N1-pressure.bin -format intellisense -replay-interval 20ms -replay-loop=false -data-dir /tmp/ce-3a -addr :8090`
- `curl -s localhost:8090/api/profile` -> active profile with the 13 seeded Intellisense channels (each with enabled flag).
- `curl -X PUT .../api/profile` to set `unit2.pressure`, `unit2.rate`, `water.rate`, `density.2` `enabled=false`; GET again confirms.
- Inspect the WS greeting: the FIRST frame on `/ws/live` must be `{type:"profile",...}` listing ENABLED channels only (after the PUT, those 4 are absent). (Use a tiny WS check — e.g. a short `go run` websocket client, or `websocat` if present; if neither, write+remove a throwaway `_`-prefixed Go ws client so git stays clean.)
- Seed idempotency: stop, restart against the SAME `-data-dir`; `GET /api/profile` shows ONE profile (no duplicate), and your earlier enabled/label edits persisted.
- Re-confirm **axiom #1**: `/debug/stats` row count climbs regardless of any profile call (profile/display never gates ingestion).
- Load `http://localhost:8090/` and confirm scope-grouped cards render with real labels/units (eyeball; note it in the report).
- Static gate: `gofmt -l` (empty), `go vet ./...`, `go test ./...` all green; `make build` clean.

═══════════════════════════════════════════════════════════════════════
8) INVARIANTS (all must hold)
═══════════════════════════════════════════════════════════════════════
- Store is the SOLE DB owner; ONE `*sql.DB`/connection; zero handler-side DB access; sample `writeLoop` untouched (axiom #4 / D2).
- Profile/display path never gates ingestion or live readout (axiom #1).
- `internal/store` does NOT import `internal/daqformat` (main wires the seed vocab).
- Client is framework-free; renders from the profile; enabled-only.
- Seed is idempotent. gofmt/vet/test green; CGO-free static build.
- Phase-1/2 paths still work (synthetic + intellisense replay).

═══════════════════════════════════════════════════════════════════════
9) REPORT BACK (final message = data for the PA)
═══════════════════════════════════════════════════════════════════════
(a) worktree path + branch + final tip SHA; (b) files-touched; (c) exact verify commands + output (the /api/profile JSON, the WS first-frame proof, seed-idempotency proof, /debug/stats-climbs-while-profile-called proof, gofmt/vet/test); (d) maps load-bearing?; (e) deferred items/follow-ups; (f) anything that contradicted this brief or the scope/design docs; (g) confirm `data-model.md` was updated with the realized contract. Ensure `git status` is clean in the worktree before reporting DONE (commit everything).
