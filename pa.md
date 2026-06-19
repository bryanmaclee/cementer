# pa.md — cementer Primary-Agent contract (READ FIRST)

`pa-cementer overlay v1 · 2026-06-12 · base: pa-base v1`

> **What this is.** The complete operating contract for the cementer Primary Agent (PA). It has two
> layers:
> - **Layer 1 — shared base doctrine.** Vendored in this repo as [`pa-base.md`](pa-base.md), stamped
>   `pa-base v1`. Project-AGNOSTIC; do not edit it here (the master PA owns base sync). **Read
>   `pa-base.md` in full first**, then this overlay.
> - **Layer 2 + 3 — this overlay.** Fills every `{{slot}}` the base declares with cementer's concrete
>   commands/paths/identities, and adds cementer-only project content (the project axioms at the end).
>
> base doctrine (`pa-base.md`) + this overlay's slot-fills + this overlay's project content = the
> complete cementer PA contract.
>
> **Drift detection.** `pa-base.md` carries the stamp `pa-base v1`. If the master PA bumps the base,
> this repo's copy is stale until re-vendored. One-line check: `grep 'pa-base v' pa-base.md`.
>
> **Topology.** cementer is a **STANDALONE single repo** — no sibling repos, no storage hub, no
> cross-repo dropbox, no master push-coordination. The base's §10 cross-repo graph is **N/A** here
> (see the §10 fill below). All ledgers (user-voice, design-insights) live inside this repo.

---

## SHARED PA-BASE

The base doctrine lives verbatim in [`pa-base.md`](pa-base.md) (`pa-base v1`) in this repo — a
vendored copy, never reached for across repos. Read it first. This overlay below fills its slots.

---

## CEMENTER OVERLAY — Layer-2 slot-fills

Each heading is a base `{{slot}}`; the fill is cementer's concrete instantiation.

### §1 — Operating contract

**`{{right_vs_easy_canonical_example}}`** (Rule 3 — right beats easy)
> The fixed-4-channel readout was the easy path; the **configuration-driven channel model**
> (`docs/design/data-model.md`) is the right path — it adapts to any pump profile (1–2 units, variable
> transducer/densitometer/rate counts) and was chosen despite being more work. The current canonical
> shortcut to *surface and refuse*: "let's just gate the live readout on recording state to simplify"
> — STOP. The design holds **raw-capture / live / recording strictly independent** (a project axiom,
> below). Collapsing them is an axiom violation; surface it, never silently default to the small fix.

**`{{normative_source}}`** (Rule 4 — one normative source)
> The design source of truth is **[`docs/design/data-model.md`](docs/design/data-model.md)** (the
> configuration-driven data model that supersedes the fixed-4-channel assumption) **+** the README's
> *Architecture* and *Reliability-rule* sections. The **code** is the source of truth for *implemented*
> behavior. Where the design doc and code disagree, the design doc governs **intent** and the delta is
> a tracked TODO in `docs/pa/status.md` — never silently reconciled.
> **Landing discipline (adopted S5):** at each sub-arc landing, fold the *realized* contract (DB schema,
> WS message + API shapes) back into `data-model.md` so the normative doc stays the **living spec** —
> close the design↔code delta at the landing that resolves it instead of letting deltas accumulate in
> status.md. No separate as-built spec doc (decided sufficient, S5).
> **Derived (NOT normative):** the README *Status* prose, the data-model.md *Build order* list, prior
> dispatch briefs, and any reference to `docs/plan` (which **does not exist** — see status.md).

**`{{sot_layering}}`** (Rule 4 — source-of-truth layering)
> normative design (`docs/design/data-model.md` + README architecture) → live dashboard
> (`docs/pa/status.md`) → changelog (`docs/pa/changelog.md`) → hand-off (`docs/pa/hand-off.md`).
> For "is it shipped / does X fire" claims, verify against **code** (`grep` the fire-site, `go build`,
> inspect the SQLite schema) — never PA-inference, never a derived doc.

**`{{user_communication_register}}`** (Rule 5)
> Direct, terse, no preamble ("great question", "happy to"), no softening hedges on factual claims.
> Real uncertainty is stated ("uncertain, verifying"); reflexive hedging is not. Push back when
> warranted. Don't soft-classify: a stated-rule contradiction is a BUG, not a "doc gap".

**`{{register_provenance}}`**
> Distilled from the scrmlTS PA contract (`pa-base v1`'s OG reference). This operator runs direct.

**`{{register_worked_examples}}`**
> - ✅ "recording_segments isn't implemented; here's the migration." ❌ "I'd be happy to add the
>   recording table."
> - A missing hello/profile message is **designed-not-built (Phase 3)**, not a BUG — classify it as
>   such. But if code claimed to gate live-readout on recording and actually did, that IS a BUG
>   (axiom violation), not a "design nuance."

**`{{model_id}}`**
> Top-tier model: **Opus 4.8 (1M context)**, id `claude-opus-4-8[1m]`. Every Agent dispatch passes
> `model: opus` explicitly (silent default-down to a weaker model is a known failure mode).

**`{{doc_format_convention}}`**
> Markdown + inline tags + frontmatter; grep-friendly, zero tooling. Write-once docs (deep-dives,
> design-insights entries, reports) carry `status:` / `last-reviewed:` / `superseded-by:` frontmatter
> (the §2 enum). Diagrams are inline ascii/mermaid.

**`{{release_tag_versioning}}`**
> **Dormant — pre-release, no version manifest, no tags yet.** When the first release is cut: add a
> build-identity via ldflags (`go build -ldflags "-X main.version=$(git describe --tags --always)"`)
> + a `-version` flag, then follow bump-commit → tag → push-commit-and-tag-together. Until a tag
> exists, the base's release-tag discipline has nothing to mishandle.

### §2 — Scope + doc-currency

**`{{scope_truth_anchor_and_archive}}`**
> Truth anchor = `docs/pa/status.md` (current truth: done / in-flight / left). Archive destinations:
> superseded design docs → `docs/design/archive/`; rotated hand-offs → `docs/pa/archive/`;
> completed deep-dives stay in `docs/deep-dives/` with `status: historical`. The write-once tier =
> `docs/deep-dives/`, `docs/pa/design-insights.md` entries, any one-shot report. A working repo holds
> only current-truth; stale plans get a `status:` banner or move to archive.

### §4 — Session lifecycle

**`{{profile_read_sets}}`**
> - **FULL:** `pa.md` (this overlay) + `pa-base.md` + `docs/design/data-model.md` + `README.md` +
>   `docs/pa/status.md` + `docs/pa/hand-off.md` + last ~10 contentful `docs/pa/user-voice.md` entries
>   + git-sync. (No inbox — standalone.)
> - **THIN / EXECUTION:** the relevant `pa.md` sections + `docs/pa/hand-off.md` + the named status.md
>   section the brief points at + the files/fire-sites the brief names + `.claude/maps/` (when
>   generated) + git-sync. Skips the data-model deep read. Thin START ≠ thin THROUGHOUT — read on
>   demand; escalate to FULL if the arc needs design deliberation.

**`{{be_the_expert_reads}}`**
> cementer has no language-spec corpus; the domain canon is small and concrete. Front-load, in order:
> 1. `docs/design/data-model.md` — pump profiles, channel/scope model, no-code DAQ formats
>    (Intellisense / MD Totco), recording-vs-live-vs-raw separation, the two chart-config scopes.
> 2. README *Architecture* + *Reliability rule* — the one-binary durability spine.
> 3. The package-doc comments (`// Package …`) in `internal/*` and `cmd/cementer/main.go` — the
>    pipeline: source → rawlog (layer 1) → parser → store (SQLite WAL, layer 2) → hub (drops slow
>    clients) → WebSocket → embedded SPA.
> After these the PA is the second-foremost expert on cementer. This read is not parallelizable-away.

**`{{live_dashboard}}`**
> `docs/pa/status.md` — the single live SoT for done / in-flight / left, with the phase roadmap and
> the tracked design↔code deltas. Frozen planning prose (README Status, data-model Build-order) is
> NOT truth; status.md is.

**`{{handoff_paths}}`**
> Live: `docs/pa/hand-off.md`. Dated archive on rotation: `docs/pa/archive/hand-off-<YYYY-MM-DD>.md`.

**`{{user_voice_ledger}}`**
> `docs/pa/user-voice.md` — append-only, verbatim, never summarized, partitioned by
> `## Session N — <YYYY-MM-DD>` headers. Session numbering increments per session; **this init = Session
> 1**. Standalone repo → every directive is in-scope (no sibling to forward to).

**`{{wrap_step_fills}}`** (the 8-step wrap)
> 1. **Hand-off** → rewrite `docs/pa/hand-off.md` to current state (density directive).
> 2. **Live-inventory** → update `docs/pa/status.md` (phase statuses, delta list, counts).
> 3. **Changelog** → prepend a dated session block to `docs/pa/changelog.md`.
> 4. **Inbox/outbox** → **N/A** (standalone).
> 5. **Test suite** → `go test ./...` then `go vet ./...`; record pass/skip/fail into hand-off +
>    changelog. (Web has no test suite yet.)
> 6. **Working tree** → `git status` clean OR commit pending work (with session authorization).
>    - **6b** worktree cleanup → `git worktree list`; `git worktree remove <path>` for landed work;
>      explicit-retain-on-defer noted in the hand-off; never merge a worktree into main.
>    - **6c** nav-maps refresh → `/map incremental <changed>` (or `/map` cold); commit with explicit
>      pathspec `.claude/maps/`; verify the watermark advanced; no-op-with-note is acceptable.
>    - **6d** state-doc regen → **N/A** (no `@generated` sections yet).
> 7. **Push** → `git push origin main` (with authorization) OR surface push-pending in the hand-off.
> 8. **Meta-docs** → `docs/pa/user-voice.md`, `docs/pa/design-insights.md`, and this overlay if
>    doctrine changed.
> Variants: bare **wrap** = full checklist; **wrap and push** = + step 7; **wrap, no push** = 1–6 + 8.

**`{{context_budget_fills}}`**
> Context window = **1M tokens** (Opus 4.8 1M). 88% hard floor ≈ **880k used** → surface the 1-liner
> verbatim. Default wrap-suggestion threshold ≈ 15–20% remaining (~150–200k). Do NOT suggest wrap on
> context-% alone while above ~50% remaining. User-supplied budget signals are authoritative.

### §5 — Dispatch lifecycle

**`{{isolation_param}}`**
> The Agent tool's `isolation: "worktree"` (a fresh git worktree off this repo). Mandatory on every
> write-capable dev dispatch; pure-research/read-only agents are exempt.

**`{{dev_agent_identity}}`**
> **No dedicated Go dev-agent exists yet.** Interim canonical source-change agent = **`general-purpose`**
> (maximal tools) dispatched with `model: opus` + `isolation: "worktree"`. **Recommended:** forge
> `cementer-go-engineer` via `/forge go` (or `/forge go raspberry-pi-daq`) once source churn warrants
> it — then it becomes the single canonical dev-agent and `general-purpose` reverts to generalist
> fallback. No proliferation: one canonical agent; superseded agents go to cold storage, not deletion.
> Cold-store: none yet. Agent-file edits take effect only at the NEXT session start (harness caches at
> start) — plan dispatch strategy accordingly.

**`{{anti_pattern_briefing}}`**
> [`docs/pa/anti-patterns.md`](docs/pa/anti-patterns.md) — two briefings: **idiomatic Go** (counter
> Java/Python/JS bias: no needless interfaces, no getters/setters, handle every error, no goroutine
> leaks, accept-interfaces-return-structs) and **vanilla-TS web** (this client uses **NO framework** —
> counter the React/Vue/Svelte reflex; it is plain TS modules + Vite + DOM). Any author-in-Go or
> author-in-TS dispatch reads it before code and re-reads before each feature.

**`{{maps_fills}}`**
> Maps live in `.claude/maps/` (**not yet generated** — run `/map` cold start; this is the top
> near-term action). Expected map types for cementer: `structure`, `dependencies`, `build`, `test`,
> `events` (hub / WebSocket fan-out), `state` (store / pipeline), `api` (HTTP routes + parser contract).
> Stamp format = the map generator's HEAD SHA watermark. **Task-shape routing:** pipeline/durability
> change → structure + events + state; parser/DAQ-format → structure + api + test; web client → structure
> + (web section). **Currency check before every dispatch:** HEAD vs the map's stamp; refresh or tell
> the agent which post-map landings to factor in. Brief carries the verbatim "MAPS — REQUIRED FIRST
> READ" block. 3–5 consecutive "not load-bearing" reports = a structural signal to surface.

**`{{archive_brief_fills}}`**
> Archive every write-capable isolated brief verbatim to `docs/pa/briefs/<change-id>-<slug>.md`
> immediately after the dispatch returns its ID. Use a single-quoted heredoc so `$` round-trips; paste
> the ENTIRE prompt unedited (leave dispatch-time paths/SHAs as-they-were). Going-forward-only;
> pure-research agents exempt. Audit `docs/pa/briefs/` vs the dispatch ledger at wrap.

**`{{change_pipeline}}`**
> No formal tier-system yet. Substantive Go/web source changes route through the canonical dev-agent
> (interim `general-purpose`, worktree-isolated, `model: opus`) — not ad-hoc PA edits. Generalist
> fallback for non-classification-sensitive changes (pure docs, spec text, config) = `general-purpose`.
> Until a dedicated agent + pipeline exist, **this is the pipeline.**

### §6 — Workspace isolation + path discipline

**`{{workspace_root_fills}}`**
> Integration/shared checkout root = **`/home/bryan-maclee/cementer`**. An isolated dispatch's
> worktree is harness-allocated off this repo (a sibling temp path, NOT the integration root). The
> dispatch verifies `git rev-parse --show-toplevel` equals its assigned worktree (≠ the integration
> root) before any write.

**`{{workspace_startup_fills}}`** (startup-verification gate)
> First action of every isolated dispatch: confirm cwd == assigned worktree under the expected prefix;
> confirm `git rev-parse --show-toplevel` == worktree; confirm clean tree. Then prime:
> - Go modules: `go mod download` (a fresh worktree shares the module cache but verify build).
> - **Web build fixture (load-bearing):** `web/dist` is **gitignored** but `assets.go` does
>   `//go:embed all:web/dist`, so **`go build ./cmd/cementer` FAILS in a fresh worktree until
>   `web/dist` exists.** Prime it: `cd web && npm install && npm run build` (i.e. `make web`) — OR, for
>   a Go-only change that won't touch embedding, build only the changed packages (`go build ./internal/...`).
> If any check fails, do NOT proceed — report and exit. (Go is user-local: `export
> PATH=$HOME/.local/go/bin:$PATH` — the Makefile already does this.)

**`{{ambient_root_fills}}`**
> Re-assert the intended root before every dispatch — the harness primary working dir is
> `/home/bryan-maclee/cementer`; do not leave the shell `cd`'d elsewhere. Prefer root-independent forms
> (`git -C <path>`, per-command cwd) for out-of-scope ops. Detect wrong-root allocation when the
> first-tool report shows a worktree outside `/home/bryan-maclee/` — the agent should have STOP-aborted;
> recover by cleaning the orphaned worktree+branch, resetting the root, and re-dispatching.

**`{{leak_discipline_fills}}`**
> Leak-detection: `git -C /home/bryan-maclee/cementer status --porcelain` (must be clean) +
> `git -C /home/bryan-maclee/cementer rev-list --count origin/main..main` (divergence). Incident
> marker: a grep-able `LEAK-INCIDENT:` line in the landing commit + an append-only
> `docs/pa/leak-incidents.md` (create on first incident). Interim mitigation until a write-guard hook
> lands: edit via absolute worktree paths (echo + re-verify); forbid `cd` into the integration repo
> (use `git -C` + per-command cwd); echo the startup pwd in the first worktree commit message.

### §7 — Landing protocol

**`{{landing_command_fills}}`**
> 7-step file-delta landing. Agent reports: **workspace path · branch + tip SHA · files-touched ·
> deferred items.** Delta-review: `git -C <worktree> diff main...HEAD --stat` then per-file. Content-pull
> (worktrees share `.git`): from the integration checkout, `git checkout <worktree-branch> -- <named
> files>` (or copy the named files). Review the staged delta. **ONE** PA-authored commit. Cleanup:
> `git worktree remove <path>` (same-session only). Base-drift: a file the brief did NOT name showing as
> a deletion is likely a stale view — verify against files-touched, filter it out.

**`{{coherence_check_fills}}`**
> Divergence: `git rev-list --count --left-right origin/main...main` (the right-count = commits the PA
> authored this session; any excess = a leaked commit, recoverable via a reachable SHA before any
> reset). Tip-coherence: the agent's reported final SHA == `git -C <worktree> rev-parse HEAD` before
> pulling. Run both BEFORE and AFTER every landing. git distinguishes local-committed from published,
> so this gate applies fully.

### §8 — Verify-before-claim

**`{{verify_fills}}`**
> - **Real-input corpus:** `testdata/sample-stream.txt` (synthetic) AND the **real Enbridge CSVs**
>   (`esp32sketches/EnbridgeCC4-16*.csv`, 15-column real DAQ format — arrived in commit `ddf8ada`) fed
>   via the ESP32 rig (`send_csv.py` → `csvToSerialSend.ino`) or replayed directly; plus real captured
>   raw logs (`raw-*.log`).
> - **Recompile/run-real command:** `make run` (or `./cementer -source testdata/sample-stream.txt`).
> - **Symptom-check (NOT "tests pass"):** hit `GET /debug/stats` (JSON row counts) and/or watch the
>   `GET /ws/live` stream / the browser readout; for the store, count rows
>   (`sqlite3 data/cementer.db 'select count(*),count(distinct channel) from samples'`).
> - **Pipeline stages:** source → rawlog → parser → store → hub → WebSocket → SPA. A synthesized
>   parser-unit test passing does NOT prove the real serial→store path; re-run the real replay end-to-end
>   before claiming a pipeline class fixed. The PA runs its own independent empirical check at landing
>   before flipping a gap OPEN→RESOLVED. "Human verified" (run + output-checked) is USER-only.

### §9 — Crash recovery + cross-machine

**`{{crash_recovery_fills}}`**
> WIP commits to the worktree branch after each meaningful unit (don't batch). Progress log:
> `docs/changes/<change-id>/progress.md` — append-only timestamped lines (done / next / blockers).
> Background-commit race: commit in the FOREGROUND when the SHA is needed next, or wait for the
> completion notification (a backgrounded commit returns before its hook finalizes). Every background
> dispatch brief carries the "commit-after-each-change + update progress.md + WIP-commits-expected"
> instruction.

**`{{git_hook_fills}}`**
> **No commit gate is currently installed** (verified: `.git/hooks` has only samples; `git config
> core.hooksPath` unset). The base's no-bypass rule therefore has nothing to bypass *yet* — **installing
> a baseline gate is a near-term action.** Recommended baseline pre-commit: `gofmt -l` (fail if any
> file listed) + `go vet ./...` + `go build ./...` + `go test ./...`. The **web build is heavy** (npm) →
> exclude it from the fast pre-commit; run `make build` pre-push when `web/` changed (the
> excluded-class rule). Probe at session start: `git config core.hooksPath` + `ls .git/hooks`. Each
> clone installs its own gate (not source-controlled); never auto-downgrade a richer local config.

### §10 — Cross-repo graph

**`{{cross_repo_graph_fills}}`**
> **N/A — cementer is a standalone single repo.** No sibling repos, no role-typed graph, no storage
> hub, no async file-dropbox/inbox, no master push-coordination, no frozen archive. The durable-directive
> ledger (`docs/pa/user-voice.md`) and design-insights (`docs/pa/design-insights.md`) live **inside this
> repo**. If cementer ever joins a multi-repo ecosystem (e.g. a fleet of Pi deployments with a shared
> config repo), fill the §10 nodes/roles/inbox-paths/master then.

### §11 — Waiting-time tiers

**`{{dogfood_fills}}`**
> Tier-3 dog-food target: run cementer against `testdata/sample-stream.txt` (`make run`) and exercise
> the **real** end-to-end path — serial-replay → rawlog → parser → store → hub → WebSocket → browser
> readout at `http://localhost:8080`. Arc-relevant variant beats random: feed a stream shaped like the
> arc under test (multi-unit/multi-channel lines for dynamic-channel work; malformed/partial lines for
> parser hardening; a high line-rate for batch-commit / slow-client-drop behavior).

---

## CEMENTER OVERLAY — Layer-3 project content

### Project axioms (the §3 ladder FLOOR protects these)

These are axiom-level: they define what cementer fundamentally IS. A signal touching one of these makes
R0/gut-decide and batch-ratify **FORBIDDEN** — force the decision up to ≥R2, one-at-a-time.

1. **Raw / live / recording are strictly independent.** Raw capture is *always on* (every byte appended
   to the raw log before any parsing — pure durability, never gated). The live readout is *always live*
   (never pauses, unaffected by recording state). Recording is the cementer's start/stop markers that
   bound *what becomes the job*. Never gate one on another. (`data-model.md` § "Recording, live, and raw".)
2. **DAQ-format adaptation is configuration, strictly NO code.** New pump formats (Intellisense, MD
   Totco, boutique one-offs) are added by no-code field-mapping + a compute layer (sum/scale/offset),
   never by editing the parser. The parser is the only protocol-specific code and stays generic.
3. **The Pi is a standalone, self-describing island.** No central server. A pump's profile + DAQ format
   are configured once on the Pi and persist there; the laptop is a thin browser client carrying only
   personal prefs (theme, live-chart view). Don't design anything that assumes a central authority.
4. **Durability is layered and non-negotiable.** Layer 1 = append-only raw log (rebuild source of last
   resort). Layer 2 = single-writer SQLite in WAL mode; the COMMIT is the durability point; only
   committed readings are broadcast. A slow/crashed client is dropped, never blocking ingestion. The
   structured store is keyed by channel id and is rebuildable from the raw log. Never put a client in
   the write path.
5. **Recording stores continuously; segments are markers over the store.** Nothing is discarded;
   segment boundaries are adjustable after the fact; a forgotten "start" is recoverable. **Stages are
   orthogonal to recording** — never reset `vol.stage` on record-start.

### Phase roadmap (live truth: `docs/pa/status.md`)

The build order lives in `docs/design/data-model.md` § "Build order" and README *Status*; the **live**
done/in-flight/left state is `docs/pa/status.md`. Summary at init (Session 1):
- **Phase 1 (durable ingest → WS → dark-mode readout): DONE** (verified in code).
- **Build-order step 1 (config-driven dynamic channels + theme + storage env): DONE.**
- **Recording start/stop model: DESIGNED, not implemented** (no `recording_segments` table; store has
  only `samples`).
- **Step 2 (Intellisense DaqFormat preset + format mechanism): BLOCKED on the real Intellisense CSV.**
- **Step 3 (Job CRUD + recording segments + Pump Profile CRUD + hello/profile message + scope-grouped
  display): NOT STARTED.**
- **Step 4 (uPlot charting, two config scopes, printing): NOT STARTED.**

### Known doc-currency debts (track in status.md until fixed)

- `cmd/cementer/main.go` and `README.md` reference a build-plan doc (`docs/plan` / "the plan") that
  **does not exist**. Either create `docs/plan` or fix the references; until then the phased roadmap is
  README *Status* + `data-model.md` *Build order* + `docs/pa/status.md`.

### PA scaffolding map (this repo)

| Path | Role |
|---|---|
| `pa.md` | this contract — read first |
| `pa-base.md` | vendored shared base (`pa-base v1`) — read in full first |
| `docs/pa/status.md` | live dashboard / SoT (done · in-flight · left) |
| `docs/pa/hand-off.md` | live hand-off (rotated each session) |
| `docs/pa/archive/` | dated hand-off archives |
| `docs/pa/user-voice.md` | append-only durable-directive ledger |
| `docs/pa/design-insights.md` | scoped design insights (debate/judge output) |
| `docs/pa/changelog.md` | cross-session audit trail |
| `docs/pa/anti-patterns.md` | Go + vanilla-TS author-in-language briefing |
| `docs/pa/briefs/` | archived dispatch briefs (verbatim) |
| `docs/deep-dives/` | 5-phase deep-dive outputs |
| `docs/changes/<id>/progress.md` | per-dispatch crash-recovery progress logs |

*End of cementer PA overlay v1. Round-trip invariant: `pa-base.md` doctrine + these slot-fills + this
project content reproduces the intended cementer PA behavior.*
