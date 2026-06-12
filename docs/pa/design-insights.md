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

## INS-002 · For an offline no-IT appliance whose deliverable is a printed document, the single-binary wins the decisive axes — 2026-06-12

**Scope:** storage-engine + visualization choice for an offline, field-deployed edge device.
**Insight:** when the product is a single offline box with no on-site admin and its KEY deliverable is a
*controlled printed document* (not interactive exploration), a heavy TSDB + dashboard tool
(InfluxDB + Grafana) loses to a single self-contained binary + embedded store + purpose-built print UI
on every decisive axis: power-loss durability, single-binary recover/update/rollback, ARM footprint,
multi-year longevity, and document-control. The dashboard stack's real wins (speed-to-stand-up,
built-in downsampling, ad-hoc exploration, alerting) are off-centerpiece or bounded code. Corollary:
an **append-only raw-capture layer upstream of the structured store** makes the store engine's
durability tradeoff (e.g. SQLite WAL tail-loss on power cut) **non-fatal** — whole-job data is
rebuildable — which removes the strongest argument *for* a heavyweight TSDB.
**Why reusable:** the decision axis is "controlled-document + offline + no-IT", not "which DB is
faster". Throughput benchmarks are a distractor when the workload sits orders of magnitude below every
engine's ceiling. See `docs/deep-dives/storage-and-viz-architecture-2026-06-12.md` (R2; sourced).

---

_Future debate/judge outputs append above this line, newest-first._
