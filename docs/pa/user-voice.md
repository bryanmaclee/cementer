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

---

## Session 2 — 2026-06-13

> read the "pa.md" file for reference

> Yes, read the list you just provided, but first read the README.md

> I need assistance with updating the pi 4b with Bryan's update, so I can test the hardware on my end appropiately.

_(Decisions via AskUserQuestion: "Bryan's update" = the Go cementer binary (stack A); build path =
install Go on this machine + cross-compile to ARM64.)_

> we are simulating a live feed here. it is not raw data from daq. it is raw data from esp32 csv file. just want to make sure we are clear. Once pi is tested and good with Go and SQLite, I will be able to take the setup to the daq unit.

> good, I would like you to stand by on logging until we have verified the results of the bench top test here. I don't want to needlessly convolute the logs with maybes.

> Excellent! Now, I will need to test the raw data, but first, I will be using a rs-232 to usb adapter to go into the pi for the raw data feed from daq. so how do I need to change my code to do this and test it before driving miles away.

> okay, I want to try the send_csv.py test again to esp32 to pi usting a tty to usb adapter. I believe that is the closest I can come to the actual test on daq unit

> I have the rs-232 adapter in hand. the only hiccup might be that all this work has been done on my desktop in my garage and I will be using the laptop in the field. so how can I ensure this progress and context are saved, in case I am needing assistance there?

> go ahead and push it

_(Bench-top validation verified on both serial paths before logging, per the hold above. Cross-machine
concern → wrote the FIELD RUNBOOK into hand-off.md + committed/pushed so the field laptop picks up via
`git pull`.)_
