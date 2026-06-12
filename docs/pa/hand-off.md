# Hand-off — live

`as of: Session 1 · 2026-06-12`

> Optimize for the NEXT session's pickup, not this session's terseness. Bloat is acceptable;
> under-documentation is not. Rotate to `docs/pa/archive/hand-off-<date>.md` at next session start.

## Open questions (top of mind)

- **Forge `cementer-go-engineer`?** No dedicated Go dev-agent exists; interim canonical = `general-purpose`
  (worktree-isolated, `model: opus`). Decide whether to `/forge go` now or wait for source churn.
  (Becomes relevant as Phase 2 source work starts.)
- **`docs/plan` debt:** create the file or fix the `main.go` / README references? (See status.md.)
- **Commit gate:** install the baseline pre-commit now, or defer? (See status.md near-term action 2.)
- **Committed credentials** in `pi4b & test db/...README` — rotate / gitignore? (hygiene flag, status.md)

## Resolved this session

- **✅ Architecture fork RATIFIED** — adopt (A) Go+SQLite+uPlot; retire (B) Influx/Grafana to dev bench.
  Engineering riders folded into the plan (status.md): `synchronous=FULL`, retention-as-code → Phase 3/4,
  uPlot-print-CSS for the print artifact.
- **✅ DAQ format named: Intellisense** (the 15-column `_NN_` layout) — unblocks Phase 2 preset work.

## State as of close (Session 1)

| Item | State |
|---|---|
| PA contract | ✅ `pa.md` (overlay v1) + `pa-base.md` (vendored `pa-base v1`) landed |
| Scaffolding | ✅ `docs/pa/{status,hand-off,user-voice,design-insights,changelog,anti-patterns}.md` + dirs |
| Topology | Standalone single repo (§10 N/A) |
| Nav-maps | ❌ not generated (`/map` pending) |
| Commit gate | ❌ none installed (`core.hooksPath` unset) |
| Git | synced to `ddf8ada` (was 1 behind; clean ff). `pa-base.md` + new `pa.md` + `docs/pa/**` + `docs/{deep-dives,changes}/` UNTRACKED; **nothing committed this session** (no auth given) |

## Project state (verified at init)

- Phase 1 DONE; build-order step 1 (dynamic channels + theme + storage env) DONE.
- Recording model DESIGNED only; store has just the `samples` table.
- **Phase 2 UNBLOCKED:** real 15-column Enbridge DAQ CSVs arrived (`ddf8ada`); format decoded in
  status.md. Format mechanism (no-code mapping + compute) still unbuilt.
- **Parallel stack discovered** (Influx/Grafana PoC) — see the fork above.
- Phases 3–4 not started.
- See `docs/pa/status.md` for the full board + the MAJOR FORK + design↔code deltas + debts.

## What was done this session

Instantiated `pa-base v1` into cementer's PA contract: read the base + the real code state (not the
README narrative), filled all ~32 base slots for a Go/Vite/SQLite/Pi project, wrote the project axioms
(raw≠live≠recording, no-code DAQ, standalone island, layered durability, segments-as-markers), and
created the live scaffolding. Verified state directly: store schema, embed directive, git hooks, tests.

## Recovered-from anomalies

- None. (Init session; no dispatches, no leaks, no crashes.)

## Next priority

1. **Phase 2 — Intellisense `DaqFormat` preset + no-code mapping/compute layer.** Define the preset
   from the Enbridge CSVs (15-col `_NN_`, Excel-serial `_00_LOGTIME`, per-unit `_05/06_PRESS_*` +
   `_07/08_RATE_*`, stage totals `_11/12_*`, `_13_JOB_NUMBER`, `_14_MARKER`). Map columns → channels;
   compute layer for aggregates the pump doesn't emit. **No parser edits** (axiom #2). Verify
   `parser.DefaultConfig` (synthetic 4-channel) is replaced by config, not code.
2. Consider `/forge go` for a dedicated `cementer-go-engineer` before substantial Phase 2 source churn.
3. Pick off a debt: install the commit gate; resolve the `docs/plan` reference; README Go-version.

## File-modification inventory (this session)

- **NEW:** `pa.md`, `docs/pa/status.md`, `docs/pa/hand-off.md`, `docs/pa/user-voice.md`,
  `docs/pa/design-insights.md`, `docs/pa/changelog.md`, `docs/pa/anti-patterns.md`,
  `docs/pa/archive/.gitkeep`, `docs/pa/briefs/.gitkeep`, `docs/deep-dives/.gitkeep`,
  `docs/changes/.gitkeep`.
- **UNTRACKED (pre-existing):** `pa-base.md` (dropped in before this session).
- **No source code touched.**
