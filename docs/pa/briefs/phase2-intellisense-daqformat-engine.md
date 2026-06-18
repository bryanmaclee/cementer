---
change-id: phase2-intellisense-daqformat
slug: engine
agent: cementer-go-engineer
model: opus
isolation: worktree
agent-id: a1824a1735ff07fc1
dispatched: 2026-06-18 (Session 5)
dispatched-from-tip: 2d28a3d
status: dispatched
---

# Dispatch brief (verbatim) — Phase 2 daqformat engine + Intellisense preset

> Archived per pa.md §5 brief-archival. The text below is the exact prompt sent to
> `cementer-go-engineer` (background, worktree, model: opus). Paths/SHAs are as-at-dispatch.

---

You are the canonical cementer Go dev-agent. Build **Phase 2**: the generic, config-driven `internal/daqformat` engine + the **Intellisense** preset, so adapting to a new pump format is *configuration, not a recompile* (project axiom #2). This is the first real Go source arc on cementer — get it idiomatic and durable.

═══════════════════════════════════════════════════════════════════════
0) STARTUP-VERIFICATION GATE (do this FIRST, before any write)
═══════════════════════════════════════════════════════════════════════
- Confirm cwd == your assigned worktree under `/home/bryan-maclee/...` and `git rev-parse --show-toplevel` == that worktree (it MUST NOT equal the integration root `/home/bryan-maclee/cementer`). Confirm a clean tree. If any check fails, STOP and report — do not proceed.
- Go is user-local: `export PATH=$HOME/.local/go/bin:$PATH` (the Makefile already does this). Verify `go version` (expect 1.26.x).
- The web client is `go:embed`'d (`assets.go` → `//go:embed all:web/dist`), so `go build ./cmd/cementer` FAILS until `web/dist` exists. For engine-only steps, build/test just `./internal/...`. For the final wire-in + E2E, prime the web build: `cd web && npm install && npm run build` (i.e. `make web`) — Node 22 is available, Vite 8 builds fine. (If for any reason the web build fails, a minimal stub `web/dist/index.html` is acceptable to satisfy the embed — the E2E here exercises the data pipeline, not the UI — but note it in your report.)

═══════════════════════════════════════════════════════════════════════
1) MAPS — REQUIRED FIRST READ
═══════════════════════════════════════════════════════════════════════
Read these `.claude/maps/` files first (stamp `ee446c3`, 2026-06-12; **current** — only docs/captures have landed since, no source moved, so treat as accurate): `structure.map.md`, `state.map.md`, `schema.map.md`, `api.map.md`, `test.map.md`. Follow their navigation; if a file moved past the stamp, treat the map as a verify-against-source hypothesis. Report whether the maps were load-bearing.

═══════════════════════════════════════════════════════════════════════
2) ANTI-PATTERN BRIEFING — READ BEFORE WRITING CODE
═══════════════════════════════════════════════════════════════════════
Read `docs/pa/anti-patterns.md` Part A (idiomatic Go) and re-read before each feature. Key points for this arc: accept-interfaces-return-structs (don't add speculative interfaces); handle every error (`%w`-wrap with context); no getters/setters on plain data; no goroutine leaks; keep protocol-specific code OUT of store/hub/web; gofmt + go vet clean always; standard library + existing deps only (NO new deps, no framework, no ORM).

═══════════════════════════════════════════════════════════════════════
3) NORMATIVE SOURCES + A CRITICAL CORRECTION
═══════════════════════════════════════════════════════════════════════
- Design source of truth: `docs/design/data-model.md` (Pump Profile / Channel / DaqFormat / FieldMap / ComputedChannel / scope model; recording-vs-live-vs-raw independence).
- Phase-2 engine design + decisions: `docs/changes/phase2-intellisense-daqformat/scope.md` (decisions D1–D4; engine type shapes; work breakdown). **USE IT FOR THE ENGINE DESIGN.**
- **AUTHORITATIVE PRESET SPEC:** `docs/changes/phase2-intellisense-daqformat/intellisense-wire-capture-2026-06-16.md` (the live-wire characterization).

**⚠ CRITICAL — read carefully.** `scope.md` (2026-06-12) was written against the Intellisense **CSV file-exports** in `esp32sketches/EnbridgeCC4-16*.csv` — which are **15 columns, header present, Excel-serial-day timestamp**. The later **live-wire capture** (`intellisense-wire-capture-2026-06-16.md`, 2026-06-16) proved the **real serial wire is a DIFFERENT shape: 14 columns, NO header, `HH:MM:SS`-uptime timestamp (col 0).** This is the textbook corpus-is-artifact case: the CSV was a file export, the wire is its own thing.
→ **Build the engine generic (per scope.md), but build the Intellisense PRESET from the findings doc (the live 14-col wire), NOT the scope.md 15-col CSV table.** Do NOT build an Excel-serial timestamp parser for Intellisense (the wire has no Excel-serial value). Do NOT include a 15th `_14_MARKER` column (the wire has only 14 fields, indices 0–13).

Project axioms you MUST honor (`pa.md` §"Project axioms"): #1 raw/live/recording strictly independent; #2 format adaptation is config, strictly NO code (the engine is generic; the preset is pure data — a new format must NEVER require editing parser/store/hub/web); #4 durability layered (single-writer SQLite is sacred — do not touch the write path); #5 stages orthogonal to recording.

═══════════════════════════════════════════════════════════════════════
4) CURRENT CODE (verified — what fits, what changes)
═══════════════════════════════════════════════════════════════════════
- `internal/model` — `Reading{Seq int64, TS time.Time, Values map[string]float64}` + `Samples()` (one row per channel). Already multi-channel/channel-keyed. **NO CHANGE.**
- `internal/store` — `samples(ts_us, channel, value)` keyed by channel; `Submit(Reading)` expands via `Samples()`; single-writer WAL. **NO CHANGE.** `/debug/stats` (`store.Stats()`) reports row/channel counts — your E2E asserts against it.
- `cmd/cementer/main.go` — `onCommit` broadcasts a whole `Reading`; pipeline is channel-agnostic. Currently builds `p := parser.New(parser.DefaultConfig())` and `handleLine` calls `p.Parse(line, time.Now())`. **This is the wire-in locus** (swap to the daqformat engine; add a `-format` flag).
- `internal/parser` — `Config{Delimiter string, Channels []string}` is positional index→name ONLY (no header/timestamp/transform/compute/scope). `DefaultConfig()` = synthetic 4-channel `pressure,rate,density,volume`. `parser_test.go` exists and must stay green.
- `internal/source` — `Replay` reads a file line-by-line via `bufio.Scanner` (default ScanLines strips trailing `\r`, so a `<CR><LF>` capture replays cleanly). `-source <file>` drives it.

═══════════════════════════════════════════════════════════════════════
5) TARGET DESIGN — new package `internal/daqformat` (decision D1)
═══════════════════════════════════════════════════════════════════════
Types (names are a guide; keep them idiomatic Go):
- `DaqFormat{ ID, Name, Delimiter string; HasHeader bool; ExpectedFields int; Timestamp TimestampSpec; Fields []FieldMap; Computed []ComputedChannel }`
- `FieldMap{ Column int; ChannelID string; Transform *Transform }`
- `Transform{ Scale, Offset float64 }` (value' = value*Scale + Offset; nil = identity)
- `ComputedChannel{ ChannelID string; Op string /* "sum" | "mean" */; Inputs []string; Transform *Transform }`
- `TimestampSpec{ Column int; Kind TimestampKind }` with `TimestampKind` ∈ at least `{ ServerStamp, HMSUptime }` (design the enum so `ExcelSerial`/`Unix`/`RFC3339` can be added later without a redesign — generic engine — but you only need to IMPLEMENT what Intellisense + synthetic require, i.e. ServerStamp). For Intellisense the embedded col-0 `HH:MM:SS` is *uptime, not a date*, so the Reading.TS MUST be the **server stamp** (honors decision D2: embedded is a hint only, server provides the real date). Recognize col 0 as the timestamp position so it is not accidentally field-mapped; do not try to derive a date from it.

Engine (the heart):
- `Engine` built from a `DaqFormat` (e.g. `daqformat.New(DaqFormat) *Engine`), holding a monotonic seq counter. Safe for a single ingest goroutine (same contract as today's parser).
- `Apply(line []byte, serverTS time.Time) (model.Reading, bool)`:
  1. trim; skip empty and `#`-comment lines → `(zero, false)`.
  2. split by Delimiter.
  3. **FIELD-COUNT GUARD (load-bearing, from the findings doc):** if the token count != `ExpectedFields`, skip the line → `(zero, false)`. The live wire emits torn fragments at power interruption (e.g. `?,,,,,,,,,,,,,00:00:00,...`); the raw log keeps the bytes, the structured store drops the bad line.
  4. if `HasHeader` and this is the first line, skip it (Intellisense: `HasHeader=false`).
  5. determine TS per `TimestampSpec` (Intellisense/synthetic: `serverTS`).
  6. for each `FieldMap`: parse `fields[Column]` as float64; tolerate an individual unparseable field by omitting just that channel (matches today's permissive parser); apply `Transform`; set `Values[ChannelID]`.
  7. compute pass: for each `ComputedChannel`, gather `Inputs` from `Values`, apply `Op` (sum/mean), apply `Transform`, set `Values[ChannelID]`. (No-op for Intellisense — it field-maps the emitted aggregates.)
  8. if `Values` empty → `(zero, false)`; else bump seq, return `Reading{Seq, TS, Values}`.
- **Do NOT filter warmup negatives** (density/pressure < 0 are real, not errors — raw fidelity; any clamping is a Phase-4 display concern).

Presets (pure data — this is what makes it "no-code"):
- `func Intellisense() DaqFormat` — `ID:"intellisense"`, `Name:"Intellisense"`, `Delimiter:","`, `HasHeader:false`, `ExpectedFields:14`, `Timestamp:{Column:0, Kind:ServerStamp}`, `Computed:nil`, and these 13 `Fields` (col → channelId), straight from the findings doc:

  | col | channelId | col | channelId |
  |-----|-----------|-----|-----------|
  | 1 | `density.1`     | 8  | `unit2.rate`      |
  | 2 | `agg.pressure`  | 9  | `water.rate`      |
  | 3 | `agg.rate`      | 10 | `density.2`       |
  | 4 | `vol.job`       | 11 | `vol.water.stage` |
  | 5 | `unit1.pressure`| 12 | `vol.stage`       |
  | 6 | `unit2.pressure`| 13 | `job.number`      |
  | 7 | `unit1.rate`    |    |                   |

  (col 2 `agg.pressure` is the DAQ-EMITTED sum of unit pressures → field-mapped, NOT computed, exactly as data-model.md says aggregates the pump provides are field-mapped.)
- `func Synthetic() DaqFormat` — replicate the existing 4-channel synthetic layout so the Phase-1 replay path keeps working: `Delimiter:","`, `HasHeader:false`, `ExpectedFields:4`, `Timestamp:{Kind:ServerStamp}`, `Fields:` col0→`pressure`, col1→`rate`, col2→`density`, col3→`volume`. (Comment `#` lines are skipped by Apply.)

Minimal channel vocabulary (bundled metadata — NOT the Phase-3 Pump Profile CRUD, just a default `Channel[]` so the engine/later-client have id/role/scope/uom/decimals/label):
- `func IntellisenseChannels() []Channel` (define a small `Channel{ ID, Role, Scope, UoM, Label string; Decimals int }` struct — or reuse data-model naming). UoM values are **project-attested** (see `testdata/sample-stream.txt` header: pressure(psi), rate(bbl/min), density(ppg), volume(bbl); density 8.21 ppg was empirically confirmed at the rig):
  - density.1 → density, unit:1, **ppg**, "Density", dec 2
  - agg.pressure → pressure, aggregate, **psi**, "Pressure (total)", dec 0
  - agg.rate → rate, aggregate, **bbl/min**, "Rate (total)", dec 2
  - vol.job → volume, job, **bbl**, "Job Volume", dec 1
  - unit1.pressure → pressure, unit:1, **psi**, "Unit 1 Pressure", dec 0
  - unit2.pressure → pressure, unit:2, **psi**, "Unit 2 Pressure", dec 0
  - unit1.rate → rate, unit:1, **bbl/min**, "Unit 1 Rate", dec 2
  - unit2.rate → rate, unit:2, **bbl/min**, "Unit 2 Rate", dec 2
  - water.rate → rate, aggregate, **bbl/min**, "Water Rate", dec 2
  - density.2 → density, unit:1 (backup), **ppg**, "Density (backup)", dec 2
  - vol.water.stage → volume, stage, **bbl**, "Water Stage Volume", dec 1
  - vol.stage → volume, stage, **bbl**, "Stage Volume", dec 1
  - job.number → meta/job, (no scope/uom), "Job Number", dec 0

Wire-in (`cmd/cementer/main.go`):
- Add `-format string` flag, default `"intellisense"`. Resolve `"intellisense"` → `daqformat.Intellisense()`, `"synthetic"` → `daqformat.Synthetic()`; unknown value → a clear error.
- Replace `p := parser.New(parser.DefaultConfig())` + `p.Parse(line, time.Now())` with `eng := daqformat.New(format)` + `eng.Apply(line, time.Now())`.
- Recommended composition (your judgment, stay idiomatic): make `daqformat` self-contained (own tokenize/map/compute/TS/seq/guard); leave `internal/parser` + `parser_test.go` in place and green (it's the Phase-1 artifact). If parser ends up off the main path, say so in your report as a follow-up cleanup candidate — do NOT delete it in this arc.
- **Opportunistic doc fix while you're in main.go:** the package-doc comment line ~7 says `Pipeline (see docs/plan): ...` but `docs/plan` does not exist — change that reference to `docs/design/data-model.md`.

═══════════════════════════════════════════════════════════════════════
6) WORK BREAKDOWN — one WIP-commit per unit (crash-recovery)
═══════════════════════════════════════════════════════════════════════
Commit after EACH unit (WIP commits expected — don't batch). Maintain an append-only `docs/changes/phase2-intellisense-daqformat/progress.md` with timestamped lines (done / next / blockers) updated after each step. If you crash, your commits + progress.md are how the next agent resumes.
1. `internal/daqformat` types + doc comments (no behavior).
2. `Apply` engine: tokenize + comment-skip + field-count guard + header-skip + field-map + transform; table-driven unit tests on crafted lines (incl. a torn `?,,,...` line that must be skipped, and a warmup-negative that must pass through).
3. Compute pass (sum/mean + transform); unit test (`agg.rate = sum(unit1.rate, unit2.rate)` on a crafted line).
4. Intellisense + Synthetic presets + `IntellisenseChannels()`; table-driven test mapping a REAL captured line → expected channel-keyed values. Use this real idle line from `captures/capture-2026-06-16T150318-19200-8N1.bin`: `14:53:41,0.04,0,0.00,42.5,0,0,0.00,0.00,0.0,0.00,0.0,42.5,0` → assert e.g. `density.1=0.04`, `vol.job=42.5`, `vol.stage=42.5`, and the full 13-channel set present.
5. Wire `-format` into `main.go`; keep the synthetic path green. Fix the `docs/plan` reference.
6. End-to-end verification (next section).
7. Docs: update `README.md` if you moved protocol mapping off `internal/parser` (the "the only protocol-specific code" line + the layout table), and add a one-line `-format` note to the flags list. Leave nav-map regeneration to the PA.

═══════════════════════════════════════════════════════════════════════
7) VERIFY-BEFORE-CLAIM (do NOT mark done on unit tests alone)
═══════════════════════════════════════════════════════════════════════
After building, run the REAL wire end-to-end on the post-change binary:
- Build the binary (prime `web/dist` first): `make build` (or `make web` then `go build ./cmd/cementer`).
- Run against a REAL capture (the live wire — NOT the Enbridge CSV, which is the superseded 15-col export):
  `./cementer -source captures/capture-2026-06-16T161347-19200-8N1-pressure.bin -format intellisense -replay-interval 20ms -replay-loop=false -data-dir /tmp/cementer-daqtest`
  (the pressure capture shows cols 2 & 5 move together = `agg.pressure == unit1.pressure` when only unit 1 is pressurized — the sum relationship from the findings doc.)
- **Symptom check (NOT "tests pass"):**
  - `curl -s localhost:8080/debug/stats` → assert the distinct-channel count == **13** and they are exactly the Intellisense channel ids; row counts plausible (≈ lines × 13).
  - Or query the db directly: `sqlite3 /tmp/cementer-daqtest/cementer.db 'select count(*), count(distinct channel) from samples'` and `'select distinct channel from samples order by 1'`.
  - Spot-check a value: in the pressure capture, `agg.pressure` and `unit1.pressure` should reach the same nonzero max while `unit2.pressure` stays 0 (single-unit rig) — confirm via a quick `select channel, max(value) from samples group by channel`.
- Also confirm the **synthetic path still works**: `./cementer -source testdata/sample-stream.txt -format synthetic -data-dir /tmp/cementer-syn` → 4 channels (pressure/rate/density/volume) present.
- Run `gofmt -l` (must list nothing), `go vet ./...`, `go test ./...` (all green, incl. `parser_test.go`). Record the exact commands + output in your report.

═══════════════════════════════════════════════════════════════════════
8) INVARIANTS (must all hold at the end)
═══════════════════════════════════════════════════════════════════════
- A NEW pump format requires only a new `DaqFormat` value — zero edits to parser/store/hub/web (axiom #2).
- Phase-1 synthetic replay path stays green; store single-writer untouched; model untouched; NO new deps.
- TS for Intellisense = server-stamp; field-count guard skips torn lines; warmup negatives pass through unfiltered.
- gofmt/vet/test clean.

═══════════════════════════════════════════════════════════════════════
9) REPORT BACK (your final message = data for the PA, not prose for a human)
═══════════════════════════════════════════════════════════════════════
Report: (a) worktree path + branch + final tip SHA; (b) files-touched list; (c) the exact verify commands run + their output (the /debug/stats or sqlite channel set + counts, the gofmt/vet/test results); (d) whether the maps were load-bearing; (e) any deferred items / follow-ups (e.g. parser now off the main path?); (f) anything that contradicted this brief or the scope/findings docs. Ensure `git status` is clean in the worktree before reporting DONE (commit everything — "work in the worktree, uncommitted" is NOT acceptable).
