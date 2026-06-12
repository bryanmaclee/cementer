# user-voice — durable-directive ledger

Append-only · verbatim · never summarized, paraphrased, or truncated · partitioned by session.
Standalone repo → every directive here is in-scope for cementer.

---

## Session 1 — 2026-06-12

> read pa-base.md and init workflow

_(Init session. The directive above is the verbatim instruction that bootstrapped the PA workflow.
Decisions captured via AskUserQuestion: scope = "full contract + scaffolding"; topology = "standalone
island".)_

> commit and push. run the dd. maps

> ratify A, retire B to dev bench. intellisense, phase 3/4

_(Ratified the storage+viz deep-dive: adopt stack A (Go single-binary + SQLite + custom uPlot UI);
retire stack B (Python/InfluxDB/Grafana) to a dev/diagnostic bench. The real 15-column `_NN_` DAQ
format is **Intellisense**. Retention/downsampling scoped to Phase 3/4.)_

> scope phase 2

_(Phase 2 decisions, via AskUserQuestion: D2 timestamp = embedded LOGTIME + server fallback; D4
live-serial fidelity = get a live-serial capture FIRST (build gated on it); dev agent = forge
`cementer-go-engineer` before the build.)_
