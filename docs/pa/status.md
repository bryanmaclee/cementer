---
status: current
last-reviewed: 2026-06-12
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
| 2 | Real DAQ format + format mechanism (mapping + compute) | 🟢 **UNBLOCKED** — real data arrived | commit `ddf8ada` added 3 real Enbridge CSVs; format decoded below. Mechanism still unbuilt |
| 3 | Job CRUD + recording segments + Pump Profile CRUD + hello/profile message + scope-grouped display | ⬜ **NOT STARTED** | no job/profile/segment tables; no auth |
| 4 | uPlot charting (two config scopes) + printing (company default + per-job overrides) | ⬜ **NOT STARTED** | — |

## In-flight

- **PA workflow init (Session 1)** — instantiating `pa-base v1` into the cementer contract +
  scaffolding. See changelog 2026-06-12.

## ⚠ MAJOR FORK — needs deliberation (≥R2, axiom-level: storage engine + viz)

Commit `ddf8ada` (collaborator Peter Oliver, 2026-06-09) revealed a **parallel, different stack** built
to prove the hardware data flow end-to-end — and it does NOT match the cementer Go binary's stack:

| Concern | cementer Go binary (this repo's code) | Collaborator's working prototype (`ddf8ada`) |
|---|---|---|
| Ingest | Go: serial → rawlog → parser → store | Python script on the Pi parses the CSV |
| Store | **SQLite** (modernc, WAL, single-writer) | **InfluxDB 2.9.1** (`cement_data` bucket) |
| Viz | **custom embedded vanilla-TS** dark-mode client | **Grafana 13.0.2** dashboards |
| DAQ feed | replay file / real serial | laptop CSV → USB → **ESP32** → UART2 → Pi |

Collaborator's note: hardware flow "**Working!**"; "get proper DB in place and **serve it in whatever
way you feel best**"; "**Customize UI and charting (collaborator handoff)**". Read most plausibly: the
Python+InfluxDB+Grafana stack was a throwaway proof-of-concept, and the Go binary IS the
productization handoff — but the DB choice is left **explicitly open**. **Do not resolve silently.**
This is the storage-engine + visualization axiom (project fundamentally IS): surface as a deliberation
point. The ESP32 rig (`csvToSerialSend.ino`, `send_csv.py`) is a reusable real-data injector
regardless of which way the fork resolves — it feeds real CSV over serial into whatever ingests it.

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
- **Parser vs real format (UNVERIFIED):** `internal/parser` (`parser.DefaultConfig()`) was written
  against the synthetic `testdata/sample-stream.txt`; whether it handles the real 15-column Enbridge
  CSV is unverified. Don't hard-code this format into the parser — project axiom #2 says format
  adaptation is the no-code mapping/compute layer, not parser edits.

## Doc-currency / hygiene debts

- **Stale `docs/plan` reference.** `cmd/cementer/main.go` (pkg-doc line ~7 and a comment ~145) and
  `README.md` cite a build-plan doc that **does not exist**. Fix: create `docs/plan` OR correct the
  references to point at `data-model.md` § Build-order + this doc.
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
