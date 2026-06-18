# Phase 2 — Intellisense DaqFormat engine: build progress

Append-only log (done / next / blockers). Times in local.

## 2026-06-18

- 2026-06-18 — Startup gate PASSED: worktree == repo top-level (not integration root),
  clean tree, Go 1.26.4. Maps + anti-patterns + scope.md + findings doc read. Verified
  captures match the findings doc (idle line, pressure-capture sum relationship
  col2==col5 max=1306, col6 flat 0, all 61 lines 14-field). Confirmed store.Stats reports
  only Rows/LatestTS (no distinct-channel count) -> E2E will query sqlite3 directly for the
  channel set.
- DONE unit 1 (commit efefb1f): internal/daqformat/daqformat.go — types + doc comments.
- DONE units 2+3 (commit 7fdfe4a): internal/daqformat/engine.go Apply (tokenize, comment-skip,
  field-count guard, header-skip, field-map, transform, server-stamp TS, seq) + compute pass
  (sum/mean) + engine_test.go (all green: torn-line skip, warmup-negative passthrough,
  ts-column-not-mapped, header-skip, transform, compute sum/mean, compute-from-one-input).
- DONE unit 4 (commit 19d0cf8): internal/daqformat/presets.go — Intellisense()/Synthetic() presets +
  IntellisenseChannels()/SyntheticChannels() vocab; presets_test.go maps the REAL captured idle line
  (all 13 channels + named asserts density.1=0.04, vol.job=42.5, vol.stage=42.5), pressure-sum,
  torn-line skip, synthetic 4-channel, channel-vocab-covers-preset. All green.
- DONE unit 5 (commit 8f8b925): cmd/cementer/main.go — added -format flag (intellisense default |
  synthetic) + resolveFormat(); swapped parser.New/Parse for daqformat.New/Apply; fixed the
  docs/plan -> data-model.md package-doc ref; log line now reports the active format. parser package
  left in place (now OFF the main path -> follow-up cleanup candidate). web/dist primed via npm build.
  go build ./... + go vet ./... + gofmt clean; go test ./... green (daqformat + existing parser tests).
- DONE unit 6 — E2E verify on the post-change binary (go build -o /tmp/cementer-bin ./cmd/cementer,
  web/dist primed). sqlite3 CLI not installed -> verified via /debug/stats + a throwaway modernc Go
  query helper (created in /tmp + a temp cmd/ pkg, run, then REMOVED; git clean).
  - INTELLISENSE (captures/...-pressure.bin, -format intellisense, 20ms, loop=false): raw log = 61
    lines (durability layer 1). /debug/stats rows=793. DB: distinct_channels=13 == exactly the
    Intellisense ids; 61 rows/channel (61*13=793). agg.pressure max=1306 == unit1.pressure max=1306,
    unit2.pressure flat 0 (field-mapped sum on a single-unit rig — confirmed end-to-end, not just in
    units). TS server-stamped to today (col-0 uptime ignored). density 0 here (correct: density was a
    different capture file) — raw fidelity, not a bug.
  - /ws/live: frames type=reading, 13 channels each; a pressurized frame showed aggP=10 ==
    u1P=10 + u2P=0 (sum-ok live).
  - SYNTHETIC (testdata/sample-stream.txt, -format synthetic): 4 channels pressure/rate/density/volume,
    600 rows each = 2400. Phase-1 replay path preserved under the daqformat engine.
- DONE unit 7 (commit 45ff04b): README — architecture diagram (parser -> daqformat engine), layout
  table (added internal/daqformat, marked internal/parser superseded/off-main-path), -format flag in
  the flags list, run examples (synthetic + intellisense + prod 19200).
- FINAL: gofmt -l clean (touched files), go vet ./... OK, go build ./... OK, go test ./... green
  (daqformat + existing parser tests). make build OK (CGO_ENABLED=0, statically linked — CGO-free
  invariant holds). git status clean (web/dist + cementer binary gitignored).
- DEFERRED / follow-ups:
  - internal/parser is now OFF the main path (main.go uses daqformat). Kept in place per brief (Phase-1
    artifact + its tests). Cleanup candidate (delete parser + parser_test.go, or fold into a daqformat
    table-driven case) — NOT done in this arc; surface to PA.
  - TimestampKind has reserved-but-unimplemented kinds (HMSUptime/ExcelSerial/Unix/RFC3339); only
    ServerStamp is implemented (all shipped formats need only that). HMSUptime currently behaves as
    ServerStamp by design.
  - Nav-map regeneration (schema/api/state/structure) left to PA per brief.
- PHASE-2 STATUS: complete (D4 live-wire gate already closed by the 2026-06-16 capture; engine + preset
  built from that real wire and E2E-verified against it).
