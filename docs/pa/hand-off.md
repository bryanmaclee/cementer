# Hand-off — live

`as of: Session 4 close · 2026-06-16` (machine: field LAPTOP — `C:\Users\pjoli\...\cementer`, Windows)

> Optimize for the NEXT session's pickup. Session 4 captured the **Intellisense** DAQ wire live and
> closed the Phase-2 **D4** gate for it. Prior hand-offs: Session-2 snapshot at
> `archive/hand-off-2026-06-13.md`; Session-1 at `archive/hand-off-2026-06-12.md`.
>
> **Correction carried from S4 start:** the Session-3 hand-off said "PAUSED, uncommitted, nothing
> committed." That was wrong — commit `04ba031` had committed + pushed S3's work. Don't trust a prior
> hand-off's push/commit narrative; verify git STATE (`git status` + `git rev-list --count
> origin/main...main`) at session start.
>
> **Push pending at S4 close:** this session's doc edits + `captures/` + a leftover S3 changelog block
> are committed-pending (user pushes via **GitHub Desktop** — `git push` hangs on Git Credential
> Manager). Surface this; do not assume it's pushed.

## ✅ Session 4 result — Intellisense wire captured (D4 CLOSED for Intellisense)

Captured **real raw-data off a live Intellisense unit** via direct-laptop serial (Prolific PL2303GT
adapter, **COM7**; `tools/serial-read.ps1`). No Pi, no Go, no Node.

**Wire contract:** **19200 8N1**, `<CR><LF>`, comma-delimited, **no header**, **14 fields**, ~1 line/s,
timestamp = **`HH:MM:SS` uptime** (resets to `00:00:00` on boot — NOT a wall-clock date → server-stamp
on ingest, per D2).

**Column map** (8/14 empirically confirmed by actuating the rig):

| # | channel | # | channel |
|---|---|---|---|
| 0 | time (uptime) ✅ | 7 | rate unit1 (=3) ✅ |
| 1 | **density 8.21=interface** ✅ | 8 | rate unit2 (none) |
| 2 | **pressure agg =5+6** ✅ | 9 | water rate (no flow meter) |
| 3 | rate agg ✅ | 10 | backup density (none) |
| 4 | volume job total ✅ | 11 | water stage total (idle) |
| 5 | **pressure unit1 0→1306** ✅ | 12 | volume stage total (=4) ✅ |
| 6 | pressure unit2 (none) | 13 | job number (idle) |

The 6 flat columns are correct for this **1-unit, no-backup-density, no-flow-meter** rig; the format
keeps them for multi-unit rigs (DaqFormat defines all 14; PumpProfile enables what a unit physically
has). **`agg.pressure = sum(unit pressures)` proven.**

**Full characterization + Phase-2-ready DaqFormat preset:**
[`docs/changes/phase2-intellisense-daqformat/intellisense-wire-capture-2026-06-16.md`](../changes/phase2-intellisense-daqformat/intellisense-wire-capture-2026-06-16.md).
Raw captures: `captures/*.bin` (10 files, committed — not gitignored).

## ▶ Next priority

1. **Build Phase 2** — `internal/daqformat` engine (no-code field-mapping + compute layer) + the
   **Intellisense preset** from the findings doc. Now **fully unblocked** (D4 closed for Intellisense).
   Dispatch via `cementer-go-engineer` (canonical Go dev-agent), worktree-isolated, `model: opus`.
   - **Parser robustness:** skip any line that isn't the 14-field shape (boot produces torn `?,,,,...`
     fragments). Raw log keeps the bytes; structured store drops the bad line.
   - **Timestamp:** treat embedded `HH:MM:SS` as uptime/hint; server-stamps the real date (D2).
   - **UoM to confirm with user:** pressure (psi?), rate (bbl/min?), density (ppg — 8.21 fits), volume.
2. **Totco preset — still TODO.** Unit was not accessible this trip. Same direct-laptop method applies;
   resume steps for the prior Totco/COM6 blocker are in the §"Totco (deferred)" below.

## 🔌 Direct-laptop serial capture — the proven method (reuse for Totco)

`tools/serial-read.ps1` (PowerShell `System.IO.Ports.SerialPort`):
- normal read + hex/ASCII dump + save to `captures/`:
  `powershell -File tools/serial-read.ps1 -Port COM7 -Baud 19200 -Seconds 15`
- baud unknown → find it by **printable ratio** (the built-in `-Sweep` only counts bytes; the
  per-baud-printable-ratio loop used this session is the better tool — 100% printable = right baud,
  garbage = wrong baud, silence at ALL bauds = physical).
- `-Loopback` (jumper DB9 pin 2↔3, unplugged) = adapter self-test.

**Column-mapping recipe that worked:** capture per actuation (rate / pressure / density), parse 14-field
lines, report which columns changed (distinct-count + min/max) — the moved column = the actuated channel.
Narrate the action while capturing; give a wide window (60–90 s) so timing isn't tight.

## Totco (deferred — not accessible 2026-06-16)

Totco settings from its config screen: **COM6 · 9600 8N1 · Protocol 1 · 250 ms**. Session-3 hit **total
silence on COM6 at every baud** → diagnosed physical/electrical, not settings. Resume order when the unit
is reachable: (1) loopback self-test (`-Loopback`, pin 2↔3 jumpered) — decisive about our side; (2)
confirm a **null-modem/crossover** cable (two DTE ends); (3) confirm the Totco is actually **streaming**
(not just configured). Then capture as for Intellisense.

## ⚡ FIELD RUNBOOK — running the full cementer binary against a real DAQ (Pi deployment)

Still valid for actual *deployment* (vs the laptop-direct *discovery* above). The Pi (`CementSerial`)
carries the built binary at `~/cementer-arm64` (aarch64, static). At the DAQ: DAQ RS-232 → adapter →
Pi USB.
1. Find the stable device path: `ls -l /dev/serial/by-id/` (use the `usb-...-if00-port0` path, never
   `/dev/ttyUSB0`).
2. `./cementer-arm64 -serial /dev/serial/by-id/<adapter> -baud 19200 -data-dir ~/cementer-daqtest -addr :8080`
   (Intellisense is **19200 8N1** — confirmed this session; note `-baud` is the only serial flag,
   params are hard-wired 8N1; a non-8N1 DAQ would need a ~20-min flag addition).
3. Verify: `tail -f ~/cementer-daqtest/raw-*.log` + `curl -s http://<pi-ip>:8080/debug/stats`.
4. cementer is SILENT on stdout when healthy — check the raw log / `/debug/stats`, not the console.

## Bench-top validation — VERIFIED 2026-06-13 (Peter, on CementSerial / 10.0.0.105)

Go+SQLite Pi stack proven on both serial-ingress paths (GPIO UART @115200: 2,812 rows; CP2102 USB
@115200: 4,404 rows) — raw log + SQLite WAL + `/debug/stats`. Transport was **simulated** (ESP32-replayed
Enbridge CSV). The **real-DAQ wire contract is now confirmed for Intellisense** (this session) — the bench
validated the stack, this session validated the wire.

## State as of close (Session 4)

| Item | State |
|---|---|
| Intellisense wire contract (D4) | ✅ CLOSED — 19200 8N1, 14-col, mapped |
| Totco wire contract | ⬜ TODO (unit not accessible) |
| Phase 2 `internal/daqformat` build | ⬜ not started — now fully unblocked |
| Source code | UNCHANGED since Phase 1 (this session = docs + captures, no source) |
| Git | S3 committed+pushed (`04ba031`); **S4 push pending** (GitHub Desktop) |
| Canonical Go dev-agent | `cementer-go-engineer` (active) |

## Parked debts (unchanged, non-blocking)

- No commit gate installed (`core.hooksPath` unset). Baseline: `gofmt -l` + `go vet` + `go build` +
  `go test`; `make build` pre-push when `web/` changed.
- Stale `docs/plan` reference in `cmd/cementer/main.go` + `README.md` (doc doesn't exist).
- README "Go 1.22+" vs `go.mod` 1.26.4.
- Plaintext test-rig credentials committed in `pi4b & test db/...README` (rotate if repo ever shared).
