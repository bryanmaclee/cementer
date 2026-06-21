---
status: current
last-reviewed: 2026-06-21
decision: RESOLVED (S6) — git model = PR-flow (A), ruled directly (no debate); coordination layer (§6) + meta-doc partition (§7) ratified, to implement
topic: multi-party PA orchestration (two co-equal operators, one shared repo)
session: 6
---

# Deep-dive — Multi-party PA orchestration for cementer

> **DECISION (S6).** Two co-equal PA operators (Bryan + Peter) on one shared GitHub repo.
> The user **ruled the git-model fork directly: PR-flow (A)** — protected `main` +
> branch-per-operator + PR — declining the §8 debate. The coordination layer (§6) and the
> meta-doc partition (§7) are **ratified** (model-agnostic) and are the next implementation
> work. The §8 debate framing is retained below as the record of the considered alternative
> (trunk + handshake), not as a live fork. Baseline actions chosen S6: install the commit
> gate (done, `scripts/git-hooks/`) + settle Peter's GitHub access; `93011e6` held unpushed.

## 1. Why this exists (the topology reopening)

`pa.md` declares cementer a **"STANDALONE single repo … single operator"** and marks the
base's §10 cross-repo graph **N/A**. Peter Oliver joining as a co-committer/pusher breaks
that foundational assumption. Per the deliberation ladder this is **axiom-level** (the FLOOR
forbids gut-deciding it) — hence this R2 DD, requested verbatim by the user ("lets DD it").

The contract is **silent** on multi-party, but **not empty**: the base already carries
transferable shapes —
- **§4 same-repo concurrency catch:** "a sub-session OWNS a dedicated worktree + sub-hand-off
  and does NOT commit to the integration branch (the main PA lands), OR owns its arc's files
  exclusively." (single-operator parallelism — the seed of branch-per-actor.)
- **§9 cross-machine sync hygiene:** fetch + ahead/behind + rebase-if-behind + surface
  unpushed; the "paranoia protocol" for local-behind+dirty. (extends cleanly cross-operator.)
- **§10 async file-dropbox (inbox) + coordinated push via a master node.** (designed
  cross-repo; the SHAPE — create-only message files, ack-to-`read/` — transfers to
  cross-operator coordination in ONE repo.)

## 2. Scope lock (Phase 1 — locked with the user, S6)

**The precise question:** *How do two PA operators, each running their own Claude Code PA
against the same GitHub repo, coordinate work reaching `main` and the single-writer PA
meta-docs — without losing work, clobbering pushes, or corrupting shared state — and what
handshake opens/bounds a session?*

| Fork | User ruling (S6) | Consequence for the design |
|---|---|---|
| Git integration model | **Research both** (PR-flow vs trunk+handshake) | §8 — the one live fork; → debate |
| Operator roles | **Co-equal peers** | symmetric protocol; NO master/coordinator (rules out family D) |
| Handshake nature | **Autonomous PA-to-PA** (machine-readable in-repo) | §6 — a `coord/` substrate the PAs read+write, not a human board |
| Concurrency | **Mostly sequential / async** | **optimistic, not pessimistic** — verify-before-push + version stamp, NOT real-time locks |

**In scope:** git model · the session handshake · single-writer meta-doc ownership · conflict
avoidance · dispatch/landing under two parties.
**Out of scope:** the cementer *product* axioms (unchanged — this is workflow, not the
appliance) · CI/CD build-out · the scrml ecosystem · a master/coordinator topology (ruled out).
**Known:** `origin = git@github.com:bryanmaclee/cementer.git` (personal repo → PR-flow
available; Peter needs collaborator access or fork+PR) · **NO CI gate, not even a local
commit gate** (`core.hooksPath` unset) · today's flow = one PA lands ONE commit to `main` +
pushes.

## 3. What concretely BREAKS with two co-equal PAs

Enumerated against the **real** meta-docs, not in the abstract:

1. **`main` push contention.** Both push → non-fast-forward rejects, silent divergence, risk of
   a force-push stomping the other's commits. (git-level; the core problem.)
2. **Single-writer meta-docs collide.** `docs/pa/hand-off.md` is ONE live baton;
   `docs/pa/status.md` is THE single SoT; `docs/pa/changelog.md` and `docs/pa/user-voice.md`
   are append-only. Two PAs editing the same files every session = merge conflicts +
   *semantic* corruption (worse than textual: two "current truth" rewrites).
3. **Global session numbering.** "Session N" is one monotonic counter assuming one operator.
   Two operators interleaving → `Session 6` collisions; user-voice partitioning (`## Session N`)
   becomes ambiguous.
4. **The landing protocol assumes ONE lander.** "the main PA lands ONE integrator commit" — with
   two co-equal PAs, who lands what, and against which baseline?
5. **"Current truth" on session start.** Each PA must learn what the *other* operator did since
   its last session — today's git-sync only checks origin vs local, not "what did the human peer
   change and why."
6. **Worktree branch-name collisions.** Both use `.claude/worktrees/` (gitignored, per-clone —
   fine) but dev-agent branch names (`worktree-agent-<id>`) and operator branches could collide
   without a namespace.
7. **The handshake substrate itself is a shared write.** Any coordination file two PAs both edit
   re-introduces the very conflict it's meant to prevent → it must be append-only or
   per-operator-partitioned.

## 4. Approach families

- **A — PR-flow (GitHub-native).** Each operator's PA works on a namespaced branch
  (`bryan/<arc>`, `peter/<arc>`), opens a PR, merges to a **protected `main`**. Handshake =
  the PR + branch protection + CODEOWNERS. `main` is always reviewed + clean.
- **B — Trunk + autonomous handshake.** Both push `main` directly; a machine-readable `coord/`
  layer (claims + session ledger + push-intent) the PAs read/write; strict fetch-rebase-push.
  Closest to today; autonomous handshake is first-class.
- **C — Owner-partitioned meta-docs (a LAYER, not a standalone model).** Split the single-writer
  docs per-operator (per-operator hand-off, user-voice, session numbering) + CODEOWNERS; keep
  `status.md` a shared SoT with an append/section-tolerant structure. **Composes with A or B** —
  it removes problem #2/#3 under either git model.
- **D — Coordinator/master PA.** One lands; the other contributes inbound (base §10 master-node).
  **Ruled out by the co-equal ruling** — recorded only as a fallback if symmetry proves painful.

## 5. Trade-off matrix

| Axis | A — PR-flow | B — Trunk + handshake |
|---|---|---|
| New infra needed | **none** (GitHub gives PR + protection + CODEOWNERS free) | a custom `coord/` protocol to build + maintain |
| Red-`main` risk (no CI exists) | **low** — review gate + protected main | **higher** — direct-to-main without CI; prior art: TBD "demands automation" |
| Push-contention handling | git/GitHub (merge queue, non-ff blocked) | manual fetch-rebase-push; race on simultaneous push |
| Autonomous PA-to-PA handshake | partial — a PR is async+visible but PAs still need an in-flight signal | **first-class** — the `coord/` layer IS the handshake |
| Ceremony for a 2-person high-trust team | a PR per arc may feel heavy | minimal — commit + push as today |
| History cleanliness | **clean, reviewed, bisectable** | depends on rebase discipline |
| Matches 2026 multi-agent standard | **yes** (branch-per-actor + merge gate) | partial (shared branch is the anti-pattern that standard warns against) |
| Closeness to today's cementer flow | moderate (adds PR step) | **high** |
| Verify-before-claim fit (§8 PA doctrine) | the PR diff IS the review surface | review happens post-push, weaker gate |

**Reading:** A wins on safety / zero-infra / standard-alignment; B wins on low-ceremony /
closeness-to-today / autonomous-handshake-nativeness. They win on **different axes with no
dominant** → ladder rung **R3 (debate)**.

## 6. The autonomous coordination layer (model-agnostic — the "handshake")

This is what the user asked to "start with," and it works under EITHER git model. A small
append-only / per-operator-partitioned substrate under **`docs/pa/coord/`** (so problem #7 —
the handshake being a shared write — is structurally avoided):

```
docs/pa/coord/
  ledger.md                 — APPEND-ONLY session ledger; every PA appends one block on
                              session open AND close: {operator, session-id, branch, tip SHA,
                              arcs touched, push state}. The cross-operator extension of §9
                              sync hygiene — read it first on session start to learn what the
                              peer did since you last worked.
  claims/<operator>.md      — PER-OPERATOR (no shared write): "I am working arc X on branch Z,
                              push-intent at <SHA>." OPTIMISTIC, not a lock (async ⇒ OCC, not
                              leases): a claim is advisory; the other PA reads it to avoid the
                              same arc, but nothing blocks. Stale claims expire by session-close.
  inbox/<operator>/         — base §10 dropbox, repurposed cross-operator: create-only notices
                              ("landed X at <SHA> — rebase before your next push"); ack → read/.
```

**Handshake protocol (per session):**
1. **Open:** fetch; read `coord/ledger.md` tail + the peer's `claims/` + your `inbox/`; surface
   "peer did X since your S5; arc Y is claimed by peer." Append your open-block to the ledger.
2. **Claim:** write your `claims/<you>.md` (arc + branch + intent). Optimistic — if the peer
   already claims that arc, surface the overlap and pick another (async ⇒ rare).
3. **Land/push:** verify-before-push (fetch + rebase + the §7 coherence checks); on landing, drop
   a peer-inbox notice with the new `main` SHA.
4. **Close:** append your close-block (final SHA, arcs, push state) to the ledger; clear your claim.

All four files are either **append-only** (ledger) or **per-operator-owned** (claims/inbox) →
two PAs never write the same bytes → the coordination layer cannot itself conflict.

## 7. Meta-doc partition (family C — ratifiable under either model)

- **Per-operator session numbering:** prefix by operator — `B6`, `P3` — kills the global-counter
  collision (problem #3). `user-voice.md` → `user-voice-bryan.md` / `user-voice-peter.md` (each
  append-only, each operator's own).
- **Per-operator hand-off:** `hand-off-bryan.md` / `hand-off-peter.md` (each PA owns its baton;
  the ledger §6 is the *shared* cross-operator view). Removes problem #2 on the hand-off.
- **`status.md` stays the single shared SoT** (one source of truth is a feature, not a bug) but
  becomes **section-owned** (a "Bryan in-flight" / "Peter in-flight" block + a shared phase board)
  so edits rarely collide; last-writer reconciles the shared board.
- **CODEOWNERS** (`.github/CODEOWNERS`): route `web/**`, `internal/**`, the per-operator meta-docs
  to their owners; auto-request review on PRs. (Caveat from prior art: GitHub reads the PR's copy
  of CODEOWNERS, not main's.)
- **Branch namespace:** operator branches `@{operator}/<arc>`; dev-agent worktree branches stay
  `worktree-agent-<id>` (gitignored worktrees, per-clone — no collision).

## 8. Recommendation + Phase 5 (feed to debate)

**Ratify directly (model-agnostic, low-risk, high-value):** the §6 coordination layer + the §7
meta-doc partition. These solve problems #2/#3/#5/#6/#7 regardless of the git-model outcome, and
they ARE the "handshake" the user wants to start with.

**Send to debate (R3) — the one live fork (§5: two approaches, different axes, no dominant):**

> **`/debate` — git integration model for two co-equal async PA operators on a no-CI repo.**
> - **Participant 1 — PR-flow / GitHub-native advocate:** protected `main`, branch-per-actor +
>   PR + merge gate; argues safety + zero-infra + 2026-multi-agent-standard alignment; concedes
>   PR ceremony.
> - **Participant 2 — trunk-based + coordination-protocol advocate:** direct-to-`main` + the §6
>   handshake + fetch-rebase-push; argues low-ceremony + closeness-to-today + autonomous-first;
>   concedes the no-CI red-main risk.
> - **Challenge:** "two high-trust co-equal operators, async, NO CI gate — how does work reach
>   `main` safely with the least ceremony?"
> - Neither expert exists yet → **forge** a `pr-flow-expert` and a `trunk-based-expert` (per
>   `/forge`), or run `debate-curator` autonomously. Judge on the 6-dim rubric; record the scoped
>   insight to `design-insights.md`.

**PA lean (stated, non-binding — the user/debate rules):** **A (PR-flow), because cementer has
NO CI gate.** Trunk-based's headline cost in the prior art is that it "demands automation" to keep
`main` green; without even a local commit gate, direct-to-main is the riskier bet, and protected
`main` + review is the cheapest way to buy back that safety. A also extends cementer's EXISTING
discipline (dev work is ALREADY branch-isolated in worktrees; making the *operator's* landing a
branch→PR is a small step, not a new paradigm). The §6 handshake then rides ON TOP of PR-flow as
the in-flight-awareness layer PRs alone don't give.

## 9. Open questions for the user to rule on

1. **Ratify §6 + §7 now** (model-agnostic), and **run the §8 debate** for the git model? (PA
   recommendation.) Or rule the git model directly without a debate?
2. **Peter's GitHub access:** collaborator with push (enables branch-per-actor on one repo) vs
   fork+PR? (Determines whether A needs an org/protection setup.)
3. **Install a baseline commit gate now** (gofmt+vet+build+test) regardless of model — it lowers
   the red-`main` risk that most separates A from B. (Already a standing parked debt.)
4. **The unpushed `93011e6` (Phase 4b):** push it now as the clean single-operator baseline before
   any multi-party machinery lands? (PA recommends yes.)

## Prior art

| Source | Bears on | Takeaway |
|---|---|---|
| [DeployHQ: TBD vs Gitflow](https://www.deployhq.com/blog/trunk-based-development-vs-gitflow), [LaunchDarkly](https://launchdarkly.com/blog/git-branching-strategies-vs-trunk-based-development/) | §5 A vs B | GitHub Flow = main + short-lived PR branches, simplest for small teams; strict TBD "demands automation"/feature-flags to keep main green |
| [Augment: worktrees for parallel AI agents](https://www.augmentcode.com/guides/git-worktrees-parallel-ai-agent-execution), [MindStudio](https://www.mindstudio.ai/blog/git-worktrees-parallel-ai-coding-agents) | §3, §4, §8 | 2026 multi-agent standard: branch-per-agent + worktree + one shared spec + merge gate; "never commit to a shared branch"; worktrees stop file collisions, NOT semantic ones |
| [Wikipedia: Optimistic concurrency control](https://en.wikipedia.org/wiki/Optimistic_concurrency_control), [Azure lease vs ETag](https://oneuptime.com/blog/post/2026-02-16-how-to-implement-lease-management-for-blob-concurrency-control-in-azure-storage/view) | §2 (async), §6 | low-contention ⇒ OCC/ETags (verify-before-commit, version-stamp) beat pessimistic leases — "ETags simpler and more scalable"; matches "mostly sequential" |
| [GitHub Docs: CODEOWNERS](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners), [Aviator guide](https://www.aviator.co/blog/a-modern-guide-to-codeowners/) | §7 | per-path ownership + auto-review request; caveat: GitHub reads the PR's CODEOWNERS, not main's |

*End — DD output (Phase 4). Phase 5 = the §8 debate framing. Decision pending the user's §9 rulings.*
