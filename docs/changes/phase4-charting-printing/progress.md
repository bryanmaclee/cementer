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
