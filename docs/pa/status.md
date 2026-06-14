---
status: current
last-reviewed: 2026-06-13
---

# cementer — live status (the SoT)

The single live source of truth for **done · in-flight · left**. Frozen planning prose (README
*Status*, `data-model.md` *Build order*) is NOT truth — this is. Layering: normative design
(`docs/design/data-model.md` + README architecture) → **this doc** → changelog → hand-off.

_Verify "is it shipped" claims against code (grep / `go build` / the SQLite schema), not this doc's
prose — but keep this doc honest at every wrap._

## Phase board

| # | Phase / step | State | Evidence (verified 2026-06-12) |
|---|---|---|---|
| P1 | Durable ingest → WS → dark-mode readout | ✅ **DONE** | `cmd/cementer/main.go` wires source→rawlog→parser→store→hub→WS + embedded SPA; `internal/store` SQLite WAL single-writer |
| 1 | Config-driven dynamic channels + theme + storage env | ✅ **DONE** | store `samples` keyed by `channel`; `web/src/theme.ts`; `-data-dir`/`$CEMENTER_DATA_DIR` in `main.go` |
| — | Recording start/stop model | 🟡 **DESIGNED, not built** | `data-model.md` § Recording (commit 94f02b6); **no `recording_segments` table** — store has only `samples` |
| 2 | **Intellisense** `DaqFormat` preset + format mechanism (mapping + compute) | 🟢 **UNBLOCKED** | format = **Intellisense** (ratified); define the preset from the 15-column Enbridge CSVs (`ddf8ada`), decoded below. Mechanism still unbuilt |
| 3 | Job CRUD + recording segments + Pump Profile CRUD + hello/profile message + scope-grouped display | ⬜ **NOT STARTED** | no job/profile/segment tables; no auth. Includes **retention/downsampling-as-code** (DD rider #3) |
| 4 | uPlot charting (two config scopes) + printing (company default + per-job overrides) | ⬜ **NOT STARTED** | print artifact = uPlot-at-high-DPI + print-CSS (not a dashboard export) |

## ✅ Bench-top stack validation — VERIFIED 2026-06-13 (Peter, on `CementSerial` / 10.0.0.105)

The Go+SQLite Pi stack is proven on **both** serial-ingress paths, single static aarch64 binary, no
recompile to switch source (only the `-serial` flag). Topology = laptop `send_csv.py` → ESP32
(`csvToSerialSend`) → [GPIO UART **or** CP2102 USB] → Pi `cementer`. **Simulated transport** (recorded
Enbridge CSV, not a live DAQ).

| Ingress | Device | rows | result |
|---|---|---|---|
| GPIO UART | `/dev/serial0`→`ttyS0` @115200 | 2,812 | ✅ raw log + SQLite WAL + `/debug/stats` 200 |
| USB adapter | CP2102→`/dev/ttyUSB0` (by-id) @115200 | 4,404 | ✅ fresh `~/cementer-usbtest` db |

**Proven:** serial RX, raw-log durability (L1), SQLite commit (L2), HTTP/WS serve across LAN, aarch64
binary. **NOT proven (still open):** real-DAQ **wire contract** (framing/timing/serial params — only
confirmed at the unit, this is the Phase 2 **D4** item) and **channel semantics** (4-col parser vs 15-col
format → Phase 2 no-code mapping). Field runbook + gotchas live in `hand-off.md` (⚡ FIELD RUNBOOK).
Build provenance: Go 1.26.4 on the garage desktop; web `dist/` stubbed (Node 18 < Vite 8); cross-compiled
`GOOS=linux GOARCH=arm64`; binary on the Pi (gitignored, not in repo).

## In-flight

- **Phase 2 SCOPED + decisions locked; build GATED.** [`scope.md`](../changes/phase2-intellisense-daqformat/scope.md).
  Generic `internal/daqformat` engine + Intellisense preset + minimal channel set; model/store already
  fit (no change). Decisions: D1 new package · D2 embedded LOGTIME (+server fallback) · D3 map `meta.*`
  now / semantics Phase 3 · **D4 GATE: get a live-serial capture before "done"**
  ([capture request](../changes/phase2-intellisense-daqformat/live-serial-capture-request.md)).
  **D4 status (2026-06-13):** bench-top capture DONE but **simulated transport** (ESP32-replayed CSV — see
  bench-validation block above); the **real-DAQ wire capture is still pending** — collaborator Peter has
  the Pi + RS-232→USB adapter in hand, field runbook ready in `hand-off.md`. Canonical dev agent
  **`cementer-go-engineer` active**. **Next: dispatch the engine+preset build (agent) + obtain the
  real-DAQ capture in the field.**

## ✅ RESOLVED FORK — storage engine + viz (RATIFIED 2026-06-12)

**Decision (user, R2):** adopt **(A) Go single-binary + SQLite(WAL) + custom uPlot UI**; **(B)
Python→InfluxDB→Grafana is retired to a dev/diagnostic bench** (`esp32sketches/`, `pi4b & test db/` —
real-data injection + ad-hoc exploration only, no claim on the product). Full rationale + sources:
[`docs/deep-dives/storage-and-viz-architecture-2026-06-12.md`](../deep-dives/storage-and-viz-architecture-2026-06-12.md)
(RATIFIED). **Engineering riders folded into the build plan:** explicit `PRAGMA synchronous=FULL` +
chosen commit cadence; retention/downsampling as scoped code → **Phase 3/4**; the print artifact is
uPlot-at-high-DPI + print-CSS (not a dashboard export). Background (kept for provenance):

| Concern | cementer Go binary (this repo's code) | Collaborator's working prototype (`ddf8ada`) |
|---|---|---|
| Ingest | Go: serial → rawlog → parser → store | Python script on the Pi parses the CSV |
| Store | **SQLite** (modernc, WAL, single-writer) | **InfluxDB 2.9.1** (`cement_data` bucket) |
| Viz | **custom embedded vanilla-TS** dark-mode client | **Grafana 13.0.2** dashboards |
| DAQ feed | replay file / real serial | laptop CSV → USB → **ESP32** → UART2 → Pi |

Collaborator's note: hardware flow "**Working!**"; "get proper DB in place and **serve it in whatever
way you feel best**"; "**Customize UI and charting (collaborator handoff)**". The ESP32 rig
(`csvToSerialSend.ino`, `send_csv.py`) is a reusable real-data injector — it stays as the dev bench's
real-CSV-over-serial feed for stack (A).

## Real DAQ format (decoded from `ddf8ada` CSVs)

Comma-delimited, **has header**, 15 columns. Maps cleanly onto `data-model.md`'s channel/scope model:

```
_00_LOGTIME      timestamp (Excel serial day-number, e.g. 46171.24 ≈ 2026; parse hint needed)
_01_DENSITY      density            _08_RATE_2        rate    (unit 2)
_02_PRESS        pressure (agg?)    _09_WTR_RATE      water rate
_03_PUMP_RATE    rate     (agg?)    _10_DENS_BKUP     density (backup → density.2)
_04_PUMP_TTL     volume   (job?)    _11_WATER_STG_TTL water stage total (scope=stage)
_05_PRESS_1      pressure (unit 1)  _12_PUMP_STG_TTL  pump  stage total (scope=stage)
_06_PRESS_2      pressure (unit 2)  _13_JOB_NUMBER    job number
_07_RATE_1       rate     (unit 1)  _14_MARKER        marker (stage/recording marker?)
```

This is a **real DaqFormat** to define against (data-model.md said the preset would come "from a real
CSV — incoming"). It is NOT confirmed to be "Intellisense" — it's this Enbridge job's format; classify
the preset name with the user. Files: `esp32sketches/EnbridgeCC4-16-CICR@344.csv` (2.5k rows),
`@3250.csv` (8.7k), `Shoe344.csv` (14.4k).

## Design ↔ code deltas (tracked TODOs)

- `recording_segments` (and `jobs`) tables: designed in `data-model.md`, absent from
  `internal/store/store.go` (only `samples`). Lands with Phase 3.
- Pump Profile / DAQ Format / hello-profile WS message: designed, no code yet.
- Computed/derived channels (`agg.rate = sum(...)`): designed, no code yet.
- **Parser vs real format (CONFIRMED MISMATCH):** `parser.DefaultConfig()` is the synthetic **4-channel**
  layout (pressure/rate/density/volume) — NOT the real **15-column** Enbridge format. Adaptation is the
  no-code mapping/compute layer (project axiom #2), not parser edits. (Surfaced by nav-map cold-start.)
- `internal/api/` and `web/src/chart/` are **empty placeholder dirs** for unbuilt phases (per nav-maps).

## Doc-currency / hygiene debts

- **Stale `docs/plan` reference.** `cmd/cementer/main.go` (pkg-doc line ~7 and a comment ~145) and
  `README.md` cite a build-plan doc that **does not exist**. Fix: create `docs/plan` OR correct the
  references to point at `data-model.md` § Build-order + this doc.
- **README Go version drift.** `README.md` says "Go 1.22+"; `go.mod` is `go 1.26.4`. (nav-map catch.)
- **Nav-maps generated** (`.claude/maps/`, 13 maps + non-compliance report, stamp `ee446c3`). Note:
  the mapper is scrml-flavored — its flag that the deep-dive "belongs in scrml-support" is a FALSE
  POSITIVE; this standalone repo keeps deep-dives in `docs/deep-dives/` by overlay design.
- **⚠ Plaintext credentials committed** in `pi4b & test db/credetials&currentDB.README` (commit
  `ddf8ada`): SSH / InfluxDB / Grafana logins (weak identical test passwords). Test-rig creds on a LAN
  Pi, but committing credentials is a flag — rotate + move to a non-committed secret if this repo is
  ever shared/public; consider `.gitignore` for the creds file. Surfaced, not changed (collaborator's
  file).

## Near-term actions (not yet done)

1. **Generate nav-maps:** run `/map` (cold start) → `.claude/maps/` (structure, dependencies, build,
   test, events, state, api). None exist yet.
2. **Install the commit gate:** no pre-commit hook exists (`core.hooksPath` unset). Baseline:
   `gofmt -l` + `go vet ./...` + `go build ./...` + `go test ./...`; `make build` pre-push when `web/`
   changed.
3. **Resolve the `docs/plan` debt** (above).

## Test surface

- `go test ./...` — only `internal/parser/parser_test.go` exists. Web has no tests.
- **Last full run (2026-06-12 wrap):** `go build ./...` ✅ · `go vet ./...` ✅ · `go test ./...` ✅
  (parser passes; all other packages report "no test files"). `web/dist` was present so the
  embed-dependent root + `cmd/cementer` compiled.
