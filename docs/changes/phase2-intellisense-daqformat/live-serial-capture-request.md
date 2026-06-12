---
status: current
last-reviewed: 2026-06-12
for: hardware collaborator (Peter Oliver)
re: Phase 2 — Intellisense DaqFormat preset
---

# Request: a raw LIVE-SERIAL capture from a real Intellisense pump

**Why.** We're defining the Intellisense `DaqFormat` preset from the Enbridge **CSV exports** in
`esp32sketches/`. Those are file exports (laptop → ESP32 → Pi replay). Before we lock the preset and
call Phase 2 done, we need to confirm the **on-the-wire** frames a live Intellisense pump actually
emits — they may differ from the export (header presence, embedded timestamp, framing, field order).

## What we need (a raw byte capture)

A few minutes of the **raw serial bytes** as they arrive on the Pi from a real (or shop-simulated)
Intellisense unit — *unmodified*, before any Python parsing. E.g. on the Pi:

```sh
# raw, no transformation — replace ttyXXX + baud with the real device/rate
stty -F /dev/ttyXXX 9600 raw -echo        # set the line params you actually use
cat /dev/ttyXXX > live-serial-capture.txt # ctrl-C after ~2-5 min of live data
# (or a pyserial logger that writes bytes verbatim, incl. line terminators)
```

Capture during a period with **non-zero activity** (some pressure/rate) if possible, and ideally
across a **stage change** and a **marker press** so we see those columns move.

## Questions to answer with the capture

1. **Header:** does the live wire send the `_00_LOGTIME,_01_DENSITY,...` header line (once at start?
   periodically? never)? Or only the CSV *export* has it?
2. **Timestamp:** is `_00_LOGTIME` present in the live frames, and in the same Excel-serial day-number
   format (e.g. `46170.290613`)? Or does the live stream omit it (so we server-stamp)?
3. **Framing:** line terminator — `\n`, `\r\n`, or other? Any leading/trailing control characters,
   STX/ETX, or handshake bytes?
4. **Delimiter + field count/order:** still comma, still 15 columns in the same order as the export?
5. **Baud rate** and serial params (data bits / parity / stop bits) for the real pump.
6. **Sample rate** on the wire (lines/sec).
7. **Intellisense version/model** if known (helps anticipate format variants on other pumps).

## What we do with it

The generic `internal/daqformat` engine is format-agnostic, so we can build it now; the **Intellisense
preset's exact `hasHeader` / `timestamp` / framing fields** get finalized against this capture, and the
Phase-2 "done" verification runs the real live-serial bytes end-to-end through the binary.

> **Alternative if a real pump isn't reachable soon:** explicitly ratify the **CSV-export shape**
> (header + Excel-serial LOGTIME + 15 comma-separated cols + `\n`) as the accepted wire contract, and
> we proceed on that — revisiting only if a live pump later proves different.
