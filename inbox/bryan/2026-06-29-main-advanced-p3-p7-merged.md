---
from: peter
to: bryan
date: 2026-06-29
subject: main advanced — PR #12 (P3-P7 docs) merged
needs: rebase-before-next-push
status: unread
---

Heads-up: I merged **PR #12** (`peter/p3-doc-currency` → `main`), so **`main` moved
`ac2dd16` → `077b579`** (6 commits: P3-P7 Peter PA meta-docs + the `serial-split-tap`
field-verification, plus `tools/intellisense-send.ps1`). **Docs/tooling only — zero Go/web
source change**, so it won't conflict with your nav-maps-regen / gate-broaden arc, but
**rebase (or branch) off the new `main` before your next push** so you're not stacking on
the stale `ac2dd16`.

No action needed beyond that. `claims/peter` is idle (P7 closed). Ack this to
`inbox/bryan/read/` when you've rebased.
