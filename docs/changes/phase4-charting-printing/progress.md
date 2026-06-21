# Phase 4a — charting core: progress

Append-only. Timestamped: done / next / blockers.

---

## 2026-06-19 — startup

- Verified worktree (`agent-ac1820ce112d2db48`), clean tree, Go 1.26.4, branch
  `worktree-agent-ac1820ce112d2db48`.
- Read normative sources: `docs/design/data-model.md` (two chart scopes; chart defaults to
  recorded segments), `docs/pa/anti-patterns.md` (both parts), the brief (authoritative — the
  phase4 scope.md did NOT exist yet; brief carries the 4a decisions inline; created the change dir).
- Read current code: `internal/store/{store,profile,jobs,recording}.go`, `internal/api/{api,jobs}.go`,
  `cmd/cementer/main.go`, `web/src/{main,types,ws,readout,controls,theme}.ts`, `styles.css`,
  build config. Channel roles confirmed: pressure(psi), rate(bbl/min), density(ppg), volume(bbl),
  meta. Decimation must preserve spikes (pressure peaks ~1306).
- Test baseline: api/daqformat/parser/store GREEN; root + cmd/cementer FAIL only on the missing
  `web/dist` embed fixture (expected pre-`make web`).

### NOTE — scope.md absent
`docs/changes/phase4-charting-printing/scope.md` did not exist. The brief is authoritative and
carries all 4a decisions inline (X=time, all enabled channels role-grouped, live chart replaces
readout w/ legend values, job chart over segments, decimation min+max per bucket, uPlot bundled
offline). Proceeding on the brief; surfacing this to the PA.

## 2026-06-19 — unit 1 done (store series)

- Added `idx_samples_channel_ts(channel, ts_us)` to `initSchema`.
- `internal/store/series.go`: `Series(from,to,channels,maxPerChannel) -> map[ch][][2]float64`
  with min/max-per-bucket decimation (spikes preserved, keeps real ts of the extreme); empty
  channels => all-in-range; empty request channel => empty non-nil slice; from>to errors; cap
  clamped to 20000. `JobSeries(jobID,...) -> (segs, series, ok, err)`: union span over segments,
  in-segment filter (gaps stay gaps), open segment extends to now, ok=false for unknown job (404).
- `series_test.go`: boundaries inclusive, channel filter, empty-channels-all-in-range, empty-range
  empty-slice, from>to error, decimation cap + spike preservation + time-order, no-decimation-under-cap,
  job-series unknown/in-segment-only/gap-between-segments. All store tests GREEN.
- Fix: clamp tail bucket index so emitted points stay within cap.

## 2026-06-19 — unit 2 done (API series routes)

- `internal/api/series.go`: `GET /api/samples?from&to&channels[&max]` -> `{series:{ch:[[ts,v]]}}`;
  `GET /api/jobs/{id}/series?channels[&max]` -> `{segments:[...],series:{ch:[[ts,v]]}}`. Validation:
  from/to required ints, from<=to, max>=0, default cap 4000. 404 unknown job. Registered in api.go.
- `series_test.go` (httptest): returns-series, empty-channels-all-in-range, validation matrix,
  job 404, job in-segment-only. All api tests GREEN.

## 2026-06-19 — web deps

- `npm install` + `npm install uplot` (uplot ^1.6.32, ESM + .d.ts + min.css). Bundled via Vite =>
  offline (no CDN), embedded in web/dist.

## 2026-06-19 — units 3-6 done (web charts + view shell + config)

- `web/src/chart/roles.ts`: scaleKey=uom, role rank, distinct color palette, orderScales.
- `web/src/chart/config.ts`: personal live-view config in localStorage (scope #1): hidden lines,
  per-channel colors, rolling windowMs. setHidden/setWindowMs persist.
- `web/src/chart/livechart.ts`: rolling uPlot, one scale per uom, axes alternate left/right,
  distinct per-channel color, custom always-on legend with LATEST value (readout glance utility
  preserved) + click-to-toggle (persisted), rolling ring buffer (default 5min, cap 4000), rAF-
  coalesced repaint. applyProfile rebuilds + preserves surviving channel data.
- `web/src/chart/jobchart.ts`: fetch /api/jobs/{id}/series, union-x merge (gaps as null),
  role-grouped axes, segment-shading plugin (drawClear hook, accent band per segment, open seg
  extends to now), pan/zoom drag. 404/empty states handled.
- `readout.ts` restructured into the SHELL: header (brand + status + theme + Live|Job tabs +
  window select) + controls host + view area (LiveChart default, JobChart hidden) + footer. Value
  grid removed; live view IS the rolling chart now.
- `controls.ts`: onActiveJob callback so Job History loads the active job; main.ts wires it.
- `styles.css`: chart/legend/tabs/window-select styles; value-grid CSS replaced. uPlot legibility
  in both themes.
- `css.d.ts`: ambient decl for side-effect CSS import.
- Web build GREEN (tsc + vite, 15 modules). Offline bundle VERIFIED: index.html refs only local
  /assets/, uPlot JS + CSS baked in, zero CDN/http refs. Full Go binary builds CGO-free; vet/fmt/test
  all green.

## 2026-06-19 — units 7-8 done (E2E + docs)

E2E against `captures/...-pressure.bin -format intellisense -replay-interval 40ms -replay-loop`:
- `/api/samples?channels=agg.pressure,unit1.pressure`: 891 pts each, max=1306 (matches brief ~1306).
- Decimation over realistic tight range: cap=50→50 pts, cap=200→181, cap=4000→1741 (all), spike
  1306 always kept, time-ordered. Fixed a degenerate case (range dwarfs data, e.g. from=0) by
  bucketing over the channel's DATA EXTENT; `[0,now]` max=100 now → 64 pts spike kept.
- Validation: from>to=400, missing from=400, empty range→empty slices, unknown job series=404.
- AXIOM #1 verified: rows climbed 7384→8034→8684 while hammering the read-only chart endpoints.
- Job flow: created+activated job, recorded ~4s segment, stopped; `/api/jobs/1/series` returned the
  segment [start,stop] + in-segment series (agg.pressure 0..1306, agg.rate 0.3..2.0).
- Embedded SPA served at :8090 references only local /assets/; the served JS bundle contains uPlot
  (valToPos/u-legend) — OFFLINE, no CDN. WS frames confirm 13 enabled channels w/ roles + RFC3339
  ts (Date.parse-able by LiveChart).
- Docs: data-model.md "Realized contract — Phase 4a" (index, Series/JobSeries, the two HTTP routes,
  the uPlot live/job charts, localStorage scope #1). README Status + layout (uPlot dep) updated.

VISUAL EYEBALL GATED: no headless browser available in this env (no chromium/chrome/puppeteer).
Could NOT render the page to visually confirm the chart paints. Verified instead: tsc strict clean,
vite build clean, offline bundle present in served JS/CSS, WS contract matches the chart's parser,
and all the data endpoints the charts call return correct shapes. The DOM/paint confirmation is the
one item I could not close here — needs a human/browser load.

Static final: gofmt clean, go vet clean, make build clean (CGO-free), go test ./... all green.

STATUS: complete for the server + data path + chart code; PARTIAL only on the human visual eyeball
(gated on a browser load — see above).

## 2026-06-19 — Phase-4a fix-up (time-units + varied demo)

Base: 1f65c13. Two demo-found issues.

### done — ISSUE 2 (varied demo asset)
- Built `testdata/intellisense-demo.txt`: concatenated the ten 19200-8N1 captures in
  chronological filename order, EXCLUDING the 9600 garbage (`...150051-9600-8N1.bin`).
  478 lines; 4 torn boot fragments (1/12/20/27 fields) dropped by the field-count guard.
- Per-channel min/max over clean lines: agg.pressure 0->1306, density.1 0->8.21,
  agg.rate 0->4.60, vol.job 0->43.3, unit1.pressure 0->1306, unit1.rate 0->4.60,
  vol.stage 0->43.3 MOVE; unit2.*, water.rate, density.2, vol.water.stage, job.number
  flat at 0. 7 moving channels — real multi-phase variety, not one ramp.

### next
- Point `make demo` at the new file; update README "Quick demo" wording (channels that move).
- ISSUE 1: feed uPlot SECONDS (live x = Date.parse/1000; job x = us/1e6; trim/window +
  shade-plugin in seconds; LiveConfig window in seconds).
- Build web + binary; gofmt/vet/test; E2E replay -> per-channel maxes via /api; axiom-#1 rows climb.

### done — ISSUE 1 (uPlot time-units = seconds, end-to-end)
- livechart.ts: push x = Date.parse(r.ts)/1000 (epoch seconds); ring xs in seconds;
  windowSec + DEFAULT_WINDOW_SEC; trim cutoff in seconds; setWindowSec().
- config.ts: LiveConfig.windowSec (seconds); setWindowSec().
- readout.ts: window selector values in seconds (mins*60); setWindowSec.
- jobchart.ts: union x = p[0]/1_000_000 (us->s); shade plugin startedAtUs/stoppedAtUs
  /1_000_000 before valToPos. Wire (RFC3339) + store (us) UNCHANGED.

### verify (post-change binary, real demo stream)
- gofmt -l: clean. go vet ./...: clean. go build ./...: ok. go test ./...: all GREEN
  (api/daqformat/parser/store) — no Go logic changed, nothing broke.
- make build: ok (web tsc-strict + vite, 71KB JS chunk; server binary 15.8MB).
- Offline bundle: no external http(s) URLs in web/dist; uPlot inlined in the JS chunk (no CDN).
- E2E demo replay (-replay-interval 50ms, /tmp/ce-demo, :8123) per-channel max via /api/samples:
  agg.pressure 1306, unit1.pressure 1306, density.1 8.21, agg.rate 4.60, unit1.rate 4.60,
  vol.job 43.30, vol.stage 43.30 MOVE; unit2.*/water.rate/density.2/vol.water.stage flat 0.
  => 7 moving channels, real variety (not one ramp). Matches offline analysis.
- Seconds proof: sample tsUs 1.7819e15 /1e6 = 1.7819e9 (correct era); built JS has /1e3 (live)
  and /1e6 (job); a 2026 RFC3339 ts -> 1.78e9 s (not 1.78e12 ms).
- Axiom #1: /debug/stats rows climbed 2743 -> 4810 -> 13195 -> 32799 while chart endpoints hit;
  /ws/live upgrade = HTTP 101. Ingestion independent of reads/clients.
- README Quick demo reworded for the multi-phase stream + accurate moving/flat channels.

### NEEDS HUMAN BROWSER EYEBALL
- The VISUAL time-axis tick LABELS (correct dates/times) need a real browser load — no headless
  browser here. Verified: the x VALUES are now seconds, tsc-strict passes, the bundle builds. The
  axis-label correctness (the actual fix) is confirmed by units, not by a rendered screenshot.

---

# Phase 4b — print template + per-job overrides + printing/PDF

## 2026-06-21 — startup

- Worktree `agent-a41de8b11395cc83f` (pwd
  `/home/bryan-maclee/cementer/.claude/worktrees/agent-a41de8b11395cc83f`), clean tree, Go 1.26.4,
  branch `worktree-agent-a41de8b11395cc83f`.
- Read maps (api/schema/state/structure/style — current at 1465bd9), anti-patterns (A+B),
  data-model.md "Two chart-config scopes" + 4a realized contract, scope.md §4b, and the relevant
  source (store jobs/recording/profile/series/store.go, api jobs/series/api.go, daqformat presets,
  web readout/controls/types/livechart/jobchart/roles/config/styles/main).
- Primed web/dist (`make web`). Test baseline GREEN: api/daqformat/parser/store all ok; root +
  cmd/cementer no test files.
- STORAGE-SHAPE DECISION: `print_config TEXT NOT NULL DEFAULT ''` JSON column on `jobs` (D-cfg2,
  recommended path). The store persists ONLY the raw override JSON (stays company-agnostic, mirrors
  how it stays format-agnostic for the profile vocab). The merge default+override -> effective lives
  in the API layer (which owns the company default). One column, no new table, no second writer.
- Plan: (1) two cosmetics, (2) company default Go literal in new internal/printcfg, (3) store
  print_config column + Get/SetJobPrintConfig, (4) GET/PUT /api/jobs/{id}/print-config, (5) web
  Report view + minimal override editor + window.print() + @media print CSS, (6) docs realized block.

## 2026-06-21 — cosmetics done

- (a) styles.css `.newjob-form[hidden]{display:none}` so the form is collapsed on load (grid
  was overriding UA [hidden]). Verified collapsed on load + expands on "+ New job…".
- (b) livechart.ts (x2) + jobchart.ts orderedChannels: filter `c.scope!=="meta" && c.role!=="meta"`
  so job.number (role:meta, scope:job) no longer charts a flat-0 trace; vol.job (role:volume) kept.

## 2026-06-21 — server done (printcfg + store + api)

- internal/printcfg: CompanyDefault() (letter, legend on, all channels), Override (pointer deltas),
  Merge(def,ov) -> effective. Axis layout NOT a knob (automatic role/uom grouping). + unit tests.
- store: `print_config TEXT NOT NULL DEFAULT ''` column on jobs (DDL for fresh DBs) + idempotent
  ADD-COLUMN migration (PRAGMA-guarded) for existing DBs. JobPrintConfig/SetJobPrintConfig
  (company-agnostic: raw JSON only). + round-trip/missing/migration-idempotent tests.
- api: GET/PUT /api/jobs/{id}/print-config -> {effective,override,default}; PUT validates pageSize,
  canonicalizes (only deltas stored), DisallowUnknownFields. 400/404 matrix. + tests.
- All store/api/printcfg tests GREEN; gofmt/vet clean; CGO-free binary builds.

## 2026-06-21 — web Report view done

- types.ts: PrintConfig/PrintOverride/PrintConfigResponse mirrors.
- JobChart: optional channel allow-list (setChannelFilter), legend toggle (setLegendVisible),
  explicit setSize for print width. Reused by the Report view.
- report.ts: ReportView — job header block + recorded chart (JobChart reuse, segment-aware) +
  minimal override editor (title/page-size/legend/channels) + Save/Reset + Print/Save-as-PDF
  (window.print()). Builds a MINIMAL override (only deltas vs company default). Managed <style>
  rewrites @page size at print time (an @page can't be selector-scoped).
- readout.ts: third "Report" tab beside Live | Job History; setView handles three views.
- styles.css: report editor + white printable sheet + @media print (hide topbar/controls/footer/
  editor/other-views; print only .report-sheet).
- Web build GREEN (tsc strict + vite, 16 modules). Offline: no external URLs in web/dist.

## 2026-06-21 — E2E + Playwright verify (post-build binary)

- Static: gofmt -l empty; go vet clean; go test ./... all GREEN (api/daqformat/parser/printcfg/
  store); CGO-free statically-linked binary; make build / web bundle clean + offline.
- E2E (real capture: testdata/intellisense-demo.txt -format intellisense, :8137):
  - Ingestion independent (axiom #1): /debug/stats rows climbed 2873->3198->3523->...->15210 while
    creating job, recording, and hammering /print-config.
  - Job flow: created+activated "Smith 4-12H", recorded a 3s segment, stopped; /api/jobs/1/series
    returned the segment + 13 in-segment channel series.
  - print-config: GET default (letter/legend/all). PUT override (title+a4+4-channel subset) ->
    effective reflects it, showLegend FALLS BACK to default (not in override blob), default
    unchanged. re-GET persists. PUT {} resets to default. Validation: bad pageSize=400,
    unknown field=400, missing job=404.
- Playwright (headless chromium-1228, executablePath override; 1.60.0 wanted 1223 not cached):
  - SYMPTOM-GONE (cosmetic a): .newjob-form computed display = "none" on load; "grid" after
    clicking "+ New job…". Screenshots 01-live / 05-form-open confirm.
  - SYMPTOM-GONE (cosmetic b): live legend has 12 rows, NO "Job Number", Job Volume PRESENT (43.3).
  - Report view: title "Smith 4-12H Surface Cement", 8 job-header meta cells, 4 channels checked
    (matching the override), recorded chart with segment band paints. Screenshot 03-report-screen.
  - @media print emulation: topbar/controls/footer/editor all display:none; .report-sheet display
    block with the chart canvas present; title shows. Screenshot 04-report-print confirms ONLY the
    report sheet (header + chart) prints, sized to page width (not blank/clipped). console_errors: [].

STATUS: complete. Clean git status before report.
