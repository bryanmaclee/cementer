# design-insights — cementer

Scoped, reusable design insights — the write-once output of debates (`debate-judge`) and deep-dives.
Each entry: a scoped rule + the context that earned it. Cited as authority indefinitely; mark
`superseded-by:` when overtaken (the §2 same-landing discipline).

This is the **local** ledger (cementer is standalone — insights live in-repo, not in a shared hub).

---

## INS-001 · PA-base instantiation: verify code state, not commit-message narrative — 2026-06-12

**Scope:** instantiating a project overlay from `pa-base`.
**Insight:** when filling the live-status slots, the dashboard must be built from the **real code**
(SQLite schema, embed directives, installed git hooks, actual test files) — not from commit messages
or README *Status* prose. At cementer init this surfaced three facts the narrative hid: the recording
model is designed but the `recording_segments` table doesn't exist; a fresh worktree can't `go build`
because `web/dist` is gitignored yet `go:embed`-required; and no commit gate is installed at all.
**Why it's reusable:** the base's spine ("verify the authoritative real thing, not a narrative")
applies first and hardest at init — a status doc seeded from prose starts the corpus-ouroboros on day
one.

---

_No debates run yet. Future debate/judge outputs append above this line, newest-first._
