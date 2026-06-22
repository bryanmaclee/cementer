---
status: current
last-reviewed: 2026-06-21
---

# cementer — live status (the SoT)

The single live source of truth for **done · in-flight · left**. Frozen planning prose (README
*Status*, `data-model.md` *Build order*) is NOT truth — this is. Layering: normative design
(`docs/design/data-model.md` + README architecture) → **this doc** → changelog → hand-off.

_Verify "is it shipped" claims against code (grep / `go build` / the SQLite schema), not this doc's
prose — but keep this doc honest at every wrap. `data-model.md` now carries the **realized** Phase-2/3/4a
contracts (landing discipline, adopted S5)._

## Operator in-flight (per-operator · section-owned — each operator edits ONLY their own block)

cementer is run by **two co-equal operators** since S6. The **phase board** below + `changelog.md`
are the shared truth; this section is each operator's current focus. Live cross-operator
coordination (claims, push intents, notices) is on the **coord branch** — `make coord` →
`.coord/` (the at-a-glance "who's doing what"). Layout: [`README.md`](README.md).

### Bryan
- **Phase 4b** (printable per-job report) **LANDED** `93011e6` (unpushed). Multi-party groundwork:
  commit gate ✅, DD ✅, coord branch ✅, meta-doc partition (in progress). Holding 4+ unpushed
  commits on `main` pending the PR-flow cutover (Peter access + branch protection).

### Peter
- **P1 (2026-06-21, Windows field laptop):** adopted the S6 multi-party model (PR-flow + coord +
  meta-doc partition); landed P1 onboarding docs via **PR #2 → `main` `0a96095`**. Stood up the Windows
  toolchain (Go 1.26.4 + Node 24.17.0), installed the commit gate, fixed a Windows CRLF/gofmt break
  (`autocrlf`), and **PA-verified Phase 4b end-to-end** (built/ran/recorded → report + print render via
  Edge headless). Filed ruleset issue **#3** to Bryan (exempt `coord`; allow feature-branch deletion).
  - _Ruleset blocks RESOLVED same-day (issue #3 closed, verified):_ Bryan scoped the rules → coord
    push-direct + merged-branch deletion both work; coord pushed/synced (`d1028bc`); P1 wrap landed (PR #4).
  - _Still open (not claimed):_ `.gitattributes` durable CRLF fix; parser cleanup. (`pa.md` topology
    rewrite ✅ done by Bryan, PR #6 `42ef5f2`.)
- **P2 (2026-06-21):** opened the **`serial-split-tap`** hardware arc — designed an isolated, listen-only
  serial tap (6N137 opto → Pi GPIO UART) so the Pi can ingest a live DAQ stream **without disturbing the
  existing consumer**. **Scope doc landed on `main` (PR #7, `1b942eb`)**; build **PAUSED** pending operator
  measurement #1 (DAQ TXD idle voltage). Spec: [`serial-split-tap/scope.md`](../changes/serial-split-tap/scope.md).

## Phase board

| # | Phase / step | State | Evidence (verified 2026-06-21 P1, tips `1465bd9` / `0a96095`) |
|---|---|---|---|
| P1 | Durable ingest → WS → dark-mode readout | ✅ **DONE** | `cmd/cementer/main.go` pipeline; `internal/store` SQLite WAL single-writer |
| 1 | Config-driven dynamic channels + theme + storage env | ✅ **DONE** | store keyed by `channel`; `web/src/theme.ts`; `-data-dir`/`$CEMENTER_DATA_DIR` |
| 2 | **Intellisense** `DaqFormat` preset + format engine (mapping + compute) | ✅ **DONE** (`83f036a`) | `internal/daqformat` generic engine + `Intellisense()`/`Synthetic()` presets + `IntellisenseChannels()`; `-format` flag; built from the **live wire** (14-col), not the CSV export |
| 3a | Pump Profile persistence + hello/profile message + scope-grouped display | ✅ **DONE** (`cd71beb`) | `pump_profiles`/`profile_channels`; per-conn WS profile frame; `GET/PUT /api/profile` + reset; scope-grouped client (enabled-only) |
| 3b | Job CRUD + recording segments + active-job | ✅ **DONE** (`cf46ab3`) | `jobs`/`recording_segments`; `/api/jobs*` + `/api/recording/*`; client controls; **axiom #1 proven** (recording is a marker, never gates ingest) |
| 4a | uPlot charting core (series API + live + historical) | ✅ **DONE** (`5c69e07` + `1465bd9`) | `store.Series`/`JobSeries` (min/max decimation); `GET /api/samples` + `/api/jobs/{id}/series`; uPlot live rolling chart (replaces readout) + job-history chart w/ segment shading; live-view config in localStorage. Time axis in **seconds** (fixed). **Playwright-verified render.** |
| 4b | Print template (company default + per-job overrides) + print-CSS + PDF | ✅ **DONE** (`93011e6`, merged PR #1 `c952c54`) | `internal/printcfg` (company default + per-job override) + `GET/PUT /api/jobs/{id}/print-config`; `web/src/report.ts` **Report tab** (job header + segment-shaded chart + Save-as-PDF via `@media print`). **PA-verified E2E render (P1, Windows/Edge headless).** PDF = browser Save-as-PDF only (D-pdf) |
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
- **Headless verify:** temp-install `playwright@1.60.0` and drive a browser to screenshot the web UI; the
  chart/report paint is no longer a USER-only check. Linux: cached browsers. **Windows (P1): drive system
  Edge (`chromium.launch({channel:'msedge'})`, `PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1`) — no browser download.**

## ✅ Bench-top stack validation — VERIFIED 2026-06-13 (Peter, on `CementSerial`)

Go+SQLite Pi stack proven on both serial-ingress paths (GPIO UART @115200: 2,812 rows; CP2102 USB
@115200: 4,404 rows). Transport was **simulated** (ESP32-replayed CSV); the **real-DAQ wire contract is
confirmed for Intellisense** (S4 capture). Field runbook lives in `hand-off-bryan.md`.

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
  was **removed in B6** (`bryan/cleanup`) — dead code (nothing imported it; `daqformat` has its own coverage).
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
- ~~No commit gate installed~~ — ✅ **installed S6** (source-controlled `scripts/git-hooks/`,
  `core.hooksPath=scripts/git-hooks`; install per-clone via `make hooks`). pre-commit = gofmt+vet+build+test;
  pre-push = `go test ./internal/...` (or `make build` when `web/` changed).
- **⚠ No `.gitattributes` → Windows CRLF break** (found P1). Git-for-Windows `autocrlf=true` checks the tree
  out as CRLF; `gofmt` is LF-only, so the pre-commit gate rejects every Go change on Windows. Mitigated this
  clone (`autocrlf=false` + renormalized). **Durable fix = add `.gitattributes` (`* text=auto eol=lf`)** —
  not yet done (a `peter/<arc>` PR; coordinate with Bryan).
- ~~Repo ruleset too broad~~ — ✅ **RESOLVED + verified** (issue **#3** closed): Bryan scoped the
  require-PR + restrict-deletions rules; `coord` push-direct and feature-branch deletion both work now.
- Minor (still open): pre-push runs the gate on a branch *deletion* (no Go in range) → fails on a no-Go
  machine; minor hook refinement for Bryan (skip delete / empty range).

## Near-term actions (not yet done)

1. ~~**`.gitattributes` durable CRLF fix**~~ ✅ added (Bryan, `bryan/cleanup` PR).
2. **`serial-split-tap` build** (Peter P2 — **PAUSED**) — solder the isolated 6N137 tap + bench/field test
   per [`serial-split-tap/scope.md`](../changes/serial-split-tap/scope.md); resumes when the operator
   returns with measurement #1 (DAQ TXD idle voltage) + parts. Scope landed (PR #7).
3. ~~**Parser cleanup**~~ ✅ removed the off-path `internal/parser` (Bryan, `bryan/cleanup` PR).
4. **Totco preset** — when a Totco unit is reachable (same direct-laptop capture method).
5. ~~Phase 4b~~ ✅ (Bryan, PR #1). ~~Install commit gate~~ ✅ (S6). ~~Fix repo ruleset~~ ✅ (issue #3).

## Test surface

- `go test ./...`: `internal/daqformat`, `internal/store`, `internal/api` have tests;
  others report "no test files". Web has no unit suite (tsc-strict + Playwright screenshot are the checks).
- **P2 wrap run (2026-06-21, Windows):** docs-only arc (zero source change) — `go vet ./internal/...` ✅ ·
  `go test ./internal/...` ✅; `web/dist` present (embed intact).
- **Last full run (2026-06-21 P1 wrap, Windows):** `go test ./...` ✅ · `go vet ./...` ✅ · `gofmt -l`
  clean · `go build ./...` ✅ (embed) · web build (tsc strict + vite) ✅. (`make` absent on Windows — ran
  the steps directly; CGO-free, uPlot bundled offline.)
