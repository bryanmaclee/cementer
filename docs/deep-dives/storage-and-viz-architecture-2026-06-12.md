---
status: current
last-reviewed: 2026-06-12
topic: storage-engine + visualization architecture
trigger: collaborator commit ddf8ada surfaced a parallel Python→InfluxDB→Grafana stack vs the repo's Go→SQLite→custom-UI stack
rung: R2 (one approach dominates on the decisive axes; USER rules)
---

# Deep-dive: storage engine + visualization architecture

**Question.** For a single Raspberry Pi 4B that is the *entire* compute of an offline oilfield
cement-pump DAQ "island" — durably capturing ~15 channels of ASCII serial at ~4–20 Hz over multi-hour
jobs (tens of thousands of rows/job, low-millions lifetime), showing a live readout, and producing the
**customizable, printable, multi-line job chart that is the stated product centerpiece** — and
field-maintainable for years with **no IT on site** — which architecture wins?

- **(A) Go single binary** — SQLite (WAL, pure-Go modernc, CGO-free) + embedded vanilla-TS UI + uPlot.
  *(the architecture this repo's code already implements.)*
- **(B) Multi-service** — Python ingest → InfluxDB 2.x → Grafana. *(the collaborator's working PoC in
  commit `ddf8ada`.)*
- **(C) Hybrid** — ship (A) as the product; keep (B) as a throwaway dev/diagnostic bench only.

## Phase 1 — scope

**In scope:** storage engine (SQLite vs InfluxDB), visualization/UI (custom uPlot vs Grafana),
deployment/ops model (single binary vs multi-service), fit to the project axioms (single binary,
offline island, layered durability, no-code DAQ format, the print centerpiece), ARM footprint, field
reliability, longevity, existing code investment.
**Out of scope:** the ESP32/serial hardware layer (it feeds *either* stack identically — the
`csvToSerialSend.ino` + `send_csv.py` rig is a reusable real-data injector regardless of outcome);
DAQ-format field-mapping; auth.
**Known going in:** the 5 project axioms (`pa.md` Layer-3); Phase 1 + config-driven display already
shipped on stack (A); the collaborator's note frames Influx/Grafana as a *data-flow proof* and hands
off "Customize UI and charting" — and leaves the DB choice **explicitly open** ("serve it whatever way
you feel best").

## Phase 2/3 — evidence (researched 2026-06-12, three parallel agents; sources inline)

### Storage: InfluxDB 2.x on an offline Pi — the durability/footprint/longevity case against

- **Compaction is a memory/OOM hazard on small boards.** InfluxDB's TSM compaction holds the
  shard being compacted in RAM; on constrained boards it OOM-loops while writes continue. Documented on
  a 1 GB Pi at *5 points/sec* ([GitHub #11339](https://github.com/influxdata/influxdb/issues/11339)),
  and — critically — **still happening on v2.7.10 with 32 GB RAM** as data volume grows
  ([GitHub #26524](https://github.com/influxdata/influxdb/issues/26524), opened 2025-06-16). The memory
  model is structural, not a tiny-Pi artifact. Mitigation knobs exist but need an admin — which we don't
  have in the field.
- **Power-loss corruption / data loss is a recurring field report.** Won't-start-after-power-loss,
  corrupt WAL skipped on replay, TSI index corruption "from sudden power loss"
  ([#12530](https://github.com/influxdata/influxdb/issues/12530),
  [#9657](https://github.com/influxdata/influxdb/issues/9657),
  [openHAB](https://community.openhab.org/t/influxdb-lost-data-after-emergency-shutdown/126427)).
  Power loss is a *real* field scenario for this device.
- **Flash wear + SSD requirement.** InfluxData's own sizing guidance says run on locally-attached SSD
  ≥1000 IOPS and that non-SSD "may not be able to recover from even small interruptions"
  ([influx hardware sizing](https://archive.docs.influxdata.com/influxdb/v1.1/guides/hardware_sizing/)).
  Compaction adds write amplification on top of raw inserts.
- **Longevity bomb.** OSS 2.x is a *frozen* branch (patched, but no published support policy); InfluxDB
  **3 Core is explicitly not a drop-in successor** (no migration path, Flux deprecated), and 3 Core's
  long-range query relies on a **compactor gated to Enterprise**, leaving a **~72-hour single-query
  cap** in the free tier ([InfluxData 3.0 alpha blog](https://www.influxdata.com/blog/influxdb3-open-source-public-alpha-jan-27/),
  [QuestDB analysis](https://questdb.com/blog/influxdb3-core-alpha-benchmarks-and-caveats/)). A product
  whose value is reviewing multi-hour/multi-day jobs and trending across jobs hits that ceiling. The
  1→2→3 line has been rewritten from scratch twice.
- **Genuine InfluxDB strengths (don't dismiss):** built-in downsampling/retention/continuous tasks,
  query ergonomics, native Grafana integration, and write throughput far above our need (one benchmark
  ~8.65× SQLite on batched inserts). Throughput is *not* our constraint; durability/footprint/longevity
  are.

### Visualization: Grafana for a *printable, company-standard, per-job* chart — the centerpiece mismatch

- **Reporting/PDF is Enterprise/Cloud-gated, not in OSS.** Scheduled reports + PDF generation require
  Grafana Enterprise/Cloud ([create-reports docs](https://grafana.com/docs/grafana/latest/dashboards/create-reports/);
  [community](https://community.grafana.com/t/how-to-export-dashboards-to-pdf-in-grafana-oss/110918)).
- **Grafana 13 *removed* the bundled image-renderer plugin** (the collaborator is on **13.x** — the
  worst version for this need); server-side render now needs a *separate* headless-Chromium service
  ([what's-new v13](https://grafana.com/docs/grafana/latest/whatsnew/whats-new-in-v13-0/)), which has a
  **broken ARM/Pi history open since 2018**
  ([Pi renderer issue](https://community.grafana.com/t/grafana-image-rendering-not-support-for-raspberry-pi-linux-arm64/77119)).
- **Not pixel-perfect / not paginated.** Browser-print renders only the first page; reports recommend
  ≤20 panels; layout limited to Grid/Simple. Grafana is positioned as an interactive *dashboard* tool,
  not a controlled *document* tool ([print limitation](https://community.grafana.com/t/not-able-to-print-entire-grafana-dashboard-it-will-only-contain-one-page-of-panels-or-dashboard/76771)).
  Templating is dashboard-variable-shaped, not "locked company template + bounded per-job overrides".
- **uPlot (the (A) choice):** ~50 KB, MIT, Canvas-2D, ~166k points in 25 ms, multi-axis/multi-series,
  syncs devicePixelRatio for crisp high-DPI print; the browser's own print engine gives exact page
  control ([uPlot](https://github.com/leeoniya/uPlot)). Cost: uPlot *deliberately* omits aggregation,
  the template/override layer, pagination, and print CSS — **you build the document layer yourself.**
- **Where Grafana genuinely wins:** speed-to-stand-up (live multi-channel readout in hours, zero
  chart-UI code), built-in alerting, ad-hoc exploration, and a solved offline-*kiosk* display path
  ([grafana-kiosk](https://github.com/grafana/grafana-kiosk)). If the deliverable were "a screen the
  foreman watches," Grafana would be pragmatic. It isn't — it's a controlled printed document.

### SQLite + single-binary at this scale — sound, with two honest caveats

- **~4 orders of magnitude below the crossover.** ~60–300 points/sec vs the "~200k points/sec" rule of
  thumb for leaving SQLite ([edge SQLite](https://www.sqliteforum.com/p/scaling-sqlite-on-edge-devices-iot)).
  Batched-in-transaction inserts hit ~23k/sec; per-row commits are the 85/sec trap
  ([squeezing SQLite](https://medium.com/@JasonWyatt/squeezing-performance-from-sqlite-insertions-971aff98eef2)).
  A ts index makes window queries cheap at low-millions rows.
- **Caveat 1 — durability pragmas are a deliberate choice.** WAL + `synchronous=NORMAL` is
  corruption-safe but can roll back the WAL-buffered tail on power loss; `synchronous=FULL` fsyncs each
  commit. **modernc/popular Go drivers may default to NORMAL — set it explicitly.** Pick a commit
  cadence = max tolerable loss window (e.g. 1 s → worst case loses ~1 s of the live job, never the DB)
  ([SQLite durability](https://www.agwa.name/blog/post/sqlite_durability), [pragma docs](https://sqlite.org/pragma.html)).
  *Note:* this device's **layer-1 raw log already makes whole-job loss recoverable regardless** (axiom 4)
  — the SQLite tail-loss window only affects the structured store, which is rebuildable from raw.
- **Caveat 2 — retention/downsampling are YOUR code.** No built-in TTL/rollups: scheduled
  `DELETE WHERE ts<?` + `VACUUM`, `strftime()` grouping for rollups, roll completed jobs into archive
  `.db` files. Simple, but real work a TSDB hands you ([modernc](https://pkg.go.dev/modernc.org/sqlite)).
- **Single-binary ops decisively simpler for no-IT recovery.** One process + systemd `Restart=`;
  update = copy one binary; rollback = swap the file; operator vocabulary = "power-cycle / restart."
  vs N services (DB + Grafana + Python + renderer) each a failure point that needs an admin
  ([golang+systemd](https://www.amazingcto.com/simplicity-of-golang-systemd-deployments/)). CGO-free
  modernc cross-compiles to a single static ARM artifact (`make pi` already does this).

## Trade-off matrix (weighted to THIS device's decisive axes)

| Axis (★ = decisive for this product) | (A) Go + SQLite + uPlot | (B) Python + InfluxDB + Grafana |
|---|---|---|
| ★ Offline durability under power loss | Strong (raw-log layer-1 + WAL; tail-loss bounded, store rebuildable) | Weak (documented corruption/loss reports; SSD-required) |
| ★ Printable company-standard per-job chart | Full control via uPlot + browser print (**you build it**) | Enterprise-gated PDF; renderer removed in v13; not paginated |
| ★ Single-binary field ops (no IT) | One file + systemd; trivial recover/update/rollback | 3–4 services; OOM hazard; needs an admin to diagnose |
| ★ ARM/Pi footprint | ~10 MB; tiny | ~1.2 GB+ Influx RAM + Grafana + Chromium renderer |
| ★ Longevity (multi-year bet) | SQLite = decade-stable format | 2.x frozen; 3.x no-migration; 72 h OSS query cap |
| Fit to existing code + axioms | Native (Phase 1 + data-model already on this stack) | Conflicts with single-binary + offline-island axioms |
| Speed to stand up | Slower — must write the chart/UI code | Fast — working PoC already exists |
| Built-in downsampling / retention | Build it yourself (caveat 2) | Built-in (genuine win) |
| Ad-hoc exploration / alerting | Build it | Built-in (genuine win) |
| Live on-screen readout | Already shipped (custom client) | Solved via Grafana kiosk |

## Phase 4 — recommendation

**Adopt (A); treat (B) as (C)'s dev bench.** On every axis that is *decisive for this specific
product* — offline power-loss durability, the printable company-standard per-job chart, single-binary
no-IT field ops, ARM footprint, multi-year longevity, and fit to the already-shipped code + the project
axioms — stack (A) wins, most of them decisively. Stack (B)'s real wins (speed-to-stand-up, built-in
downsampling/retention, ad-hoc exploration, alerting) are either not the centerpiece or are bounded
amounts of code on (A). The collaborator's framing agrees: Influx/Grafana proved the *data flow*; the
Go binary is the *productization*. Keep the ESP32 rig + the Influx/Grafana bench as a **dev/diagnostic
tool** (real-data injection + quick exploration) — that's the hybrid value, with **no claim on the
shipped product**.

**Engineering riders if (A) is ratified (from the research, fold into the build):**
1. Set `PRAGMA journal_mode=WAL; PRAGMA synchronous=FULL;` **explicitly** — don't trust the modernc
   default; pick a commit cadence = max tolerable store-loss window. (Raw-log layer-1 still backstops
   whole-job loss.)
2. Always batch inserts in a transaction (the store already uses a single batched writer — verify the
   cadence). Index `ts_us` (already present).
3. Plan retention/downsampling as explicit code (scheduled prune + `VACUUM`; `strftime()` rollups;
   per-job archive `.db`) — this is the real cost (A) imports vs a TSDB; scope it before Phase 4 charting.
4. Industrial SD/eMMC + periodic backup (file copy / `VACUUM INTO`); document the durability tradeoff.
5. The print artifact is uPlot-at-high-DPI + print CSS / headless-Chromium print of *our own* page —
   not a dashboard export. Design the company-default template + bounded per-job overrides as a
   first-class document, per `data-model.md` § "Two chart-config scopes".

## Phase 5 — debate?

**No full debate (R3) needed.** This is an **R2**: investigation surfaced one approach that dominates
on the decisive axes; no two approaches each win on a *different* decisive axis. The USER rules on
ratification. If the user wants to stress-test the "build the printable-chart UI ourselves" cost
specifically (the one place (B) has a real speed advantage), that single sub-question — *custom uPlot
report UI vs an embedded lightweight dashboard* — could be an R3 micro-debate; but the storage engine
and the overall architecture are not close.

## Open sub-questions for the user

1. **Ratify (A) + retire (B) to dev-bench?** (recommended) — or do you see a constraint the research
   missed (e.g. a client requirement for Grafana specifically)?
2. The DAQ format that arrived (`_NN_` 15-column) — is it "Intellisense", "MD Totco", or a one-off?
   Naming the preset unblocks Phase 2's no-code mapping work.
3. Should retention/downsampling be in the Phase-3/4 scope, or deferred (raw-log makes deferral safe)?
