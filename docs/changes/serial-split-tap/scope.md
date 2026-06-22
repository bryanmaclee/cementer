---
status: in-progress
last-reviewed: 2026-06-21
change-id: serial-split-tap
operator: peter
phase: hardware (field-ingest enabler; not a numbered software phase)
depends-on: a multimeter reading of the DAQ TXD idle voltage ("#1", pending)
---

# Serial-split tap — scope (isolated DAQ → Pi listen tap)

> **Status: design locked, build pending one measurement.** This captures a hardware-design chat
> (P2, 2026-06-21) so the next session resumes from a written spec, not from re-derivation. The
> only open input is **#1** (the DAQ line idle voltage), which sets one resistor value. Parts are
> on order; everything else is decided.

## Goal

Let the Pi 4B ingest a real DAQ serial stream **without disturbing the system that already consumes
that serial today** — a non-invasive, galvanically-isolated **listen-only tap**. This is the field
enabler that moves cementer from bench/replay validation to a live pump. It directly serves the
project axioms: #1 (the Pi *observes*; it never gates or alters the source), #3 (the Pi is a
standalone island that just watches), #4 (durability — a clean, fault-isolated ingest path).

## What this IS / ISN'T

| In scope | Out of scope |
|---|---|
| A one-channel, **listen-only**, **opto-isolated** tap: DAQ RS-232 TX → 6N137 → **Pi GPIO UART** | Bidirectional / handshaking serial (the DAQ streams unsolicited — listen-only suffices) |
| Galvanic isolation so a Pi-side fault can't touch the production data path | Powering the tap from the DAQ line (input is self-powered by the signal; output by the Pi) |
| Bench bring-up + a coexistence test against the existing consumer | Any change to the existing ("current system") consumer wiring |
| A buildable BOM + circuit + 3-step test plan | The Totco variant (same circuit, a 2nd channel, when a unit is reachable) |

## Decision — topology

**Architecture: opto front-end → Pi GPIO UART, bypassing the USB-serial adapter.** (Ruled by the
operator, 2026-06-21: "assume we are bypassing the rs-232 to usb adapter and splitting off from
DAQ.")

Rejected alternatives and why:
- **Bare Y-cable** (TX wired to both consumer RX + Pi RX): no isolation (ground loop in a pump
  truck), and a Pi fault can pull the shared production line. ❌ unsafe for a live job.
- **Opto → MAX3232 → Waveshare USB adapter:** works, but adds a MAX3232 (SMD sourcing / more parts)
  with no benefit over feeding the Pi UART directly. Kept only as a fallback if the GPIO UART is
  ever unavailable.
- **Pre-built "Split Option B" isolated module:** not readily sourceable — hence the self-built
  protoboard route.

### Why the 6N137 (the opto choice)

The input voltage range (~0 V idle to a positive peak, with a negative excursion if the line is true
bipolar RS-232) is handled by **the resistor and diode, not the opto's rating** — the LED only ever
sees ~1.5 V forward. So the opto is chosen on: **speed** (6N137 = 10 Mbit/s, ~500× the 19200-baud
need), **through-hole DIP-8** (solders straight to protoboard — modern 3.3 V-native parts like the
Toshiba TLP2362 are SMD-only and would need breakout adapters), and a **3.3 V-capable output**. The
6N137's open-collector output has a **pull-up rail independent of its 5 V supply**, so pulling it to
**3.3 V** yields Pi-UART-safe logic with no level shifter.

**Polarity is correct without inversion:** line idle = mark = negative → LED off → output pulled
HIGH (3.3 V) = UART idle ✓; start bit = space = positive → LED on → output LOW = start ✓.

## The circuit

```
        DAQ ground domain          │ optical │      Pi ground domain
                                   │ barrier │
  DAQ TXD ──[ Rin ]──┬─────( 2 Anode)        (8 Vcc )──┬──[0.1µF]──┐
                     │         6N137 │         │        │           │
                  [1N4148]            │         │   Pi +5V (pin 2)  Pi GND
                  (cathode→pin2,      │         │                  (pin 6)
                   anode→pin3)        │         │   (7 VE)──tie──(8 Vcc)
  DAQ GND ───────────┴─────( 3 Cath)  │         │                   │
        (+ TVS across the line,       │  (5 GND)│──── Pi GND ───────┘
         see "field hardening")       │         │
                                      │  (6 VO)──┼──[ Rpu 1k ]── Pi +3.3V (pin 1)
                                      │      │   │
                                      │      └───┼──── Pi GPIO15 / RXD (pin 10)
```

**The one inviolable rule: DAQ ground and Pi ground never touch anywhere on the board.** The only
coupling between the halves is the light path inside the 6N137. Lay the protoboard out with a literal
gap down the middle — DAQ-side components left, Pi-side right.

6N137 pinout (8-DIP): `1 NC · 2 Anode · 3 Cathode · 4 NC · 5 GND · 6 Vo · 7 VE(enable) · 8 Vcc`.
Tie VE(7)→Vcc(8) or the output stays disabled.

⚠ **Fatal mistake to avoid: pull Vo up to 3.3 V, NOT 5 V.** Pi GPIO is 3.3 V and not 5 V-tolerant.

### Component values

- **Rin (input current-limit)** — sized from **#1** for ~5 mA LED drive: `Rin = (Vline − 1.5) / 0.005`.
  - ~±5 V → **680 Ω** · ~±9 V → **1.5 kΩ** · ~±12 V → **2.2 kΩ**.
  - This same ~5 mA is the load the tap adds to the DAQ driver (see coexistence risk).
- **1N4148 antiparallel across the LED** — clamps the negative half-cycle so the LED never exceeds its
  ~5 V reverse limit against a −12 V line. (Any small-signal or rectifier diode works; not in the
  signal path.)
- **Rpu (output pull-up)** — **1 kΩ** to Pi 3.3 V.
- **0.1 µF ceramic** — Vcc decoupling, pin 8 → pin 5, close to the chip.

## BOM (status as of 2026-06-21)

| Item | Role | Status |
|---|---|---|
| 6N137 (DIP-8) ×5–10 | isolation barrier (+ spares, Totco 2nd channel) | **ordered** |
| DIP-8 socket ×2–3 | swappable chip; no solder heat into the IC | **ordered** |
| 1N4148 ×pack | antiparallel LED protection | **ordered** |
| Resistors (680 Ω / 1.5 k / 2.2 k input range + 1 k pull-up) | current-limit + pull-up | **ordered / in kit** |
| P6KE12CA (TVS) | input surge protection (field) | **ordered** — see caveat |
| 0.1 µF ceramic cap | Vcc decoupling | have |
| Protoboard | build substrate | have |
| DB9 tap + terminal-block adapter | pull off TXD + GND (pin 5) without cutting the cable | have |
| Hookup wire + header pins | to Pi pins 1 (3.3 V), 2 (5 V), 6 (GND), 10 (RXD/GPIO15) | have |

**⚠ TVS voltage caveat (gated on #1):** the P6KE12CA is bidirectional, ~10.2 V standoff / ~11.4 V
breakdown. If the DAQ line swings a **full ±12 V**, a 12 V TVS will clip legitimate signal peaks and
load the driver — step up to **P6KE15CA / P6KE18CA**. If the line is **~±5–6 V** (MAX232-class), the
12 V part is correct. **Do not populate the TVS until #1 confirms the swing** — and the TVS is **not
needed for the bench build** at all (field hardening only).

## Build & test plan (each step a go/no-go gate)

1. **Solder + bench (no pump).** Populate Rin for the measured #1 voltage. Use a 2nd USB-serial
   adapter (the Waveshare or PL2303) as a *fake DAQ* replaying a captured `.bin` into the tap input;
   read on the Pi at 19200. **Gate: clean ASCII matching the capture.**
   - Pi UART setup: `raspi-config` → serial **hardware ON**, serial **console OFF**; device =
     `/dev/serial0` (`/dev/ttyAMA0`).
2. **Real wire, Pi-only.** Tap the live DAQ with no other consumer attached.
   `cementer -source /dev/serial0 -format intellisense`; watch rows climb at `/debug/stats`.
   **Gate: real-wire end-to-end on the Pi** (never yet proven — prior validation was laptop-to-USB).
3. **Coexistence.** Connect the tap **in parallel** with the existing system; verify it still reads
   perfectly with the Pi powered, unpowered, and physically yanked. **Gate: zero disturbance to the
   production path.**

## Open items / risks

- **#1 — DAQ TXD idle voltage** (the only blocker): operator to measure (multimeter on TXD vs GND at
  idle; reads negative). Sets Rin and the TVS rating. Pending; operator gathering it "in the next day
  or two."
- **Confirm one-way:** the existing consumer only *receives* from the DAQ (never transmits to it).
  The headerless continuous stream strongly implies this — confirm, since it's what makes a
  single-channel listen tap sufficient.
- **Coexistence loading (the real residual risk):** the ~5 mA LED drive is ~1.5–2× a normal RS-232
  receiver load. On a strong driver, paralleling with one existing consumer is fine; on a weak ±5 V
  driver it may pull the signal down. **Step 3 is the test.** Upgrade-if-it-fails (do NOT build yet):
  put a high-impedance RS-232 *receiver* (MAX3232) on the DAQ side first and drive the LED from its
  TTL output — needs DAQ-side power, complicates isolation, so only if Step 3 fails.
- **Totco:** same circuit, a second 6N137 channel, when a Totco unit is reachable (9600 8N1).

## Wire-contract anchor (verified, S4)

Intellisense: RS-232, **19200 8N1**, CR/LF, ~1 line/s, 14 comma fields, headerless continuous stream.
Full characterization + the format preset:
[`../phase2-intellisense-daqformat/intellisense-wire-capture-2026-06-16.md`](../phase2-intellisense-daqformat/intellisense-wire-capture-2026-06-16.md).
