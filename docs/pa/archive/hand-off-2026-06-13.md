# Hand-off — live

`as of: Session 2 WRAPPED · 2026-06-13` (machine: Peter's garage desktop, Windows)

> Optimize for the NEXT session's pickup, not this session's terseness. **NEXT SESSION START: rotate
> this to `docs/pa/archive/hand-off-2026-06-13.md` and open a fresh hand-off.** Session 2 is wrapped +
> pushed. Prior hand-off archived at `docs/pa/archive/hand-off-2026-06-12.md`.
> **Cross-machine note:** this session ran on Peter's garage DESKTOP; the DAQ trip uses the field
> LAPTOP. The laptop syncs everything via `git pull` — this hand-off + the field runbook below are the
> pickup contract.

## ⚡ FIELD RUNBOOK — running cementer against the real DAQ

Self-contained: a field laptop + a fresh Claude session can execute this cold. The Pi (`CementSerial`)
already carries the built binary at `~/cementer-arm64` (aarch64, static, no C deps). At the DAQ:
DAQ RS-232 out → RS-232→USB adapter → Pi USB port. **Only the serial source changes vs the bench — no
recompile for the normal (8N1) case.**

1. Plug the RS-232→USB adapter into the Pi; find the stable device path:
   ```sh
   dmesg | tail -20                 # look for "... converter now attached to ttyUSB0"
   ls -l /dev/serial/by-id/         # copy the usb-...-if00-port0 path (replug-stable)
   ```
2. Run cementer against it (fresh data-dir = clean capture):
   ```sh
   ./cementer-arm64 -serial /dev/serial/by-id/<adapter> -baud <DAQ-baud> -data-dir ~/cementer-daqtest -addr :8080
   ```
   Expect: `cementer listening on :8080  (source: serial /dev/serial/by-id/... @ <baud>, ...)`.
3. Verify (second ssh window + a browser on the laptop):
   ```sh
   tail -f ~/cementer-daqtest/raw-*.log          # raw lines from the DAQ
   curl -s http://<pi-ip>:8080/debug/stats        # rows climbing
   ```
4. **The `raw-*.log` from THIS run is the real-DAQ wire capture that closes Phase 2's D4 gate.** Pull it
   back: `scp serial123@<pi-ip>:"~/cementer-daqtest/raw-*.log" .`

### ⚠ Field gotchas (read before driving)
- **Only `-baud` is a flag. Serial params are HARD-WIRED 8N1** (`serialreader.DefaultConfig`; `main.go`
  `buildSource` overrides only `BaudRate`). Line-splitting assumes `\n`/`\r\n` (`bufio` ScanLines). If
  the DAQ is **7E1 / 7O1 / different stop bits / bare-`\r` line endings** → a ~20-min code change is
  needed (add `-databits/-parity/-stopbits` flags + CR-tolerant split) via `cementer-go-engineer`.
  **→ GET THE DAQ'S SERIAL SETTINGS FROM ITS MANUAL BEFORE DRIVING** (baud / data / parity / stop / line
  terminator). This is the one unknown that can waste the trip.
- **Channel values will be WRONG** — `parser.DefaultConfig()` is the synthetic 4-channel layout, not the
  15-column Enbridge format. `/debug/stats` shows rows but mis-mapped values. Expected; the Phase 2
  no-code mapping fixes it. **NOT a hardware failure.**
- **Pi IP differs per network** — bench was `10.0.0.105`; in the field find it via the router or
  `hostname -I` on the Pi.
- **Always use the `/dev/serial/by-id/...` path**, never `/dev/ttyUSB0` (USB devices renumber).
- If cementer errors on open: `permission denied` → `sudo usermod -a -G dialout serial123` + re-login;
  `device busy` → another reader holds the port (`sudo lsof /dev/ttyUSB0`).

## ✅ Bench-top validation — VERIFIED 2026-06-13 (Peter, on CementSerial / 10.0.0.105)

The Go + SQLite Pi stack is proven on **both** serial-ingress paths. Single static aarch64 binary; no
recompile to switch source — just the `-serial` flag.

| Ingress path | Device | rows | result |
|---|---|---|---|
| GPIO UART | `/dev/serial0` → `ttyS0` @ 115200 | 2,812 | ✅ raw log + SQLite WAL + `/debug/stats` 200 |
| USB adapter | CP2102 → `/dev/ttyUSB0` (by-id) @ 115200 | 4,404 | ✅ fresh `~/cementer-usbtest` db |

- **Topology:** laptop `send_csv.py` (replays a recorded Enbridge CSV) → ESP32 (`csvToSerialSend`) →
  [GPIO UART **or** CP2102 USB] → Pi `cementer`. **Simulated transport** — recorded data, not a live DAQ.
- **Fidelity:** transport/plumbing/columns = REAL; **wire contract (framing/timing/serial params) =
  SIMULATED**, only confirmed at the DAQ. Channel semantics = wrong (Phase 2). The USB-adapter run is the
  exact Pi-side path the real RS-232→USB adapter will use.

## How the binary got onto the Pi (repro for the field laptop)

- Built on the garage DESKTOP (Windows): installed **Go 1.26.4** (`winget install GoLang.Go`). Node
  18.12.1 is too old for Vite ^8, so the real web client wasn't built — instead a 1-line
  `web/dist/index.html` **stub** satisfies `//go:embed all:web/dist` (gitignored; the UI is a placeholder,
  fine for a hardware test). Then:
  `CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o cementer-arm64 ./cmd/cementer` → 14.8 MB ELF aarch64.
- `scp`'d to `serial123@10.0.0.105:~/cementer-arm64`. **The binary is NOT in git** (build artifact, now
  gitignored). It lives on the Pi, which travels to the field.
- **To rebuild on the FIELD LAPTOP if ever needed:** install Go 1.26.4, create any 1-line
  `web/dist/index.html`, run the same `GOOS=linux GOARCH=arm64` build. (Real web UI would need Node ≥20 —
  not needed for the hardware test.) Belt-and-suspenders: keep a copy of `cementer-arm64` on a USB stick.

## Open threads
- **Phase 2 D4 gate — still OPEN.** Needs the real-DAQ wire capture (the field runbook produces it). The
  bench raw logs carry real columns (good for building the mapping) but are SIM transport, not the wire
  contract.
- **DAQ serial settings — UNKNOWN.** The gating unknown for the field trip (see gotchas).
- **Phase 2 engine+preset build** (`internal/daqformat`) not yet dispatched; `cementer-go-engineer`
  active. Buildable in parallel with the capture (engine is format-agnostic); don't flip Phase 2 "done"
  without the live capture.
- **Parked debts** (unchanged, non-blocking): no commit gate; stale `docs/plan` reference; README "Go
  1.22+" vs `go.mod` 1.26.4; committed test-rig credentials in `pi4b & test db/...README`.

## State as of close (Session 2)

| Item | State |
|---|---|
| Bench validation | ✅ both serial paths green (2026-06-13) |
| Binary on Pi | ✅ `~/cementer-arm64` (aarch64, static) |
| Field runbook | ✅ above |
| Source code | UNCHANGED since Phase 1 (this session = docs + a cross-compile, no source edits) |
| Toolchain (desktop) | Go 1.26.4 installed; Node 18.12.1 (too old for Vite 8 — stub used) |
| Git | Session 2 committed + pushed to `origin/main` |
| Machine | work done on Peter's garage desktop; field uses the laptop (→ `git pull` to sync) |

## Next priority
1. **FIELD:** get the DAQ's serial settings → run the FIELD RUNBOOK → capture the real-DAQ `raw-*.log` →
   **closes the Phase 2 D4 gate.**
2. Dispatch the Phase 2 `internal/daqformat` engine + preset build via `cementer-go-engineer`.

## Recovered-from anomalies (this session)
- **"Pi not capturing live"** — false alarm. `cementer` writes captured lines to the raw log and is
  SILENT on stdout (`handleLine` only logs on error). The terminal looked idle while capture was fine.
  Resolution: check `tail -f <data-dir>/raw-*.log` and `/debug/stats`, not the cementer console. Worth
  remembering in the field.
