# Hand-off — live

`as of: Session 5 close · 2026-06-19` (machine: Linux garage desktop — `/home/bryan-maclee/cementer`)

> Optimize for the NEXT session's pickup. Session 5 shipped **Phases 2, 3a, 3b, 4a** + collaborator
> quickstart + chart fix-ups — six commits, all landed + pushed. The project now has the full pipeline:
> serial/replay → daqformat engine → SQLite → WS (readings + hello/profile) → self-describing
> scope-grouped client → jobs + recording segments → **uPlot live + historical charts**. Prior hand-offs:
> S4 → `archive/hand-off-2026-06-16.md`; S2 → `archive/hand-off-2026-06-13.md`; S1 → `archive/hand-off-2026-06-12.md`.
>
> **Git: all pushed.** `origin/main = local = 1465bd9`, tree clean, no worktrees. (Push works over SSH on
> this Linux box; the GitHub-Desktop/credential-manager hang was a Windows-field-laptop-only issue.)

## ✅ Session 5 result — Phases 2 → 4a complete (commits, newest first)

| Commit | Arc | One-liner |
|---|---|---|
| `1465bd9` | 4a fix-ups | uPlot time axis ms→**seconds** (labels were 1000× off); varied `testdata/intellisense-demo.txt` for `make demo` |
| `1f65c13` | quickstart | `make demo` + fixed `make run` (`-format synthetic`); README "Quick demo"; Go-1.26 currency |
| `5c69e07` | **Phase 4a** | charting core — series API + uPlot live (replaces readout) + job-history chart w/ segment shading |
| `cf46ab3` | **Phase 3b** | jobs + recording_segments + `/api/jobs*` + `/api/recording/*` + client controls |
| `cd71beb` | **Phase 3a** | self-describing pump: profiles + hello/profile WS frame + `/api/profile` + scope-grouped client |
| `83f036a` | **Phase 2** | `internal/daqformat` engine + Intellisense preset (from the live wire) + `-format` flag |

All built via `cementer-go-engineer` (worktree, opus); each PA-verified by independent E2E at landing.

## ▶ Next priority

1. **Phase 4b — printing** (the last Phase-4 piece; project MVP after this). Scope is LOCKED:
   [`docs/changes/phase4-charting-printing/scope.md`](../changes/phase4-charting-printing/scope.md) §4b.
   - Company default print template (bundled, change-controlled) + **per-job overrides** stored on the Pi
     (a `job_print_config` table or JSON column — engineer's call).
   - Print-CSS view = the 3b job header (company/well/casing/job_type/location/cementer/date) + the chart
     at high-DPI. **PDF = browser Save-as-PDF only** (D-pdf RESOLVED S5 — no server render, no Pi archival).
   - Dispatch via `cementer-go-engineer` (worktree, opus); anti-patterns Part B (vanilla-TS, uPlot).
   - **Fold in two known cosmetics while there:** (a) `web/src/controls.ts` new-job form renders expanded
     by default (should be collapsed until "+ New job"); (b) `job.number` shows as a flat-0 chart trace
     (its profile scope is `job` not `meta`, so the live chart's `scope!=="meta"` filter misses it).
2. **Parser cleanup** — `internal/parser` is off the main path (daqformat replaced it); delete it or fold
   its test cases into a daqformat test. Not urgent.
3. **Commit gate** — still not installed (`core.hooksPath` unset). Baseline: gofmt+vet+build+test;
   `make build` pre-push when `web/` changed.
4. **Totco preset** — still TODO; unit not accessible. Same direct-laptop capture method (S4 runbook below).

## 🔍 How to run / demo / verify (reuse this)

- **Demo (no pump):** `make demo` → replays `testdata/intellisense-demo.txt` (ten real captures
  concatenated, varied multi-phase) at 200 ms → open `http://localhost:8080`. `make run` = synthetic
  stream (`-format synthetic`). cementer is **silent on stdout when healthy** — watch the browser / the
  raw log / `/debug/stats`, not the console.
- **Headless visual verify (Playwright):** browsers are cached at `~/.cache/ms-playwright`; the npm pkg
  is NOT in cementer (it's in `~/scrmlMaster/scrml`, a different repo — don't reach across). Temp-install:
  `mkdir -p /tmp/pw && cd /tmp/pw && npm i playwright@1.60.0` (1.60.0 matches the cached 1223 browsers;
  1.61.0 wants 1228 → fails). Drive a node script: `chromium.launch({headless:true})` → goto :8080 →
  waitForTimeout → screenshot; Read the PNG. (A working `shot.js` was used this session.) Saved to memory.
- **PA E2E pattern used every arc:** build to /tmp; replay `captures/...-pressure.bin -format intellisense`;
  curl the new endpoints + assert via `/api/*` or a throwaway `_`-prefixed Go DB helper (sqlite3 NOT
  installed) removed after; prove axiom #1 by polling `/debug/stats` rows-climb while hitting the new path.

## ⚙ Architecture as-built (for fast orientation)

- Pipeline: `source` (serial/replay) → `rawlog` (L1) → **`daqformat` engine** (Apply: tokenize → field-
  count guard → field-map + transform → compute pass → server-stamp TS) → `store` (SQLite WAL, **single
  writer connection, sole DB owner** — axiom #4/D2) → `onCommit` → `hub` → WS.
- WS `/ws/live`: sends ONE `{type:"profile"}` greeting per connection (enabled channels), then
  `{type:"reading"}` frames. Profile/job/recording are **read/CRUD via `internal/api` HTTP**, store
  methods only (no handler-side DB).
- Client (`web/src/`, vanilla TS, NO framework): `readout.ts` = shell (header + Live|Job tabs + window
  select + controls host + footer); `chart/livechart.ts` (rolling, role-grouped, legend = live values),
  `chart/jobchart.ts` (segments + shading), `controls.ts` (job select + Record), `ws.ts`/`types.ts`.
  uPlot bundled offline (no CDN). **uPlot x-axis is in SECONDS** (the fix this session).
- Axioms held + proven: #1 (recording/chart never gate ingest — proven by rows-climb), #2 (format =
  config; a new format = a new `DaqFormat` value), #3 (Pi self-describes via the profile), #4 (single
  writer connection), #5 (no stage-reset on record).

## Totco (deferred — unit not accessible since S3/S4)

Totco config screen: **COM6 · 9600 8N1 · Protocol 1 · 250 ms**. S3 hit total silence on COM6 at every
baud → physical/electrical. Resume: (1) `-Loopback` self-test (jumper DB9 2↔3); (2) confirm a null-modem
cable; (3) confirm the Totco is actually streaming. Then capture as for Intellisense and define the Totco
preset (a new `DaqFormat` value — no engine change).

## State as of close (Session 5)

| Item | State |
|---|---|
| Phases 1, 2, 3a, 3b, 4a | ✅ DONE + pushed (`1465bd9`) |
| Phase 4b (printing) | ⬜ scoped, not started |
| Phase 3c (retention) | ⬜ deferred by design |
| Source build | `go test`/`vet`/`gofmt`/`make build` all green (CGO-free, uPlot offline) |
| Git | clean, `origin = local = 1465bd9`, no worktrees |
| Nav-maps | regenerated at S5 wrap (were stale `ee446c3`) |
| Canonical dev agent | `cementer-go-engineer` (active) |

## Parked debts (non-blocking)

- No commit gate (`core.hooksPath` unset).
- `internal/parser` off-path (cleanup candidate).
- `controls.ts` new-job form expanded-by-default; `job.number` flat chart trace (both → fold into 4b).
- Plaintext test-rig creds committed in `pi4b & test db/...README` (rotate if repo ever shared).
- In-UI DaqFormat editing + live WS profile/recording-state push = deferred niceties (Phase 5+).
