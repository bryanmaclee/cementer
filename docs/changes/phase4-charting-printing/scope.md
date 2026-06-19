---
status: current
last-reviewed: 2026-06-19
change-id: phase4-charting-printing
phase: 4
depends-on: Phase 3 landed (profiles + jobs + recording_segments; commit cf46ab3). Storage+viz DD ratified uPlot.
---

# Phase 4 scope — uPlot charting + two config scopes + printing

> The project centerpiece: the customizable, printable multi-line job chart. Authored Session 5 from
> the user's domain answers. No build dispatched yet.

## Goal

Turn the stored, channel-keyed samples + recording segments into the cement-job chart: a live rolling
real-time chart (replacing the value readout) and a historical per-job chart (the recorded segments),
both configurable, printable on paper, and exportable as a shareable PDF.

## User domain decisions (Session 5)

| # | Decision | Value |
|---|---|---|
| D-axis | Chart X-axis | **vs Time** |
| D-traces | Channels/axes | **ALL enabled channels, auto-grouped by role** (one scale per role/uom: pressure→psi, rate→bbl/min, density→ppg, volume→bbl; distinct color per channel) |
| D-live | Live view | **REPLACE the value readout with a live rolling chart** (keep current values in the chart legend) |
| D-print | Printing | **BOTH paper print AND a PDF for file sharing** — mechanism in D-pdf below |

## Boundary — what Phase 4 IS / ISN'T

| In scope | Out of scope |
|---|---|
| uPlot integration (bundled via npm/Vite, offline — ratified lib, not a framework) | Retention/downsampling-as-code (still **3c**) |
| Samples range/series query API (+ composite index) | Auth (deferred) |
| Live rolling chart (replaces readout) — all enabled channels, role-grouped axes | Multi-pump UI |
| Historical per-job chart (recorded segments, segment shading) | Live profile/recording-state WS push (Phase-3 deferred niceties) |
| Two config scopes: personal live-view config (localStorage) + company print template (bundled default + per-job overrides on the Pi) | |
| Printing (print-CSS, high-DPI) + a shareable PDF | |

## Current state (verified 2026-06-19)

- **web**: NO chart dep (`package.json` = typescript + vite only); `web/src/chart/` empty. uPlot to be added.
- **store**: NO historical query — only live stream + `Stats()`. `samples(ts_us, channel, value)` + `idx_samples_ts` only. Needs a range/series query + a `(channel, ts_us)` composite index for efficient per-channel extraction.
- **profile (3a)**: the active profile's ENABLED channels (role/scope/unitIndex/uom/decimals) drive trace grouping/axes/labels/colors. The chart reads the same profile frame the readout uses.
- **segments (3b)**: `recording_segments` (job_id, started_at_us, stopped_at_us) bound a job's data on the samples timeline (UnixMicro). The historical chart filters samples to these.
- **single-binary/CGO-free/offline-on-the-Pi** ethos governs every choice (esp. the PDF path — see D-pdf).

## Sub-arc decomposition

### 4a — Charting core (data + live + historical)
**Server:**
- `store` range/series query: `Series(fromUS, toUS int64, channels []string) (map[string][][2]float64, error)` (or a uPlot-aligned `[]time, map[channel][]value`). Single-conn read (the store is sole DB owner). Add `CREATE INDEX idx_samples_channel_ts ON samples(channel, ts_us)`.
- **Decimation cap:** if a requested range would return more than ~4k points/channel, bucket-average (or min/max-per-bucket so spikes survive) down to ~2–4k/channel for rendering. (Full downsampling-as-storage is 3c; this is a render cap only.)
- API: `GET /api/samples?from=&to=&channels=` (raw range) and `GET /api/jobs/{id}/series?channels=` (the job's segments + the samples within them, segment boundaries included for shading). Handlers call store methods only (D2).
**Client (vanilla TS + uPlot):**
- Add `uplot` dep; build `web/src/chart/` module. Bundle offline (no CDN).
- **Live rolling chart** (replaces `readout.ts`'s value grid): all enabled profile channels, **one uPlot scale per role/uom**, axes auto-assigned, distinct colors, a **legend showing each channel's latest value** (so current values stay visible). Rolling window (configurable, default e.g. 5 min) fed from the WS reading stream. Keep the status/connection footer + theme.
- **Historical job chart**: select a job → fetch `/api/jobs/{id}/series` → render vs time with **segment shading**; pan/zoom.
- **Personal live-view config (scope #1, localStorage)**: per-channel line on/off, rolling-window length, colors. Per-laptop; not synced (axiom #3).
- **Verify (4a):** live chart updates from the real replay; historical chart renders a recorded job with segment shading; axes grouped by role; only enabled channels appear; legend shows live values.

### 4b — Print template + per-job overrides + printing/PDF
- **Company default print template** (scope #2): bundled with the deploy (a JSON config / Go literal) — which channels, axis layout, title block, legend, page size. Change-controlled (updated via deploy, not casually editable).
- **Per-job print overrides**: stored with the job on the Pi (recommend a `job_print_config` table or a JSON column on `jobs`; engineer's call at 4b). The cementer tweaks per job; overrides persist with the job.
- **Printing**: a print-CSS view = the 3b job header (company/well/casing/job_type/location/cementer/date) + the chart at high-DPI. Browser print → paper.
- **PDF for file sharing (D-pdf — RESOLVED S5: browser Save-as-PDF only):** the print-CSS view doubles as the shareable PDF — the cementer prints to paper OR "Save as PDF" from the browser print dialog and shares the file. No Pi-side archival, no server render (preserves the single-static-binary/no-CGO guarantee).
- **Verify (4b):** print preview matches the template; Save-as-PDF produces a correct shareable file; a per-job override persists and re-renders.

## Decisions (engineering — PA, unless noted)

| # | Decision | Resolution |
|---|---|---|
| D-uplot | Chart lib | **uPlot** (ratified in the storage+viz DD); npm dep, Vite-bundled, offline; a library not a framework (anti-patterns Part B OK). |
| D-series | Historical data | New store range/series query + `(channel, ts_us)` composite index; render-time decimation cap (~2–4k pts/channel, min/max buckets to keep spikes). |
| D-live | Live view | Rolling uPlot replaces the value grid; latest values shown in the legend. |
| D-axes | Grouping | One uPlot scale per role/uom; axes auto-assigned; per-channel colors. |
| D-cfg1 | Personal live-view config | localStorage (scope #1), per-laptop. |
| D-cfg2 | Print template | Company default bundled + per-job overrides on the Pi (scope #2); storage shape decided at 4b. |
| D-pdf | PDF mechanism | **RESOLVED (S5): browser Save-as-PDF only.** The print-CSS view doubles as paper print and the shareable PDF; NO Pi-side archival, NO server render — preserves the single-static-binary/no-CGO guarantee. |
| D-default | Chart default range | Historical chart defaults to the job's recorded segments (data-model); live = rolling window. |

## Test + verify strategy (pa.md §8 — not "tests pass")

- Unit: store series range (boundaries, empty range, channel filter); decimation bucket correctness; print-config round-trip.
- E2E (real capture): live rolling chart updates from `captures/...-pressure.bin` replay; record a segment, then load the job's historical chart and confirm the segment renders + shades; axis grouping + enabled-only; Save-as-PDF spot-check.
- Static: gofmt/vet/test green; `make build` (uPlot bundled; CGO-free binary).

## Recommended dispatch (when authorized)

- **Agent:** `cementer-go-engineer`, opus, worktree, per sub-arc.
- **MAPS:** stale (`ee446c3`, pre-Phase-2/3); current-truth = `data-model.md` + this scope + source. Full regen due at wrap.
- **Anti-patterns:** Part A (Go) for store/api; **Part B (vanilla-TS)** for the chart — uPlot is a library, bundle offline, no framework reflex.
- **Landing discipline:** fold the realized series API + print-config contract into `data-model.md`.
- **Crash-recovery / brief archival:** progress.md + `docs/pa/briefs/phase4-<subarc>-<slug>.md`.

## Risks / unknowns

- **Axis clutter:** all-enabled + one-scale-per-role can crowd the chart (4 axes). Design sensible left/right placement + let the live-view config hide lines.
- **Range-query perf** on multi-hour jobs: the composite index + render cap handle it; true storage downsampling is 3c.
- **PDF ethos:** keep PDF client-side to preserve the single-static-binary/no-CGO guarantee (D-pdf).
- **"Replace readout":** ensure current numeric values remain glanceable (legend) — the readout's value-at-a-glance utility must survive the switch to a chart.
