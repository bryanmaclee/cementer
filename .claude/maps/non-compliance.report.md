# non-compliance.report.md
# project: cementer
# generated: 2026-06-12T09:02:13-06:00
# scan mode: FULL_COLD_START

## Summary

Total docs scanned: 13 (11 *.md + 2 *.README that function as docs)
Compliant: 6
Non-compliant: 5
Uncertain: 2

Scanned set (excluding .git/, node_modules/, .claude/): README.md, pa.md, pa-base.md,
docs/design/data-model.md, docs/deep-dives/storage-and-viz-architecture-2026-06-12.md,
docs/pa/{status,changelog,hand-off,anti-patterns,design-insights,user-voice}.md,
esp32sketches/pythonScript.README, "pi4b & test db/credetials&currentDB.README".

Compliant (map-aligned, current truth): docs/pa/status.md (the SoT, verified against code),
docs/pa/changelog.md, docs/pa/hand-off.md, docs/pa/anti-patterns.md, docs/pa/design-insights.md,
docs/pa/user-voice.md. pa.md and pa-base.md are the PA operating contract (primary-agent scope,
not dev-map content) — not flagged.

## Non-compliant docs

### README.md
**Reason:** grep-mismatch (dead reference) + minor staleness
**Detail:** References a build-plan doc that does not exist. Line 24 ("See `docs`/the plan for
detail"), line 67-72 ("the two chart-config scopes... described in data-model.md"), and line 86
("See the build plan for the phased roadmap") cite a "plan" that grep cannot find anywhere in the
repo (no docs/plan, no PLAN.md). Separately, line 43 says "Requires Go 1.22+" while go.mod pins
go 1.26.4. The architecture/layout/build sections themselves match the code and are accurate.
**Suggested disposition:** update to match current — point "the build plan" at docs/pa/status.md
(the live SoT) and docs/design/data-model.md § "Build order"; bump the Go version note. Do not
delete; the doc is mostly current.

### cmd/cementer/main.go  (pkg-doc comment — source file, reported as a doc-currency defect)
**Reason:** grep-mismatch (dead reference in shipped source comment)
**Detail:** The package doc (line ~7: "Pipeline (see docs/plan): ...") cites docs/plan, which does
not exist. The pipeline description itself is accurate; only the path reference is dead. (This is a
source file, not a doc — noted here because it is the same dead "docs/plan" reference and a dev
agent following it will hit nothing.)
**Suggested disposition:** update to match current — change "see docs/plan" to "see
docs/design/data-model.md" (or remove the parenthetical). Source edit, for a human/dev agent.

### docs/deep-dives/storage-and-viz-architecture-2026-06-12.md
**Reason:** location
**Detail:** A deep-dive research/decision artifact. Per scope rules, deep-dives belong in
scrml-support, not the project repo. Content is internally accurate and marked `status: current`,
and its recommendation (adopt the Go/SQLite/uPlot stack) matches the shipped code — so it is not
stale, but it is out-of-place for a dev-scoped repo and describes a decision PENDING USER
RATIFICATION (per status.md), i.e. not yet ratified truth.
**Suggested disposition:** deref to scrml-support/docs/ (keep the artifact, move it out of the
project repo). Leave a one-line pointer in docs/pa/status.md (already present).

### "pi4b & test db/credetials&currentDB.README"
**Reason:** combo (out-of-scope content + committed plaintext credentials + describes a non-product stack)
**Detail:** Describes the collaborator's Python → ESP32 → InfluxDB 2.9.1 → Grafana 13.0.2 test rig
— a stack the deep-dive explicitly recommends NOT shipping. None of the identifiers (InfluxDB,
Grafana, cement_data bucket, daq_to_influx.py) exist in the Go source. It also commits plaintext
SSH / InfluxDB / Grafana credentials (identical weak test passwords). This is a diagnostic bench
artifact, not product source, AND a credential-exposure flag.
**Suggested disposition:** deref to scrml-support/archive/ (bench artifact) AND rotate the
credentials + remove from version control (gitignore the creds file). Do not map as product infra.
NOTE: per identity rules, credential values are NOT reproduced here — only that they exist.

### esp32sketches/pythonScript.README
**Reason:** location / out-of-scope (describes the dev/diagnostic bench, not product)
**Detail:** Operating notes for send_csv.py, part of the ESP32 CSV-injection rig. References Python
3.14 + send_csv.py; no corresponding identifiers in the Go product. It is bench tooling
documentation, correctly real for the bench but not part of the shipped product.
**Suggested disposition:** deref to scrml-support/archive/ with the rest of the bench, OR keep
co-located with esp32sketches/ but clearly labeled "dev bench, not product." Not product-map content.

## Uncertain docs (needs human review)

### docs/design/data-model.md
**Reason:** Marked `status: current` and is treated as the normative design doc, but the bulk of
what it specifies (PumpProfile, Channel, DaqFormat, FieldMap, ComputedChannel, RecordingSegment,
Job, hello/profile WS message, the two chart-config scopes, the compute layer) has NO code and NO
tables. The store has only `samples`. By strict "current truth only" this is design-ahead-of-code;
but the project intentionally keeps it as forward-looking normative design, and status.md tracks
the deltas explicitly. So it is neither stale-and-wrong nor fully-implemented.
**What to check:** Confirm the intent — should data-model.md remain a normative *design* doc
(recommended: keep, it is the agreed Phase-2/3/4 spec and status.md is the "what's actually built"
SoT), or be split into "shipped" vs "planned" sections? If kept, ensure every dev agent reads it as
DESIGN, not as a description of current schema. The DESIGNED-not-built items are enumerated in
schema.map.md and state.map.md so the maps do not mislead.

### docs/pa/status.md  (and the design ↔ code deltas it tracks)
**Reason:** status.md is current and verified, but it documents a "MAJOR FORK" + a deep-dive whose
recommendation is "PENDING USER RATIFICATION." The architectural direction is therefore documented
but not formally ratified — a dev agent could read the recommendation as settled.
**What to check:** Confirm whether stack (A) (Go + SQLite + uPlot) is now ratified. If yes, mark it
ratified in status.md and the deep-dive can be archived to scrml-support. If not, ensure agents know
the storage/viz architecture is the recommended-but-unratified direction. (status.md is otherwise
the most accurate doc in the repo and is treated as compliant.)

## Tags
#non-compliance #project-mapper #cleanup #cementer #design-ahead-of-code #dead-reference #credentials

## Links
- [primary.map.md](./primary.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
- [project status SoT](../../docs/pa/status.md)
- [scrml-support archive convention](../../../scrml-support/pa.md)
