---
status: current
last-reviewed: 2026-06-12
change-id: phase2-intellisense-daqformat
phase: 2
depends-on: architecture ratified (Go+SQLite+uPlot); real Intellisense CSVs in esp32sketches/
---

# Phase 2 scope — Intellisense `DaqFormat` preset + no-code mapping/compute layer

## Goal

Turn the real **Intellisense** serial format into channel-keyed `Reading`s through a **generic,
config-driven mapping + compute engine** — so that adapting to a new pump format is *configuration, not
a recompile* (project axiom #2). Deliverable: (a) a generic format engine, (b) the **Intellisense
preset** defined from the Enbridge CSVs, (c) the minimal channel vocabulary the preset keys against.

## Boundary — what Phase 2 IS and ISN'T

| In scope (Phase 2) | Out of scope (→ Phase 3/4) |
|---|---|
| Generic `DaqFormat` engine: delimiter, header, timestamp, field-map, transforms, compute layer | In-UI no-code editing of mappings (config-file/preset-driven for now) |
| The **Intellisense** preset (concrete `DaqFormat`) shipped as a bundled template | Pump Profile CRUD UI + the hello/profile WS message + scope-grouped display |
| A minimal **default Intellisense channel set** (the `Channel[]` the preset maps onto) | Job CRUD, recording segments, stage semantics from `_14_MARKER`/`_13_JOB_NUMBER` |
| Embedded-timestamp parsing (`_00_LOGTIME` Excel-serial → `time.Time`) | uPlot charting + printing |
| Tests + real-CSV end-to-end verification | Retention/downsampling (DD rider #3 → Phase 3/4) |

**The "no-code" nuance:** Phase 2 makes format adaptation require *no Go recompile* (edit a preset/config
value). Full *no-code-in-the-UI* editing is Phase 3. State this boundary so "no-code" isn't over-claimed.

## Current state (verified — what already fits, what changes)

- **`internal/model`**: `Reading{Seq, TS, Values map[string]float64}` + `Samples()` → one row per channel.
  **Already multi-channel.** No change needed. (The `TS` doc-comment already anticipates switching from
  server-stamp to an embedded pump timestamp — exactly `_00_LOGTIME`.)
- **`internal/store`**: `samples(ts_us, channel, value)` keyed by channel; `Submit(Reading)` expands via
  `Samples()`. **No change.**
- **`cmd/cementer/main.go`**: `onCommit` broadcasts a whole `Reading`; pipeline is channel-agnostic.
  **No change** (beyond wiring the chosen format into `parser.New(...)`).
- **`internal/parser`** (the change locus): `Config{Delimiter, Channels []string}` is **positional
  index→name only** — no header handling, no timestamp extraction, no transforms, no compute, no
  channelId/scope concept. `DefaultConfig()` is the **synthetic 4-channel** layout
  (`pressure,rate,density,volume`) matching `testdata/sample-stream.txt` — it must be *replaced by
  config*, not extended ad hoc.

## The real Intellisense format (curated from `esp32sketches/EnbridgeCC4-16*.csv`)

Comma-delimited, **header present**, 15 columns. Sample rate ~1 Hz (LOGTIME Δ ≈ 1.04 s). One file can
contain **multiple jobs** (`_13_JOB_NUMBER` went 1→2→3 within `Shoe344.csv`).

| Col | Header | Proposed channelId | role / scope | uom | Notes (from real data) |
|----|--------|--------------------|--------------|-----|------------------------|
| 0 | `_00_LOGTIME` | *(timestamp, not a channel)* | — | — | **Excel serial day-number** (epoch 1899-12-30), e.g. 46170.29 ≈ 2026. Parse → `time.Time`. |
| 1 | `_01_DENSITY` | `density.1` | density | ppg | -4.989 .. 16.53 (negatives = warmup/uncalibrated) |
| 2 | `_02_PRESS` | `agg.pressure` | pressure / aggregate | psi | 0 .. 967 (pump **emits** the aggregate → field-map, not compute) |
| 3 | `_03_PUMP_RATE` | `agg.rate` | rate / aggregate | bbl/min | 0 .. 4.725; **verified = RATE_1 + RATE_2** (emitted aggregate) |
| 4 | `_04_PUMP_TTL` | `vol.job` | volume / job | bbl | cumulative job volume |
| 5 | `_05_PRESS_1` | `unit1.pressure` | pressure / unit:1 | psi | -17 .. 865 |
| 6 | `_06_PRESS_2` | `unit2.pressure` | pressure / unit:2 | psi | -71 .. 967 |
| 7 | `_07_RATE_1` | `unit1.rate` | rate / unit:1 | bbl/min | 0 .. 4.239 |
| 8 | `_08_RATE_2` | `unit2.rate` | rate / unit:2 | bbl/min | 0 .. 4.464 |
| 9 | `_09_WTR_RATE` | `agg.waterRate` | rate / aggregate | bbl/min | 0 in these files (present in format) |
| 10 | `_10_DENS_BKUP` | `density.2` | density | ppg | backup densitometer |
| 11 | `_11_WATER_STG_TTL` | `vol.water.stage` | volume / stage | bbl | resets per stage (axiom #5) |
| 12 | `_12_PUMP_STG_TTL` | `vol.stage` | volume / stage | bbl | resets per stage |
| 13 | `_13_JOB_NUMBER` | `meta.job` | meta | — | increments within a stream (1→2→3) |
| 14 | `_14_MARKER` | `meta.marker` | meta / event | — | sparse pulse (35/14419 = 1); pump/DAQ stage marker, **NOT** the record button (axiom #1/#5) |

*(channelIds above are a proposal — confirm naming against `data-model.md`'s `Channel.id` examples
`unit1.pressure`, `agg.rate`, `vol.stage`, `density.1`. `meta.*` for job/marker is new — see decisions.)*

## Target design

1. **Generic format engine** (new `internal/daqformat`, see decision D1). Types from `data-model.md`:
   - `DaqFormat{ id, name, delimiter, hasHeader bool, timestamp *TimestampSpec, fields []FieldMap }`
   - `FieldMap{ column int|name, channelId string, transform *Transform }` (`Transform{ scale, offset }`)
   - `ComputedChannel{ channelId, op (sum|mean|…), inputs []string, transform }`
   - `TimestampSpec{ column, kind: excel-serial|unix|rfc3339|server }`
   - `Apply(rawLine []byte) (model.Reading, bool)`: tokenize → (skip header) → extract TS → field-map
     with transforms → run compute pass → emit channel-keyed `Reading`. Generic; **format = data.**
2. **Intellisense preset** — a concrete `DaqFormat` value (Go literal and/or bundled JSON) built from
   the table above. Aggregates are field-maps (the pump emits them); the compute layer ships but is a
   no-op for Intellisense (it exists for pumps that *don't* emit aggregates).
3. **Minimal channel vocabulary** — a bundled default Intellisense `Channel[]` (id/role/scope/uom/
   decimals/label) so the engine + (later) client have channel metadata, *without* the Phase-3 Pump
   Profile CRUD.
4. **Wire-in** — `main.go` selects the format: a `-format intellisense` flag (default) resolving to the
   preset, replacing the hard-coded `parser.DefaultConfig()`. Keep the synthetic format as
   `-format synthetic` for the existing replay/tests.

## Work breakdown (ordered; each a WIP-commit unit)

1. Add `internal/daqformat` types (`DaqFormat`, `FieldMap`, `Transform`, `ComputedChannel`,
   `TimestampSpec`) + doc comments. *(no behavior yet)*
2. Excel-serial timestamp parser + unit test (golden: 46170.290613 → expected UTC instant).
3. `Apply` engine: tokenize + header-skip + field-map + transform; unit tests on crafted lines.
4. Compute pass (sum/mean + scale/offset); unit test (`agg.rate = sum(unit1.rate, unit2.rate)`).
5. Intellisense preset + default channel set; table-driven test mapping a real CSV row → expected
   channel values.
6. Wire `-format` into `main.go`; keep synthetic path green.
7. **End-to-end verification** (do-not-claim-done-without): replay `EnbridgeCC4-16Shoe344.csv` (direct
   replay source, or via the ESP32 bench) and assert `/debug/stats` shows the expected distinct channels
   + plausible row counts; spot-check a known row through `/ws/live`. Symptom check = channel set + value
   sanity, **not** "unit tests pass".
8. Update README/data-model preset note + the nav-maps (`schema`, `api`, `state`, `structure`).

## DECISIONS (resolved 2026-06-12)

- **D1 — engine placement: new `internal/daqformat` package** (PA recommendation, not vetoed). `parser`
  stays the tokenizer; `daqformat` does map+compute.
- **D2 — timestamp: embedded `_00_LOGTIME` (Excel-serial) as `Reading.TS`, server-stamp fallback** via
  `TimestampSpec{kind: server}` when a format carries no timestamp. (USER.)
- **D3 — `meta.job` / `meta.marker`: mapped as channels now; semantics deferred to Phase 3** (PA
  recommendation, not vetoed).
- **D4 — live-serial fidelity: GET A LIVE-SERIAL CAPTURE FIRST (USER).** The engine + preset may be
  built (format-agnostic), but **Phase 2 is GATED: not "done" until validated against a real
  live-serial dump.** Capture request drafted: [`live-serial-capture-request.md`](./live-serial-capture-request.md)
  — relay to the hardware collaborator. The build dispatch waits on (or runs in parallel with, but does
  not close before) that capture.
- **Dev agent: forge `cementer-go-engineer` (DONE — effective next session).** The canonical cementer
  source-change dev agent now exists at `~/.claude/agents/cementer-go-engineer.md`; it activates at the
  NEXT session start (harness caches agent defs at start). Dispatch Phase 2 through it then.

## Original decision write-ups (for provenance)

- **D1 — engine placement (R1, PA-recommend):** new `internal/daqformat` package; `parser` becomes a
  thin tokenizer or is absorbed. *Recommend: new package; keep `parser` as the tokenizer it already is,
  `daqformat` does map+compute.* Confirms axiom #2 (format=config, engine=generic). → **your veto only.**
- **D2 — timestamp policy (R2, USER):** use embedded `_00_LOGTIME` as `Reading.TS`, or server-stamp?
  Embedded is right for replaying historical captures + accurate job charts; server-stamp is a fallback
  when a live pump omits a timestamp. *Recommend: prefer embedded LOGTIME, fall back to server-stamp via
  `TimestampSpec{kind: server}`.* **Needs your call** (affects chart time-axis fidelity).
- **D3 — `meta.job` / `meta.marker` handling (R1→R2):** store `_13`/`_14` as channels now and defer
  their *semantics* (job boundaries, stage markers) to Phase 3, or drop them from the preset until then?
  *Recommend: map them as `meta.*` channels (captured, keyed, harmless), wire semantics in Phase 3.*
- **D4 — live-serial fidelity (RISK, must verify):** we only have CSV **exports**, not a **live-serial**
  capture. The on-wire frames may lack the header, may not carry `_00_LOGTIME`, or may differ in shape.
  The Intellisense preset is a strong starting point but **Phase 2 cannot be claimed done against
  exports alone** — it needs validation against a real live-serial dump (or an explicit decision that
  the ESP32-replay-of-CSV *is* the accepted wire contract). **Surface to user / the hardware
  collaborator.**

## Risks / unknowns

- **Live-serial vs CSV-export mismatch** (D4) — the biggest. Mitigate: ask the collaborator for a raw
  live-serial capture, or ratify the CSV-export shape as the contract.
- **Warmup negatives** (density/pressure < 0) are real, not errors — do NOT filter them in the engine
  (raw fidelity); any clamping is a display concern (Phase 4).
- **Other Intellisense pumps may vary** (column count/order). The preset is *this* job's Intellisense;
  the engine's genericity is what absorbs variants — don't bake Intellisense assumptions into the engine.
- **Excel-serial epoch** — confirm 1899-12-30 (Excel's 1900 leap-year bug offset) vs 1900-01-01; verify
  a known LOGTIME against the CSV filename/job date.

## Test + verify strategy

- Unit: Excel-serial parse (golden), `Apply` field-map + transform, compute sum/mean, header-skip.
- Golden/table: real CSV row → expected channel-keyed values (use a pinned row from `Shoe344.csv`).
- E2E (verify-before-claim, §8): replay the real CSV end-to-end on the post-change binary; assert the
  distinct channel set at `/debug/stats` and value sanity at `/ws/live`. Do not mark done on units alone.

## Recommended dispatch

- **Agent:** consider forging `cementer-go-engineer` (`/forge go`) first (this is the first real Go
  source arc); else interim canonical `general-purpose`, `model: opus`, `isolation: "worktree"`.
- **Worktree startup-prime:** `web/dist` is gitignored + `go:embed`-required → `cd web && npm install &&
  npm run build` before `go build ./cmd/cementer`, OR build only `./internal/...` for engine-only steps.
- **MAPS — required first read:** `schema`, `state`, `api`, `structure` (`.claude/maps/`, stamp ≥ `b0fef5f`).
- **Crash-recovery:** WIP-commit per work-breakdown step; progress log at
  `docs/changes/phase2-intellisense-daqformat/progress.md`.
- **Brief archival:** archive the verbatim dispatch to `docs/pa/briefs/phase2-intellisense-daqformat-<slug>.md`.
