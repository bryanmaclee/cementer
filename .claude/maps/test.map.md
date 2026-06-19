# test.map.md
# project: cementer
# updated: 2026-06-19T23:05:55Z  commit: 1465bd9

## Test Framework
Runner: Go standard `testing` (no external test deps)
Config: none (stdlib)
Run all: `PATH=$HOME/.local/go/bin:$PATH go test ./...`
Run single: `go test ./internal/daqformat -run TestEngineApply`
Web tests: none (no test runner in web/package.json)

## Test Files and Function Counts
| File                                     | Count | What it covers                                    |
|------------------------------------------|-------|---------------------------------------------------|
| internal/daqformat/engine_test.go        | 7     | Engine.Apply: blank/comment/header-skip/field-count-guard/field-parse/computed/identity |
| internal/daqformat/presets_test.go       | 5     | Intellisense + Synthetic preset shapes, channel vocab counts |
| internal/api/api_test.go                 | 7     | GET/PUT/POST /api/profile via httptest.Server     |
| internal/api/jobs_test.go                | 20    | full CRUD + active-job + recording endpoints via httptest |
| internal/api/series_test.go              | 5     | GET /api/samples + GET /api/jobs/{id}/series      |
| internal/store/profile_test.go           | 7     | seed, has, active, editor, update, reset          |
| internal/store/jobs_test.go              | 8     | create, list, get, update, active, set-active, guards |
| internal/store/recording_test.go         | 8     | start/stop/state/list/adjust + ErrRecording guard |
| internal/store/series_test.go            | 10    | Series + JobSeries: empty, raw, decimated, in-segment filter |
| internal/parser/parser_test.go           | 2     | legacy parser (off-path; retained)                |
**Total: ~79 test functions** across 10 test files.

## Test Categories
Unit (isolated, in-process): `internal/daqformat/` — engine and presets, no DB.
Integration (real in-memory SQLite via `t.TempDir`): `internal/store/`, `internal/api/` — all use real SQLite opened on a temp path, closed in `t.Cleanup`.
E2E: none.
Web: none.

## Fixtures & Factories
| Path                             | What it contains                                                          |
|----------------------------------|---------------------------------------------------------------------------|
| testdata/sample-stream.txt       | Synthetic 4-column comma-separated stream; drives `make run`             |
| testdata/intellisense-demo.txt   | Multi-phase Intellisense live capture (14-col); drives `make demo`       |
| Test-internal inline data        | All Go tests inline their own test cases; no external fixture files       |

## Pattern
Table-driven subtests (`t.Run`). API tests use `net/http/httptest.NewServer` backed by a real store on a temp SQLite DB; store tests open a store directly. Assertions use `t.Fatalf` for control-flow failures (can't continue) and `t.Errorf` for value mismatches (keep running). Cleanup is always `t.Cleanup(func() { _ = st.Close() })`.

## Tags
#cementer #map #test #go-testing #table-driven #daqformat #api-tests #store-tests #httptest

## Links
- [primary.map.md](./primary.map.md)
- [build.map.md](./build.map.md)
- [schema.map.md](./schema.map.md)
- [api.map.md](./api.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
