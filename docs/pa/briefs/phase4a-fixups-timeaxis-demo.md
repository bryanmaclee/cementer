---
change-id: phase4-charting-printing
slug: 4a-fixups-timeaxis-demo
agent: cementer-go-engineer
model: opus
isolation: worktree
agent-id: ada35ac1792faf486
dispatched: 2026-06-19 (Session 5)
dispatched-from-tip: 1f65c13
status: dispatched
---

# Dispatch brief (verbatim) — Phase-4a fix-ups: uPlot time-axis units + varied demo capture

> Archived per pa.md §5. Exact prompt sent to `cementer-go-engineer` (background, worktree, opus).
> Triggered by the user demoing the chart: "stacking identical segments" (= the demo looping a 12s
> capture; server data verified clean) + a real uPlot time-axis units bug found in review.

---

You are the canonical cementer Go dev-agent. This is a small **Phase-4a fix-up** (4a landed at commit `1f65c13`). Two issues found while demoing the charts:

**ISSUE 1 (real bug, affects production) — uPlot time axis is fed milliseconds where uPlot expects SECONDS.** uPlot's time scale (`scales: { x: { time: true } }`) expects Unix timestamps in **seconds**; both charts feed milliseconds, so the time-axis LABELS are wrong (off by 1000×; dates land in the wrong era). The line *shapes* are correct — only the time axis/labels are wrong.
- `web/src/chart/livechart.ts`: `push()` does `const tMs = Date.parse(r.ts)` (ms) and pushes that as x; the rolling-window trim uses `windowMs`/`DEFAULT_WINDOW_MS` in ms; the personal-config window is in ms.
- `web/src/chart/jobchart.ts`: `render()` builds the union x with `p[0] / 1000` (µs→ms) and the `segmentShadePlugin` uses `seg.startedAtUs / 1000`, `(seg.stoppedAtUs ?? Date.now()*1000) / 1000` (ms).
**Fix:** feed uPlot **seconds** consistently (the idiomatic uPlot unit for `time: true`). I.e. x = `Date.parse(r.ts) / 1000` (live) and `p[0] / 1_000_000` (job, µs→s); make the trim/window math and the shade-plugin coordinates use the SAME unit (seconds) so everything stays consistent. (Keep the personal-config window value human-meaningful — store seconds or convert cleanly; just make it consistent.) Verify the x values handed to uPlot are seconds end-to-end. Do NOT change the data on the wire (Reading.ts stays RFC3339; the store stays µs) — this is purely the client converting to seconds for uPlot.

**ISSUE 2 (demo quality) — `make demo` loops a single 12-second capture**, so the chart shows the same `0→1306→0` pressure ramp ~25× across the window (a repetitive sawtooth; "looks wrong"). Make the demo a varied, job-like stream:
- Build a concatenated demo file `testdata/intellisense-demo.txt` from the REAL `captures/*.bin` files (they are newline-delimited 14-field Intellisense lines). Concatenate the **19200** captures in chronological (filename-timestamp) order — i.e. all of `captures/capture-2026-06-16T15*-19200-8N1*.bin` and `...T16*-19200-8N1*.bin` — **EXCLUDING the 9600 garbage file** (`...150051-9600-8N1.bin`). This yields a multi-phase stream (idle → rate → density → water → pressure → density) ~450 lines, so a loop is ~90s and the chart shows real variety (pressure to 1306, density ~8.21, rate/water/volume movement), not one repeating ramp. Torn boot fragments in the powercycle capture are fine — the field-count guard drops them.
- Point `make demo` at the new file: `./$(BIN) -source testdata/intellisense-demo.txt -format intellisense -replay-interval 200ms` (keep `-replay-loop` default true). Update the README "Quick demo" note if the path/behavior wording needs it (it currently names the single pressure capture — change it to the demo stream; keep the "what you'll see" guidance, drop or soften the "four traces flat at 0 / single-unit rig" note since the concatenated stream now exercises more channels — verify which channels actually move and describe accurately).

═══════════════════════════════════════════════════════════════════════
GATES + DISCIPLINE
═══════════════════════════════════════════════════════════════════════
- STARTUP: cwd == your worktree (NOT `/home/bryan-maclee/cementer`); clean tree; `export PATH=$HOME/.local/go/bin:$PATH`; `go version`. This touches web + needs the binary, so `cd web && npm install && npm run build`, then `make build` for E2E.
- MAPS stale (`ee446c3`); current truth = source + `docs/design/data-model.md` + `docs/changes/phase4-charting-printing/scope.md`. Likely not load-bearing for this small fix; say so.
- ANTI-PATTERNS Part B (vanilla-TS, NO framework; uPlot is a library; bundle offline). Read it before the chart edits.
- AXIOMS: unchanged — the charts stay READ-ONLY (axiom #1); no store/format/server changes beyond the Makefile + the demo asset. Do NOT touch the store, the daqformat engine, or the wire contract.
- Crash-recovery: WIP-commit per fix; append `docs/changes/phase4-charting-printing/progress.md`.

═══════════════════════════════════════════════════════════════════════
VERIFY-BEFORE-CLAIM (note the browser limitation honestly)
═══════════════════════════════════════════════════════════════════════
- `make build` clean; `gofmt -l` empty; `go vet ./...`; `go test ./...` all green (no Go logic changed, but confirm nothing broke).
- Demo plays: run `./cementer -source testdata/intellisense-demo.txt -format intellisense -replay-interval 50ms -replay-loop=true -data-dir /tmp/ce-demo -addr :8099`; confirm via `/api/samples` (or the DB) that MULTIPLE channels move across the stream — e.g. agg.pressure reaches ~1306, density.1 reaches ~8.2, agg.rate and a volume channel move — i.e. the concatenated stream is varied, not a single ramp. Report the per-channel max for the enabled channels.
- Time-units fix: show (from the built JS or by reasoning over the code) that the x values handed to uPlot are now seconds (e.g. ~1.78e9, not ~1.78e12). State clearly that the VISUAL axis-label correctness needs a human browser load (no headless browser here) — you verify the units + that it builds + that tsc-strict passes, NOT the rendered labels.
- Axiom #1 still holds: `/debug/stats` rows climb while chart endpoints are hit.
- Confirm uPlot still bundled offline (no CDN).

═══════════════════════════════════════════════════════════════════════
REPORT BACK
═══════════════════════════════════════════════════════════════════════
(a) worktree path + branch + final SHA; (b) files-touched; (c) verify output (build/gofmt/vet/test; the demo per-channel maxes proving variety; evidence the uPlot x is now seconds; axiom-#1 rows-climb; offline-bundle); (d) what needs a human browser eyeball (the time-axis labels); (e) anything contradicting this brief; (f) progress.md updated. `git status` clean in the worktree before DONE. Keep the diff tight — this is a focused fix, not a feature.
