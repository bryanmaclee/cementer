<!-- flobase:project:start (managed region — replaced by flobase assemble; do not edit by hand) -->

# cementer — flobase reconcile region

*Assembled by `/flobase` (2026-06-25), **reconcile mode**: cementer already runs a mature
scrml/flogence-lineage PA. This region does NOT replace it — it pins the executable **gate**, records
the **profile**, and points at the authoritative contract. flobase owns ONLY this fenced region.*

> **The PA contract is `pa.md` (overlay v2) + vendored `pa-base.md` (`pa-base v1`).** Read those first
> on a deliberate PA boot. Live SoT = `docs/pa/status.md`. Layout = `docs/pa/README.md`. This region is
> the flobase overlay on top — not a second source of truth.

## Profile (rehydrate from `.pa-base/profile`; don't re-derive)
- **STACK** — Go 1.26.4 (single-binary DAQ appliance; `internal/*` 9 pkgs + `cmd/cementer`) **+**
  TypeScript **vanilla, NO framework** (`web/src/*`; vite + `tsc` strict + uPlot; embedded via `assets.go`).
- **STAGE** — mid-flight (active multi-phase dev; P1→4b done; 3c deferred; serial-split-tap paused).
- **SCOPE** — medium (~10 pkgs + embedded client, ~8k LOC, **2 co-equal operators**).

## GATE — the executable source of truth (re-ground against this; never fake "done")
- **build:** `make build` (web `tsc && vite build` → embed → `CGO_ENABLED=0 go build ./cmd/cementer`);
  `make server` Go-only; `make pi` cross-compile arm64.
- **test:** the source-controlled pre-commit hook = **gofmt + go vet + build + `go test ./...`**
  (`core.hooksPath=scripts/git-hooks`, install per-clone via `make hooks`). Standalone: `go test ./...`.
  **NEVER `--no-verify` without explicit operator authorization** (global standing rule + project hook).
- **types/shape (per-stack, always-on):** `go vet ./...` · `cd web && npx tsc --noEmit`.
- **runtime/repro:** `make run` (synthetic 4-ch) · `make demo` (Intellisense replay → `:8080`) ·
  Playwright headless for web paint (temp-install `playwright@1.60.0`; Linux = cached browsers).

## Module set (TREE-SHAKE; default = minimum, scale by evidence)
- **init:** CORE + individualisation · stack-pack-go · stack-pack-ts · role-pa · continuity (`docs/pa` +
  `/wrap`) · maps (`.claude/maps` + `/map`) · vcs-drive (PR-flow + `.coord` + hooks).
- **runtime (available, not baked):** role-vpa · role-spa · role-dpa + deliberation — invoked on the
  event via `/vpa /spa /dpa /debate /forge`.
- **dropped:** stack-pack-scrml (not scrml) · role-cpa (single project) · dock (0 coverage).

## Project conventions (mined; don't fight them)
- **Web client is vanilla TS — NO framework** (`docs/pa/anti-patterns.md` Part B). Brief every web dispatch with this.
- **Pure-Go SQLite, CGO-free** (`modernc.org/sqlite`) — keeps the single-binary + cross-compile-to-Pi path.
- Pipeline: serial → `rawlog` → `daqformat` (config-driven format engine) → `store` (SQLite WAL,
  single-writer, sole DB owner) → `hub` → WS → client.
- **Axiom: raw-capture / live / recording are strictly independent** — recording is a marker, never gates ingest.
- **Canonical dev agent:** `cementer-go-engineer` (`isolation:"worktree"`, `model:opus`) for every source arc.
- **Normative design source:** `docs/design/data-model.md` + README architecture; fold realized
  contracts back in at each landing (landing discipline, S5). Code is truth for *implemented* behavior.

## Multi-party (dpa-012) + individualisation
- Two co-equal operators (Bryan + Peter) since S6; **PR-flow** to protected `main`, branch-per-operator;
  low-latency coord on the unprotected `coord` branch (`make coord` → `.coord/`).
- Meta-docs partitioned: per-operator (`hand-off-<op>.md`, `user-voice-<op>.md`, ids `B<n>`/`P<n>`) vs
  shared (`status.md` section-owned, `changelog.md`, `pa.md`). CODEOWNERS enforces per-operator ownership.
- **Vintage source (no parallel store):** voice + cadence come from `docs/pa/user-voice-<op>.md` +
  `pa.md` §user_communication_register; cross-project rules from `~/.claude/CLAUDE.md`. A directive about
  the OTHER operator's arc goes to their `coord` inbox, not acted on directly.
- **Voice (Bryan):** direct, terse, no preamble/hedging; push back when warranted; a stated-rule
  contradiction is a BUG, not a "doc gap".

<!-- flobase:project:end -->
