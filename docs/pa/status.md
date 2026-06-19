---
status: current
last-reviewed: 2026-06-19
---

# cementer — live status (the SoT)

The single live source of truth for **done · in-flight · left**. Frozen planning prose (README
*Status*, `data-model.md` *Build order*) is NOT truth — this is. Layering: normative design
(`docs/design/data-model.md` + README architecture) → **this doc** → changelog → hand-off.

_Verify "is it shipped" claims against code (grep / `go build` / the SQLite schema), not this doc's
prose — but keep this doc honest at every wrap. `data-model.md` now carries the **realized** Phase-2/3/4a
contracts (landing discipline, adopted S5)._

## Phase board

| # | Phase / step | State | Evidence (verified 2026-06-19, tip `1465bd9`) |
|---|---|---|---|
| P1 | Durable ingest → WS → dark-mode readout | ✅ **DONE** | `cmd/cementer/main.go` pipeline; `internal/store` SQLite WAL single-writer |
| 1 | Config-driven dynamic channels + theme + storage env | ✅ **DONE** | store keyed by `channel`; `web/src/theme.ts`; `-data-dir`/`$CEMENTER_DATA_DIR` |
| 2 | **Intellisense** `DaqFormat` preset + format engine (mapping + compute) | ✅ **DONE** (`83f036a`) | `internal/daqformat` generic engine + `Intellisense()`/`Synthetic()` presets + `IntellisenseChannels()`; `-format` flag; built from the **live wire** (14-col), not the CSV export |
| 3a | Pump Profile persistence + hello/profile message + scope-grouped display | ✅ **DONE** (`cd71beb`) | `pump_profiles`/`profile_channels`; per-conn WS profile frame; `GET/PUT /api/profile` + reset; scope-grouped client (enabled-only) |
| 3b | Job CRUD + recording segments + active-job | ✅ **DONE** (`cf46ab3`) | `jobs`/`recording_segments`; `/api/jobs*` + `/api/recording/*`; client controls; **axiom #1 proven** (recording is a marker, never gates ingest) |
| 4a | uPlot charting core (series API + live + historical) | ✅ **DONE** (`5c69e07` + `1465bd9`) | `store.Series`/`JobSeries` (min/max decimation); `GET /api/samples` + `/api/jobs/{id}/series`; uPlot live rolling chart (replaces readout) + job-history chart w/ segment shading; live-view config in localStorage. Time axis in **seconds** (fixed). **Playwright-verified render.** |
| 4b | Print template (company default + per-job overrides) + print-CSS + PDF | ⬜ **NOT STARTED** | scope locked: [`phase4-charting-printing/scope.md`](../changes/phase4-charting-printing/scope.md). PDF = browser Save-as-PDF only (D-pdf) |
| 3c | Retention/downsampling-as-code (DD rider #3) | ⬜ **DEFERRED** (by design) | low urgency at ~7 rows/s; design sketched in the phase3 scope |

## Decision records (locked)

- **Phase 2:** [`phase2-intellisense-daqformat/scope.md`](../changes/phase2-intellisense-daqformat/scope.md) (D1–D4) + the live-wire findings doc.
- **Phase 3:** [`phase3-jobs-recording-profiles/scope.md`](../changes/phase3-jobs-recording-profiles/scope.md) (D1–D10; D2 = store sole DB owner / single-conn CRUD; D4 auth deferred; D8 job fields).
- **Phase 4:** [`phase4-charting-printing/scope.md`](../changes/phase4-charting-printing/scope.md) (X=time; all-enabled role-grouped axes; live chart replaces readout; PDF = browser Save-as-PDF only).
- Dispatch briefs (6) archived under [`docs/pa/briefs/`](briefs/).

## Standing practices

- **Landing discipline (S5):** at each sub-arc landing, fold the realized contract (schema/WS/API) into
  `docs/design/data-model.md` so the normative doc stays the living spec — don't let deltas accumulate
  here. No separate as-built spec doc (decided sufficient).
- **Canonical dev agent:** `cementer-go-engineer` (worktree-isolated, `model: opus`) — used for every
  source arc this session. `general-purpose` is the generalist fallback only.
- **Headless verify:** Playwright browsers are cached; temp-install `playwright@1.60.0` to screenshot the
  web UI (see auto-memory). The chart paint is no longer a USER-only check.

## ✅ Bench-top stack validation — VERIFIED 2026-06-13 (Peter, on `CementSerial`)

Go+SQLite Pi stack proven on both serial-ingress paths (GPIO UART @115200: 2,812 rows; CP2102 USB
@115200: 4,404 rows). Transport was **simulated** (ESP32-replayed CSV); the **real-DAQ wire contract is
confirmed for Intellisense** (S4 capture). Field runbook lives in `hand-off.md`.

## Design ↔ code deltas (tracked TODOs)

_**Standing practice (adopted S5):** close each delta into `docs/design/data-model.md` at the sub-arc
landing that resolves it — fold the realized schema/WS/API contract into the normative doc so it stays
the living spec; don't let deltas accumulate here. No separate as-built spec doc (decided sufficient)._

- ~~`recording_segments`/`jobs` tables~~ — ✅ built (3b).
- ~~Pump Profile / DAQ Format / hello-profile WS message~~ — ✅ Profile + hello/profile built (3a); DaqFormat
  stays a code preset (in-UI format CRUD still deferred — Phase 5+).
- ~~Computed/derived channels~~ — ✅ engine has a sum/mean compute pass (no-op for Intellisense, which
  field-maps its aggregates).
- ~~Parser vs real format mismatch~~ — ✅ resolved: `internal/daqformat` is the format engine; `internal/parser`
  is now **off the main path** (kept only for its Phase-1 test). **Cleanup candidate:** delete parser or
  fold its cases into a daqformat test.
- `internal/api/` and `web/src/chart/` are now **populated** (3a/3b/4a) — no longer placeholders.
- **`job.number` charts as a flat trace** — its profile scope is `job` (role `meta`), so the live chart's
  `scope!=="meta"` filter doesn't exclude it. Harmless flat-0 line; minor follow-up.
- **`controls.ts` new-job form renders expanded by default** — cosmetic; fold into 4b.

## Doc-currency / hygiene debts

- ~~Stale `docs/plan` reference~~ — ✅ fixed in `main.go` pkg-doc + README (→ `data-model.md`).
- ~~README "Go 1.22+"~~ — ✅ now "Go 1.26+ / Node 20+".
- **Nav-maps regenerated at S5 wrap** (`.claude/maps/`, stamp = wrap HEAD) — were 5 phases stale.
- **⚠ Plaintext credentials committed** in `pi4b & test db/credetials&currentDB.README` (`ddf8ada`):
  test-rig SSH/Influx/Grafana logins. Rotate + gitignore if this repo is ever shared/public. Surfaced,
  not changed (collaborator's file).
- **No commit gate installed** (`core.hooksPath` unset). Baseline still recommended: `gofmt -l` + `go vet`
  + `go build` + `go test`; `make build` pre-push when `web/` changed.

## Near-term actions (not yet done)

1. **Phase 4b** — printing (company default template + per-job overrides + print-CSS view + browser
   Save-as-PDF). Scope locked. Fold in the two minor cosmetics above (new-job form, job.number trace).
2. **Install the commit gate** (above).
3. **Parser cleanup** (delete/fold the off-path `internal/parser`).
4. **Totco preset** — when a Totco unit is reachable (same direct-laptop capture method).

## Test surface

- `go test ./...`: `internal/daqformat`, `internal/parser`, `internal/store`, `internal/api` have tests;
  others report "no test files". Web has no unit suite (tsc-strict + Playwright screenshot are the checks).
- **Last full run (2026-06-19 wrap):** `go test ./...` ✅ · `go vet ./...` ✅ · `gofmt -l` clean ·
  `make build` ✅ (CGO-free, uPlot bundled offline).
