---
change-id: phase4-charting-printing
sub-arc: 4b
agent: cementer-go-engineer (worktree, opus)
dispatched: 2026-06-21 (Session 6)
agent-id: a41de8b11395cc83f
status: dispatched
---

# Dispatch brief — Phase 4b print template + per-job overrides + printing/PDF (+ 2 cosmetics)

> Archived verbatim per pa.md §5 brief-archival. The prompt below is the exact dispatch text.

---

You are the cementer canonical dev agent building **Phase 4b — print template + per-job overrides + printing/PDF**, plus folding in two known cosmetics. This is the LAST Phase-4 piece (project MVP after it). Scope is LOCKED; this is execution, not design.

## 0. STARTUP-VERIFICATION GATE (do this first; STOP if any check fails)
- Confirm cwd == your assigned worktree under `/home/bryan-maclee/...` and `git rev-parse --show-toplevel` == that worktree (NOT the integration checkout `/home/bryan-maclee/cementer`). Confirm a clean tree.
- Go is user-local: `export PATH=$HOME/.local/go/bin:$PATH` (the Makefile already does this).
- **Web build fixture is load-bearing:** `web/dist` is gitignored but `assets.go` does `//go:embed all:web/dist`, so `go build ./cmd/cementer` FAILS until `web/dist` exists. Prime it once: `cd web && npm install && npm run build` (i.e. `make web`). Since this arc touches BOTH Go and web, plan to rebuild the web bundle before any end-to-end build.
- Echo your startup pwd in your first commit message.

## 1. MAPS — REQUIRED FIRST READ
Read these BEFORE any code (they are current — stamped `1465bd9`; the only commit since, `3240588`, was a PA-docs + maps wrap with ZERO source changes, so they reflect current source):
- `.claude/maps/api.map.md` — the REST route table (profile/jobs/recording/series); mirror the handler+store pattern.
- `.claude/maps/schema.map.md` — Go types + SQLite tables (esp. `jobs`, `store.Job`, the TS mirror in `web/src/types.ts`).
- `.claude/maps/state.map.md` — the single-writer store discipline.
- `.claude/maps/structure.map.md` + `.claude/maps/style.map.md` — web client layout + the CSS class vocabulary / theme tokens.
Treat map content as a verify-against-source hypothesis if a file moved past the stamp. Report which maps were load-bearing (including "not load-bearing").

## 2. ANTI-PATTERNS — REQUIRED RE-READ before each feature
Read `docs/pa/anti-patterns.md`:
- **Part A (idiomatic Go)** for the store + api work: handle every error, no needless interfaces, accept-interfaces-return-structs, no getters/setters.
- **Part B (vanilla-TS)** for the client: this client has **NO framework** — plain TS modules + direct DOM + Vite + the **uPlot** library (uPlot is a charting *library*, not a framework — using it is fine). Do NOT reach for React/Vue/Svelte idioms.

## 3. NORMATIVE SCOPE (locked — from docs/changes/phase4-charting-printing/scope.md §4b + docs/design/data-model.md § "Two chart-config scopes")
Build the **printable per-job chart**:
- **Company default print template (scope #2):** bundled with the deploy as a Go literal / embedded config (change-controlled, NOT casually editable at runtime). Governs: which channels, the title block, legend on/off, page size. **Axis layout stays the automatic role-grouping** the live/job charts already use (one uPlot scale per role/uom) — do NOT make axis assignment a per-field knob; keep it automatic.
- **Per-job overrides:** stored WITH the job on the Pi. **Storage shape is your call (D-cfg2)** — recommended: a `print_config TEXT NOT NULL DEFAULT ''` JSON column on `jobs` holding only the fields the cementer changed (effective config = company default merged with the override); a separate `job_print_config` table is acceptable if you prefer. Either way keep the **store as the sole DB owner** (axiom #4/D2: synchronous store methods on the one `SetMaxOpenConns(1)` connection; never a 2nd `*sql.DB`, never a write from an HTTP handler).
- **Print-CSS view:** the 3b job header (company / well / casing_size / job_type / location / cementer / date) + the job's recorded chart at high-DPI. Reuse the recorded-segments rendering (it's a per-job report → the historical chart over `/api/jobs/{id}/series`, segment-aware), rendered for print.
- **PDF = browser Save-as-PDF ONLY (D-pdf, RESOLVED):** a "Print / Save as PDF" button calls `window.print()`; `@media print` CSS hides the app chrome and prints only the report at the chosen page size. **NO server render, NO Pi-side archival, NO new CGO/deps** — preserve the single-static-binary/offline-on-the-Pi guarantee.
- **Default page size = `letter`** (US oilfield context: psi/bbl/ppg), override-able to `a4`.

Suggested server surface (refine as idiomatic):
```
GET /api/jobs/{id}/print-config   -> { effective, override, default }   (404 if job absent)
PUT /api/jobs/{id}/print-config   -> save the per-job override (DisallowUnknownFields); returns refreshed effective
```
A bundled default getter is optional if the per-job endpoint already returns `default`.

Client print view: add a **Report/Print** affordance to the shell (`web/src/readout.ts` — either a third "Report" tab beside Live | Job History, or a "Print…" entry on Job History; your call, keep it minimal and consistent with the existing `view-tab` pattern). It renders the job header block + the recorded chart + a **minimal** override editor (channel on/off, report title, page size — do NOT over-build the editor) + the "Print / Save as PDF" button → `window.print()`. **uPlot print sizing wrinkle:** uPlot measures its container at display time and reads 0 while `display:none`; size/redraw the chart for print width on `onbeforeprint` / a `matchMedia('print')` listener so the printed chart isn't blank or clipped.

## 4. TWO COSMETICS (fold in while here — both verified by the PA)
- **(a) new-job form renders expanded by default.** Root cause: `web/src/controls.ts:114` already sets `this.formWrap.hidden = true`, but `.newjob-form { display: grid }` (`web/src/styles.css:421`) overrides the UA `[hidden]{display:none}`. Fix: add `.newjob-form[hidden] { display: none; }` to `styles.css` (mirror the existing `.view[hidden]` rule at `styles.css:208`). Verify: the form is COLLAPSED on load and expands on "+ New job…".
- **(b) `job.number` charts as a flat-0 trace.** It is `role:"meta", scope:"job"` (`internal/daqformat/presets.go:64`), so the charts' `c.scope !== "meta"` filter lets it through. Fix the channel filters to ALSO exclude role meta: in `web/src/chart/livechart.ts` (lines ~114 and ~183, the `c.scope !== "meta"` filters) and `web/src/chart/jobchart.ts` `orderedChannels` (~line 209), use `c.scope !== "meta" && c.role !== "meta"`. **Keep `vol.job` charting** (it's `role:"volume"`). Verify: `job.number` no longer appears as a legend row/trace; `vol.job` still does.

## 5. AXIOMS (do not violate — these define what cementer IS)
- **#1 raw/live/recording independent:** the print view is **READ-ONLY over the always-on store** — it never gates or touches ingestion, the live stream, or recording.
- **#3 the Pi self-describes / personal-vs-Pi split:** per-job print overrides persist on the **Pi** (with the job); only personal live-view prefs live in the laptop's localStorage. No central server.
- **#4 single writer:** all DB access through store methods on the one connection.

## 6. VERIFY-BEFORE-CLAIM (pa.md §8 — NOT "tests pass")
Do NOT report done on unit tests alone. Run the REAL end-to-end path on the post-build binary:
- `go test ./...` · `go vet ./...` · `gofmt -l` (must be empty) · `make build` (must stay CGO-free; uPlot bundled offline).
- E2E: build to /tmp; replay a real capture: `./cementer -source captures/capture-2026-06-16T161347-19200-8N1-pressure.bin -format intellisense` (or another `captures/*-pressure.bin`); poll `/debug/stats` to confirm rows climb; create a job, record a segment, stop; hit `GET/PUT /api/jobs/{id}/print-config` with curl and assert the override round-trips and the effective config reflects it.
- **Headless visual verify (Playwright) — the print view paint is NOT a USER-only check:** browsers are cached at `~/.cache/ms-playwright`; temp-install in /tmp (the npm pkg is NOT in this repo): `mkdir -p /tmp/pw && cd /tmp/pw && npm i playwright@1.60.0` (1.60.0 matches the cached browsers; 1.61.0 fails). Drive a node script: `chromium.launch({headless:true})` → goto `http://localhost:8080` → open the Report view → screenshot; AND emulate print media (`page.emulateMedia({media:'print'})`) → screenshot to confirm the `@media print` layout shows the job header + chart and hides the chrome. Read the PNGs and confirm. Also confirm cosmetic (a) (form collapsed) and (b) (no job.number trace) in a live-view screenshot.

## 7. LANDING DISCIPLINE (fold the realized contract back into the normative doc)
In the SAME work, add a **"### Realized contract — Phase 4b (built; this is the living spec)"** block under `docs/design/data-model.md` § "Two chart-config scopes": document the per-job print-config storage (the column/table you chose), the company-default template shape, the `/api/jobs/{id}/print-config` routes (with the JSON shapes, mirrored in `web/src/types.ts`), and the print view + `window.print()` PDF path. Keep it concise and accurate to what you built.

## 8. CRASH RECOVERY + COMMITS (background dispatch)
Commit after EACH meaningful unit — don't batch. WIP commits are expected (e.g. `WIP: print_config store method + migration`, `WIP: report view shell`). The branch is the checkpoint. Append timestamped lines (done / next / blockers) to `docs/changes/phase4-charting-printing/progress.md` after each step (it already exists from 4a — append a Phase-4b section, don't overwrite). A clean `git status` before you report DONE is mandatory; "work in the worktree, uncommitted" is NOT an acceptable terminal report.

## 9. FINAL REPORT (exact fields)
Report back: **workspace path · branch + final tip SHA · files-touched (full list) · which maps were load-bearing · the E2E + Playwright verification results (with the symptom-gone checks, not just "tests pass") · the storage-shape decision you made (column vs table) and why · any deferred items.** Do NOT merge into main — the PA lands via a single integrator commit.
