---
status: current
last-reviewed: 2026-06-29
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
- **B6 complete (wrapped 2026-06-22).** Phase 4b (MVP) merged (**PR #1**); the full multi-operator
  workflow — DD, commit gate, `coord` branch, meta-doc partition, **`pa.md` overlay v2** — merged (**PR #6**).
  Shared cleanup (`.gitattributes` + dead-`internal/parser` removal) on **`bryan/cleanup` (PR open, awaiting
  merge)**. Coordination proven live against Peter's P1/P2 (clean merge — partition held). Claim reset to idle.
- **Next (Bryan):** merge `bryan/cleanup`; **regenerate nav-maps** (stale at S5 `1465bd9`); broaden the
  pre-commit gate to catch deletions (`--diff-filter` incl. `D`). Idle otherwise.

### Peter
- **P1 (2026-06-21, Windows field laptop):** adopted the S6 multi-party model (PR-flow + coord +
  meta-doc partition); landed P1 onboarding docs via **PR #2 → `main` `0a96095`**. Stood up the Windows
  toolchain (Go 1.26.4 + Node 24.17.0), installed the commit gate, fixed a Windows CRLF/gofmt break
  (`autocrlf`), and **PA-verified Phase 4b end-to-end** (built/ran/recorded → report + print render via
  Edge headless). Filed ruleset issue **#3** to Bryan (exempt `coord`; allow feature-branch deletion).
  - _Ruleset blocks RESOLVED same-day (issue #3 closed, verified):_ Bryan scoped the rules → coord
    push-direct + merged-branch deletion both work; coord pushed/synced (`d1028bc`); P1 wrap landed (PR #4).
  - _All P1 follow-ups now resolved:_ `.gitattributes` durable CRLF fix ✅ + parser cleanup ✅ (both
    Bryan, PR #10 `ac2dd16`); `pa.md` topology rewrite ✅ (Bryan, PR #6 `42ef5f2`).
- **P2 (2026-06-21):** opened the **`serial-split-tap`** hardware arc — designed an isolated, listen-only
  serial tap (6N137 opto → Pi GPIO UART) so the Pi can ingest a live DAQ stream **without disturbing the
  existing consumer**. **Scope doc landed on `main` (PR #7, `1b942eb`)**; build **PAUSED** pending operator
  measurement #1 (DAQ TXD idle voltage). Spec: [`serial-split-tap/scope.md`](../changes/serial-split-tap/scope.md).
- **P3 (2026-06-23):** doc-currency reconcile after Bryan's B6/cleanup (PR #10 `ac2dd16`) — synced `main`
  + coord, confirmed both P1 follow-ups (`.gitattributes`, parser cleanup) landed by Bryan, fixed the stale
  "still open" note in this block. **P2 `serial-split-tap` build remains PAUSED** on operator measurement #1.
- **P4 (2026-06-25):** **resumed the `serial-split-tap` BUILD** -- operator measured **#1** for BOTH
  DAQs (Intellisense **-6.35 V** / Totco **-8.20 V** idle), clearing the blocker. Issued the Intellisense
  channel-1 build sheet; **new finding: Totco TX is DTR-gated** (streams only while the consumer asserts
  DTR/pin4 -- listen tap validates in coexistence, not Pi-only). Build now in the operator's hands;
  findings folded into [`serial-split-tap/scope.md`](../changes/serial-split-tap/scope.md). Wrapped before solder.
- **P5 (2026-06-27):** **serial-split tap PROVEN end-to-end on breadboard (step-1 bench gate PASSED).**
  Plan pivot: operator got a **Waveshare USB->RS232** -> bench source is now the Waveshare run as a
  transmitter (real RS-232; superseded the field-adapter + a considered ESP32-TTL "Option B"). Debugged
  through a **parallel-wired 1N4148** (fixed to antiparallel), a **DOA 6N137** (swapped a spare), opto
  **under-drive** (`Rin` 1k->560 Ω on the weak Waveshare), and a **Pi mini-UART 9600 baud trap**.
  Cross-compiled a current `cementer-arm64-new` (Pi's old binary lacked `-format`), proved ingest
  (`/debug/stats` 208->1079 rows) + the **live chart over WiFi**. Local-only wrap; pushes deferred to P6.
- **P6 (2026-06-28):** **serial-split tap step-1 gate PASSED on the SOLDERED PROTO; `Rin` locked at 1 k.**
  Re-tuned `Rin` with the good chip gauging at the **real +6.35 V amplitude** (PSU static inject -> Vo swings
  3.3 V ↔ 0.059 V, ~4.9 mA tap load), soldered the protoboard, re-ran the full gate end-to-end (clean
  14-field lines -> cementer ingest -> live chart). Found + fixed a **DAQ-GND -> cathode (pin 3) open joint**
  that left Vo stuck high (mark-path clamp intact, space-path LED loop open). Cleared the P3+P4+P5 push
  backlog: **branch pushed + PR to `main` opened.** Recipe + findings in
  [`serial-split-tap/scope.md`](../changes/serial-split-tap/scope.md) "P6 soldered-proto validation".
- **P7 (2026-06-29, field laptop `P-Tech1`):** **Intellisense DB9 split-off FIELD-VERIFIED on a real DAQ —
  steps 2 + 3 PASSED.** Real Intellisense wire (DB9 pin 2 = TXD, pin 5 = GND, ~-5.5 V) -> opto -> Pi mini-UART
  -> cementer -> SQLite -> **live chart over WiFi** (phone hotspot), clean 14-field lines, **and zero
  disturbance to the production consumer** (the cementer laptop, Pi powered + unpowered). `Rin` = 1 k frames
  clean at the field amplitude. Field gotchas captured: the DMM is the wrong instrument on a live data line
  (Vo "3.3->3.06 V bumping" is the *good* signature, not a fault); WiFi-via-microSD-boot-partition
  `wpa_supplicant.conf` workaround (multi-network + `country=US`); `ERR_CONNECTION_REFUSED` = reachable but
  cementer not running. **Design clarification (operator):** the end product is a **permanent inline
  pass-through** (passive continuous through-wire so the consumer survives any Pi-side failure) that broadcasts
  WiFi in parallel — NOT a removable branch. **Next: the v2 Amphenol pass-through prototype** (map the 6-pin
  pinout -> build -> garage gate -> field). Intellisense parallel-splitter MVP BEFORE Totco. Findings folded
  into [`serial-split-tap/scope.md`](../changes/serial-split-tap/scope.md) "P7 field validation".

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
2. **`serial-split-tap` build** (Peter — **Intellisense DB9 split-off FIELD-VERIFIED on a real DAQ, P7
   2026-06-29; steps 2 + 3 PASSED**) — real wire -> opto -> Pi -> cementer -> live chart over WiFi, **zero
   disturbance to the production consumer**. Field test (roadmap step 1) DONE. **Next (operator directive, in
   order): (a) build the v2 hardware prototype with the Amphenol pass-through connectors — permanent inline,
   passive continuous through-wire (prereq: map the 6-pin Amphenol pinout), (b) garage-test the Amphenol proto
   through this same bench gate, then field-test.** **Intellisense parallel-splitter MVP must be done BEFORE
   Totco.** Field recipe + findings (DMM-on-live-line; WiFi-via-SD `wpa_supplicant.conf`; the corrected
   inline/passive-pass-through design note) in
   [`serial-split-tap/scope.md`](../changes/serial-split-tap/scope.md) "P7 field validation".
3. ~~**Parser cleanup**~~ ✅ removed the off-path `internal/parser` (Bryan, `bryan/cleanup` PR).
4. **Totco preset / 2nd channel** — **DEFERRED behind the Intellisense parallel-splitter MVP** (operator
   directive P6). Resume only after Intellisense is field-proven: same circuit, 2nd 6N137, 9600 8N1, `Rin`
   1.5 k; DTR-gated so coexistence-validate; pin-4 DTR jumper confirm test; map the 6-pin Amphenol pinout.
5. ~~Phase 4b~~ ✅ (Bryan, PR #1). ~~Install commit gate~~ ✅ (S6). ~~Fix repo ruleset~~ ✅ (issue #3).

## Test surface

- `go test ./...`: `internal/daqformat`, `internal/store`, `internal/api`, `internal/printcfg` have tests;
  others report "no test files" (`internal/parser` removed in B6). Web has no unit suite (tsc-strict +
  Playwright screenshot are the checks).
- **P7 wrap run (2026-06-29, field laptop `P-Tech1`):** hardware + docs arc, zero Go/web *source* change --
  `go vet ./...` ok · `go test ./...` ok (api/daqformat/printcfg/store pass; rest no test files). `web/dist`
  present (embed intact); no cross-compile needed this wrap.
- **P6 wrap run (2026-06-28, Windows):** hardware + docs arc, zero Go/web *source* change -- `go vet ./...`
  ok · `go test ./...` ok (api/daqformat/printcfg/store pass; rest no test files). `web/dist` present (embed
  intact); no cross-compile needed this wrap.
- **P5 wrap run (2026-06-27, Windows):** docs + tooling arc, zero Go/web *source* change (added
  `tools/intellisense-send.ps1` + cross-compiled artifacts, both non-source) -- `go vet ./...` +
  `go test ./...` recorded at wrap. NOTE: laptop `web/dist` was a stale 315-byte placeholder until P5
  rebuilt it (needed Node 20+; laptop was on Node 18 -> upgraded to 24.18.0). The cross-compiled Pi binary
  (`cementer-arm64-new`) embeds the rebuilt SPA.
- **P4 wrap run (2026-06-25, Windows):** docs-only arc (zero source change) -- `go vet ./internal/...` ok ·
  `go test ./internal/...` ok; `web/dist` present (embed intact).
- **P3 wrap run (2026-06-23, Windows):** docs-only arc (zero source change) — `go vet ./...` ✅ ·
  `go test ./...` ✅ (api/daqformat/printcfg/store pass) · `gofmt -l` clean.
- **P2 wrap run (2026-06-21, Windows):** docs-only arc (zero source change) — `go vet ./internal/...` ✅ ·
  `go test ./internal/...` ✅; `web/dist` present (embed intact).
- **Last full run (2026-06-21 P1 wrap, Windows):** `go test ./...` ✅ · `go vet ./...` ✅ · `gofmt -l`
  clean · `go build ./...` ✅ (embed) · web build (tsc strict + vite) ✅. (`make` absent on Windows — ran
  the steps directly; CGO-free, uPlot bundled offline.)
