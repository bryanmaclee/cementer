---
from: peter
to: bryan
date: 2026-06-21
subject: repo ruleset is too broad for the multi-party workflow (exempt coord + allow feature-branch deletion)
needs: repo-config change (owner only)
status: open
---

Two ruleset rules block the multi-party workflow as designed. Found during P1 while
adopting the S6 machinery. Both are repo-owner config (your call):

1. **`coord` is caught by the "require pull request" rule.** Pushing directly to `coord`
   is rejected ("Changes must be made through a pull request"). But the design
   (`docs/deep-dives/multi-party-pa-orchestration-2026-06-21.md` + `coord` README) makes
   `coord` an **unprotected, push-direct** branch on purpose — low-latency coordination
   off `main`'s PR-flow. As-is, neither of us can push `coord`, so the handshake substrate
   is unusable. (My P1 onboarding commit `13c695a` is stuck behind this.)
   **Fix:** exempt `coord` from the ruleset (target `main` only, or add a `coord` bypass).

2. **"Restrict deletions" is hitting feature branches.** After PR #2 merged I couldn't
   delete `peter/p1-onboarding` ("Cannot delete this branch - repository rule violations"),
   so merged operator branches linger on the remote.
   **Fix:** scope deletion-restriction to `main`/`coord` only, or allow `peter/*` + `bryan/*`
   deletion (and/or enable "automatically delete head branches" on merge).

What works fine and should stay: PR-flow to a protected `main` (proven by PR #2 -> merge ->
`main` at `0a96095`), and feature-branch *creation* + push (the pre-push gate runs there).

Mirrored as GitHub issue (so this reaches you while `coord` is unpushable). Ack -> move to
`inbox/bryan/read/` once handled.
