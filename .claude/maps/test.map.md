# test.map.md
# project: cementer
# updated: 2026-06-12T09:02:13-06:00  commit: ee446c3

## Test Framework
Runner: Go standard `testing` (no external test deps)
Config: none (stdlib)
Run all: go test ./...
Run single: go test ./internal/parser -run TestParsePermissive
Web tests: none (no test runner in web/package.json)

## Test Categories
Unit: internal/parser/parser_test.go — the ONLY test file in the repo (2 tests)
Integration: none
E2E: none
Web: none

## Test Files
internal/parser/parser_test.go
  - TestParsePermissive: table-driven; full line, blank, whitespace-only, comment (#), short line, garbage field tolerated, extra trailing fields ignored, all-garbage, whitespace around values.
  - TestParseSeqIncrementsOnlyOnReadings: seq advances only on successful readings, never on skipped (comment/blank) lines.

## Fixtures & Factories
testdata/sample-stream.txt — synthetic cement-job stream (comma-separated pressure,rate,density,volume; leading "#" comment header lines). Drives `make run`, not the unit test (the test inlines its own cases).

## Pattern
Table-driven Go subtests via t.Run. Each case has line input, wantOK, wantLen, and a `check map[string]float64` of expected channel→value. Assertions use t.Fatalf for control-flow failures and t.Errorf for value mismatches. Parser is constructed with parser.New(parser.DefaultConfig()) and a fixed timestamp (time.Unix(0,0)).

## Coverage gaps (per docs/pa/status.md)
- No tests for store (SQLite batch commit), hub (fan-out / slow-client drop), rawlog, serialreader, source replay, or the HTTP/WS layer.
- No web/client tests.
- Real 15-column Enbridge DAQ format is NOT exercised by the parser test (test uses the synthetic 4-channel layout).

## Tags
#cementer #map #test #go-testing #table-driven #parser

## Links
- [primary.map.md](./primary.map.md)
- [build.map.md](./build.map.md)
- [schema.map.md](./schema.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
