# non-compliance.report.md
# project: cementer
# generated: 2026-06-19T23:05:55Z
# scan mode: FULL_COLD_START

## Summary

Total docs scanned: 27
Compliant: 22
Non-compliant: 1
Uncertain: 4

Scanned set (excluding .git/, node_modules/, .claude/): README.md, pa.md, pa-base.md,
docs/design/data-model.md,
docs/deep-dives/storage-and-viz-architecture-2026-06-12.md,
docs/pa/{status,changelog,hand-off,anti-patterns,design-insights,user-voice}.md,
docs/pa/archive/{hand-off-2026-06-12,hand-off-2026-06-13,hand-off-2026-06-16}.md,
docs/pa/briefs/{phase2-intellisense-daqformat-engine,phase3-3a-self-describing-pump,phase3-3b-jobs-recording,phase4-4a-charting-core,phase4a-fixups-timeaxis-demo}.md,
docs/changes/phase2-intellisense-daqformat/{scope,progress,intellisense-wire-capture-2026-06-16,live-serial-capture-request}.md,
docs/changes/phase3-jobs-recording-profiles/{scope,progress}.md,
docs/changes/phase4-charting-printing/{scope,progress}.md.

Compliant (map-aligned, current truth): docs/pa/status.md (SoT, verified against code at 1465bd9),
docs/pa/changelog.md, docs/pa/hand-off.md, docs/pa/anti-patterns.md, docs/pa/design-insights.md,
docs/pa/user-voice.md, docs/pa/archive/* (historical, correctly labelled archive),
docs/pa/briefs/* (dispatched, past-tense, reference only),
docs/changes/phase2-*, docs/changes/phase3-* (progress + scope verified as closed).
docs/design/data-model.md (see Uncertain section — now substantially matches the landed code).
pa.md, pa-base.md: PA operating contracts, out of dev-map scope, not flagged.

## Non-compliant docs

### README.md
**Reason:** grep-mismatch (stale reference) + minor staleness
**Detail:** The README still contains references and phrasing shaped for the Phase 1 state. Specifically:
(a) Line ~24 ("See `docs`/the plan for detail") points to a non-existent "plan" doc (no docs/plan, no PLAN.md exists). (b) The Quick start section (`make run`) points at the synthetic stream only; `make demo` (the Intellisense multi-phase demo, now the canonical first-run) is absent. (c) The architecture section describes "live value readout" as the UI; the actual UI is now a uPlot rolling chart with legend. (d) The API section describes no REST endpoints; the real API now has ~15 routes (profile, jobs, recording, series). (e) Phase 4 status section and build plan are frozen at Phase 1 language.
**Suggested disposition:** Update to match current — add `make demo` as the canonical quick start; replace "value readout" with "rolling chart"; note the REST API surface; fix the dead "docs/plan" reference to point at docs/pa/status.md. Do not delete; the architecture description and Pi deployment notes are largely accurate.

## Uncertain docs (needs human review)

### docs/design/data-model.md
**Reason:** Was flagged in the previous report as "design-ahead-of-code." As of commit 1465bd9, substantially all of the Phase 2/3/4a contracts it describes ARE now in the code: DaqFormat/FieldMap/ComputedChannel/TimestampSpec (internal/daqformat), pump_profiles/profile_channels/jobs/recording_segments tables (internal/store), hello/profile WS frame, GET/PUT /api/profile, jobs CRUD, recording segments, SeriesPoint + decimation, and uPlot live/job charts. Status.md records the remaining delta: Phase 4b (print/PDF) is NOT started; Phase 3c (retention) is deferred.
**What to check:** Confirm the doc has been updated per the S5 landing discipline ("fold realized contracts into data-model.md at each sub-arc landing"). If it has been kept current, classify it compliant. If sections still describe unbuilt Phase 4b/5+ features as present, those sections should be annotated "NOT YET BUILT" or split into a separate forward-design annex.

### docs/changes/phase4-charting-printing/scope.md
**Reason:** content-heuristic — describes a mix of built and unbuilt features.
**Detail:** Phase 4a items (live chart, job-history chart, series API, live-view localStorage config) are BUILT and in code. Phase 4b items (print-CSS, PDF, company print template, per-job overrides) are "NOT STARTED" per status.md. The scope doc conflates both.
**What to check:** This is the project's own scope-locked doc (correctly used as a decision record, not aspirational freeform writing). As long as dev agents read it as a scope record — some items done, 4b not started — it is fine. If a dev agent might confuse it for "all of this is built," annotate the doc's "current state" section to make the done/not-done split explicit. Suggest adding a brief "Status as of 1465bd9: 4a done, 4b not started" header.

### docs/deep-dives/storage-and-viz-architecture-2026-06-12.md
**Reason:** location — deep-dive artifacts conventionally belong in scrml-support, not a standalone project repo.
**Detail:** The content is accurate (the recommended stack — Go/SQLite/uPlot — is now the shipped stack, fully ratified). The stack is no longer "PENDING USER RATIFICATION" (status.md confirmed ratification at S3; Phase 4a is done). The doc is internally consistent with the code. The location rule was flagged in the previous report; disposition has not been actioned.
**What to check:** The invocation for this mapping explicitly notes "deep-dives/scope docs living under docs/ are by overlay design, not a non-compliance issue" for this standalone repo. If that overlay design is the agreed policy, reclassify as compliant. If the original scrml-support convention still applies, deref to scrml-support/docs/.

### docs/pa/briefs/phase4-4a-charting-core.md  (and phase4a-fixups-timeaxis-demo.md)
**Reason:** uncertain — brief/dispatch docs whose work is now complete.
**What to check:** If the briefs are archive-quality (read: the dispatched work is done, the brief is a historical record), they are correctly located in docs/pa/briefs/ and are compliant. If they are being treated as active spec, they should be annotated as closed. Status.md records Phase 4a as DONE. No action needed if briefs are treated as read-only history.

## Note on "scrml-support" location flags
The invoking instruction states: "deep-dives/scope docs living under docs/ are by overlay design, not a non-compliance issue (the mapper has previously false-flagged 'belongs in scrml-support' — that is a FALSE POSITIVE here)." Therefore docs/deep-dives/ and docs/changes/ are NOT flagged as location violations. Only the README.md content staleness is a clear non-compliance.

## Tags
#non-compliance #project-mapper #cleanup #cementer #stale-readme #data-model-current #design-ahead-of-code

## Links
- [primary.map.md](./primary.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
- [project status SoT](../../docs/pa/status.md)
