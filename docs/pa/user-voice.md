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

---

## Session 3 — 2026-06-14

> read pa.md and readme.md and start session

> leave it pending for now

_(Re: the uncommitted Session-2→3 hand-off rotation. Left uncommitted, not authorized to commit.)_

> field-trip for raw-data

> COM6; data interval 1/4 sec; Protocol 1;

> 9600 baud; parity enabled N; odd parity: n

> For simplicities sake, I would like to plug the raw-data USB adapter directly to the laptop and have you look at it directly. How can we make that work?

> Note: we are going to test two DAQs, 1st is the Totco system & 2nd will be the Intellisense

_(Pivot to direct-laptop serial capture — no Pi/Go/build. Confirmed Totco = COM6, 9600 8N1, Protocol 1,
1/4-sec interval. Two DAQs to capture: Totco then Intellisense → defines both Phase-2 presets.)_

> try again

> I am out of time for today, so I will have to try again next time

_(Session paused mid-diagnostic. BLOCKER: total silence on COM6 across all bauds → physical/electrical,
not settings. Exact next steps captured in hand-off.md.)_

---

## Session 4 — 2026-06-16

> read pa.md and start session

> This test is going to be raw-data from the Intellisense Dag unit through COM7

> pretty confident that it is 9600 baup, not confirmed though. I do not have the Data / Parity / Stop info available like the Totco. The unit is active now and should be sending data. We can start with 9600 baud and do a sweep if that doesn't work. It is a different cable/adapter. Let's try it!

> The truck is idle. I can try to pause and unpause the recording. hopefully that will give a header. Also, I can get the rate to move and I can toggle the density on/off which will jump to certain values and back down. but without having fluid on the unit, I cannot go through all tests, like pressure.

> I am going to put water on the pump and get the rest if possible. it will take some time

> I want to try to toggle the whole daq unit off/on and see if a header presents itself

> hold off, let me fill water, then test again through remaining list.

> this particular pumping unit only has 1 fluid pressure pump on it no backup density. there will be no way to test these. many of these systems do, so it is important to account for them in the future, but for now we can only do col 1, col 2, and maybe 9. we'll see if col 5 or col 6 respond. I believe col 2 would be for the total of col 5 & 6. We shall see what you come up with. The pump is still filling with water.

> go, water is on the pump unit.

> I will be pumping water and slowly close the valve to create pressure, then let off

> I didn't want to do too many things at once. preesure is good, so now we will do density. This should read +/- 8.33 by a tenth or so. I will to water rate just before I kick density on

> I don't think this pump has a flow meter for the water rate. but density on interface is reading 8.21

> bank it now, Totco isn't accessible today. we can move forward with project without totco for now

_(Captured the live **Intellisense** wire off a real unit via direct-laptop serial — **19200 8N1**, 14
columns, no header, `HH:MM:SS`-uptime timestamp. Empirically confirmed 8/14 columns by actuating the rig:
density (col 1 = 8.21, matched the unit's own interface), pressure (col 5 unit-1 0→1306, **col 2 aggregate
= sum(col 5, col 6)** proven per the user's hypothesis), rate (cols 3/7), volume totals (cols 4/12). The
6 flat columns are explained by this rig's hardware (1 unit, no backup density, no water flow meter) —
the format keeps them for multi-unit rigs. **D4 wire-contract gate CLOSED for Intellisense.** Totco
deferred (not accessible). Findings + Phase-2-ready preset banked in
`docs/changes/phase2-intellisense-daqformat/intellisense-wire-capture-2026-06-16.md`; raw captures under
`captures/`.)_

---

## Session 5 — 2026-06-18/19

> read pa.md and start session

> go

_(Via AskUserQuestion: proceed to **dispatch the Phase 2 daqformat build now**. Landed `83f036a`;
chose **commit + push, no wrap**.)_

> phase 3

_(Via AskUserQuestion: **scope all of Phase 3 first** before any build; **defer auth**.)_

> settle D2/D8 and 3a

_(D2 = store sole DB owner / single-conn synchronous CRUD; D8 = job header fields proposed set; then
dispatched 3a. Landed `cd71beb`; chose **commit + push, keep going**.)_

> yes, update that, enough for now.

_(Adopt the **landing discipline** — fold each realized contract into `data-model.md` at landing; no
separate as-built spec doc. Recorded in pa.md §normative-source + status.md. NOTE: the rest of that
message — flogeance/scrml "LSP for agents / FSP", naming dd+debate — was **misdirected** to this
instance; the user corrected: "that last bit ... was not for you ... talking to the wrong claude
instance." Flogeance/scrml is a SEPARATE ecosystem, out of scope for standalone cementer — not logged
here per the this-repo-only rule.)_

> 3b

_(Job fields stood as settled; dispatched 3b — jobs + recording segments. Landed `cf46ab3`; **commit +
push, keep going**.)_

> phase 4

_(Via AskUserQuestion: X-axis = **time**; traces = **all enabled channels auto-grouped by role**; live
view = **replace the readout with a rolling chart**; printing = "**both, pdf for file sharing**".)_

> server must have timed out, can you run it again

> some cross hairs show up. is that what Im suppose to be looking for?

> sure

> yes, i didnt scrutinize heavily, but yes. but somewhere along the way it starts stacking identical segments and looks very wrong

_(Demoing 4a. The "crosshairs/empty chart" = `make run` fed the synthetic stream under the now-default
`-format intellisense` → field-count guard dropped every line. The "stacking identical segments" = the
demo looping a single 12s capture; server data verified clean. PDF decision via AskUserQuestion:
**browser Save-as-PDF only** (single-binary ethos); chose **commit + push, keep going** for 4a.)_

> do both

_(Fix the uPlot time-axis ms→seconds bug + build a varied `make demo` stream. Landed `1465bd9`.)_

> then push when done.

> for headless browser I believe we have been using playwright

_(Unlocked headless verification — temp-installed `playwright@1.60.0`, screenshotted the live chart,
confirmed the seconds fix + varied data + no stacking. Saved to auto-memory.)_

> wrap

_(Full 8-step wrap: tests green, status/changelog/hand-off/user-voice updated, nav-maps regenerated,
committed + pushed.)_
