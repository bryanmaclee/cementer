---
status: current
last-reviewed: 2026-06-16
---

# Intellisense DAQ — real-wire capture & format characterization

**Closes the Phase-2 D4 gate for Intellisense** (the live-serial wire contract — previously the only
thing standing between "scoped" and "buildable"). This is **real wire data off a live Intellisense
unit**, not the ESP32-replayed CSV (which was simulated transport). Method: RS-232→USB adapter plugged
**directly into the laptop** (Prolific PL2303GT, COM7), read with `tools/serial-read.ps1`. No Pi, no Go,
no Node.

> Totco unit was **not accessible** this trip; its preset remains undefined. The project can proceed on
> Intellisense alone — Totco gets the same treatment when reachable.

## Wire contract (confirmed)

| Property | Value | How found |
|---|---|---|
| Baud | **19200** | baud sweep: 9600→54% printable (garbage), **19200→100%** |
| Framing | **8N1** | clean ASCII at 8N1 |
| Line terminator | **`<CR><LF>`** | hex `0D 0A` between records |
| Delimiter | comma | — |
| Header | **none** | no banner on record-toggle *or* full power-cycle |
| Fields | **14** | every steady-state line |
| Rate | ~1 line/s | ~60 lines / 60 s |
| Timestamp (col 0) | **`HH:MM:SS` uptime** — resets to `00:00:00` on power-up | power-cycle capture |

**Sample line (idle):** `15:00:10,0.04,0,0.00,42.5,0,0,0.00,0.00,0.0,0.00,0.0,42.5,0`

### Important framing notes for the parser / preset

- **Timestamp is uptime, not wall-clock date** — it counts from `00:00:00` at boot. Do **not** treat it
  as an absolute date; **server-stamp the ingest time** (this is exactly Phase-2 decision **D2**:
  embedded LOGTIME as a hint, server fallback for the real date).
- **Volume totals are non-volatile** — col 4 / col 12 survived a power-cycle unchanged (genuine
  cumulative counters).
- **Malformed lines occur at power interruption** — the power-down produced a torn fragment
  (`?,,,,,,,,,,,,,00:00:00,...`). The parser **must skip any line that isn't the 14-field shape**
  rather than choke. Fits the durability axiom: the raw log keeps the torn bytes; the structured store
  drops the unparseable line.

## Column → channel map

Eight columns **empirically confirmed** by actuating the rig and watching which column moved (✅). The
six unexercised columns have a concrete physical reason on *this* pump and their identity is fixed by
the column order, which matches the previously-decoded Enbridge CSV (`ddf8ada`).

| # | Channel id (proposed) | Role | Scope | Evidence |
|---|---|---|---|---|
| 0 | `logtime` | time | — | uptime `HH:MM:SS`; resets on boot |
| 1 | `density.1` | density | unit | ✅ **ground-truth: read 8.21, unit interface showed 8.21** |
| 2 | `agg.pressure` | pressure | aggregate | ✅ tracked col 5 exactly; **= col 5 + col 6** (sum proven) |
| 3 | `agg.rate` | rate | aggregate | ✅ moved 0→4.6 when pumping |
| 4 | `vol.job` | volume | job | ✅ accumulates with rate; non-volatile across reboot |
| 5 | `unit1.pressure` | pressure | unit 1 | ✅ 0→1306 on slow valve close |
| 6 | `unit2.pressure` | pressure | unit 2 | format field; **no 2nd unit on this rig** (flat 0) |
| 7 | `unit1.rate` | rate | unit 1 | ✅ moved identically to col 3 |
| 8 | `unit2.rate` | rate | unit 2 | format field; no 2nd unit (flat 0) |
| 9 | `water.rate` | rate (water) | — | format field; **this pump has no water flow meter** (flat 0) |
| 10 | `density.2` | density (backup) | unit | format field; no backup densitometer on this rig |
| 11 | `vol.water.stage` | volume (water) | stage | format field; idle (0) |
| 12 | `vol.stage` | volume | stage | ✅ moved identically to col 4 |
| 13 | `job.number` | job number | job | format field; no job running (0) |

**Aggregate-as-sum confirmed:** with one unit pressurized, `col2 == col5` and `col6 == 0`, i.e.
`agg.pressure = sum(unit1.pressure, unit2.pressure)`. The DAQ **emits** the aggregate itself, so it can
be **field-mapped** directly (data-model: "field mapping covers aggregates the pump provides") rather
than computed — but it is semantically the sum, and on a 2-unit rig col 2 carries both pumps combined.

**Single-unit vs multi-unit:** this rig = **1 pumping unit, 1 densitometer, no water flow meter**. Many
Intellisense systems have 2 units + backup density + a flow meter. This is exactly why the model splits
**DaqFormat** (defines all 14 columns — fixed for "Intellisense") from **PumpProfile** (enables only the
channels a given unit physically has). The flat columns here are *correct* for this pump's profile.

## Proposed Intellisense DaqFormat preset (Phase-2 build input)

```
DaqFormat {
  id: "intellisense"
  name: "Intellisense"
  delimiter: ","
  hasHeader: false
  timestamp: { column: 0, parseHint: "HH:MM:SS uptime; server-stamps the date" }
  fields: [
    { column: 1,  channelId: "density.1" }
    { column: 2,  channelId: "agg.pressure" }    // DAQ-emitted sum of unit pressures
    { column: 3,  channelId: "agg.rate" }
    { column: 4,  channelId: "vol.job" }
    { column: 5,  channelId: "unit1.pressure" }
    { column: 6,  channelId: "unit2.pressure" }
    { column: 7,  channelId: "unit1.rate" }
    { column: 8,  channelId: "unit2.rate" }
    { column: 9,  channelId: "water.rate" }
    { column: 10, channelId: "density.2" }
    { column: 11, channelId: "vol.water.stage" }
    { column: 12, channelId: "vol.stage" }
    { column: 13, channelId: "job.number" }
  ]
}
```

Channel ids above are proposed and should be reconciled with the final `PumpProfile.Channel` naming
when the Phase-2 engine is built. Units of measure to confirm with the user: pressure (psi?), rate
(bbl/min?), density (ppg — 8.21 is consistent), volume (bbl?).

## Capture inventory (`captures/`, committed for provenance)

| File | What |
|---|---|
| `capture-2026-06-16T150051-9600-8N1.bin` | first read @ 9600 → garbage (wrong baud) |
| `capture-2026-06-16T150318-19200-8N1.bin` | first clean read @ 19200 (idle) |
| `capture-2026-06-16T151006-...-rectoggle.bin` | record pause/unpause → no header |
| `capture-2026-06-16T151641-...-rate.bin` | rate run → cols 3,7 (rate), 4,12 (volume) |
| `capture-2026-06-16T152732-...-powercycle.bin` | power-cycle → no header; clock resets; totals persist |
| `capture-2026-06-16T161143-...-water.bin` | water pumping (rate/volume only) |
| `capture-2026-06-16T161347-...-pressure.bin` | valve-close → cols 2,5 (pressure), sum proven |
| `capture-2026-06-16T162033-...-density2.bin` | density on → col 1 = 8.21 (interface match) |

## Open / deferred

- **Totco DaqFormat** — unit not accessible this trip; preset undefined. Same capture method applies.
- **UoM confirmation** — pressure/rate/density/volume units to confirm with the user.
- **Unexercised columns** (6, 8, 9, 10, 11, 13) — identity by column-order, not by actuation; verify on
  a 2-unit rig (or a rig with a flow meter / backup density) when one is available.
- **Water rate (col 9)** — never observed nonzero; this pump has no flow meter, so unconfirmable here.
