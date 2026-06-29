---
status: in-progress
last-reviewed: 2026-06-28
change-id: serial-split-tap
operator: peter
phase: hardware (field-ingest enabler; not a numbered software phase)
depends-on: "#1" (DAQ TXD idle voltage) -- MEASURED 2026-06-25 (Intellisense -6.35V, Totco -8.20V)
---

# Serial-split tap — scope (isolated DAQ → Pi listen tap)

> **Status: step-1 bench gate PASSED on the SOLDERED PROTO 2026-06-28 (P6)** — `Rin` re-tuned + locked at
> **1 k** (validated at the real +6.35 V amplitude), proto soldered, full path re-proven end-to-end with the
> Waveshare RS-232 source (clean lines → cementer ingest → live chart). See "P6 soldered-proto validation".
> Breadboard gate passed first at P5 (see "P5 bench validation"). **Next: field steps 2-3** (real Intellisense
> wire Pi-only, then coexistence), then the v2 Amphenol pass-through prototype. This captures a hardware-design chat
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

## DAQ behavior + measurements -- `#1` RESOLVED (P4, 2026-06-25)

Operator measured both DAQ TXD idle voltages (multimeter, TXD vs GND; reads negative = RS-232 mark):

| Unit | GND / TXD | Idle (mark) | `Rin`=(V-1.5)/5mA | Pick | Read | TVS |
|---|---|---|---|---|---|---|
| **Intellisense** | pin1 / pin2 -- **transmit-only, 2-wire** (no handshake pins) | **-6.35 V** | 970 ohm | **1 kohm** | 19200 8N1 | P6KE12CA |
| **Totco** | pin5 / pin2 | **-8.20 V** | 1.34 kohm | **1.5 kohm** | 9600 8N1 | P6KE12CA |

Both lines are <+-10 V, so the **P6KE12CA TVS covers both** (no P6KE15/18CA upgrade) -- field only, skip on bench.

### Totco is DTR-gated (not command-polled) -- evidence-based finding
Totco pin 2 (TXD) is driven at -8.2 V mark whenever the unit is powered (even USB unplugged) -> the
transmitter is always alive. DATA appears on pin 2 ONLY while the consumer software runs, and exactly then
**pin 4 -> +9.25 V (DTR asserted)** while **pin 3 (RXD) stays idle mark -- no command bytes ever go in.** So
the Totco streams **only while the consumer asserts DTR**, not on a command.
- **Listen-tap implication:** works in **coexistence** (consumer holds DTR -> Totco streams -> we listen); a
  **Pi-only standalone read sees silence** unless the Pi asserts DTR. -> **validate Totco via the coexistence
  test (step 3), not the Pi-only step 2.** (Intellisense, transmit-only, streams standalone.)
- **Confirm test:** disconnect the consumer, jumper **pin 4 -> +5..9.25 V**, watch pin 2 -- streams =
  confirmed DTR-gate; silence = wrong, dig further. Likely also explains the S3 "total silence on COM6".

### Multi-DAQ: one board or two?
A "2-in-1" is electrically just **2x this identical circuit** (different `Rin` + read-baud per channel),
buildable with parts in hand. **Recommended: build/validate the Intellisense channel first** (the sure
thing), then add Totco as a 2nd opto channel (same board or separate). Keep the **three ground domains
separate** (Intellisense-GND, Totco-GND, Pi-GND never bridge).

### v2 field form factor (operator's plan)
6-pin Amphenol -> splitter protoboard (data + GND **pass straight through** to a 2nd Amphenol that continues
the normal run) -> opto branch off the same node -> Pi. The pass-through is **continuous wire**, so the
existing consumer's line is electrically unchanged except for the opto's ~5 mA tap load (= the step-3
coexistence test). **v2 prereq: map the 6-pin Amphenol pinout (data + GND).**

### Bench fake-DAQ source (no Waveshare)
The operator has no Waveshare; use the **field DB9->USB adapter run as a transmitter**. Its data exits on
**TXD = DB9 pin 3** (the field readings were on pin 2 = the adapter's RXD/receive side). Replay a captured
`.bin` out the adapter COM port @19200; tap pin 3 + GND pin 5 into the opto input via the Jienk DB9 breakout.

## Build & test plan (each step a go/no-go gate)

1. **Solder + bench (no pump). ✅ PASSED on BREADBOARD 2026-06-27 (P5) AND on the SOLDERED PROTO 2026-06-28
   (P6)** — see "P5 bench validation" + "P6 soldered-proto validation" below. Source: a **Waveshare
   USB->RS232 adapter** (operator now has one; supersedes the field-adapter plan) driven by
   `tools/intellisense-send.ps1` (PC PowerShell, real RS-232). Tap its **TXD = DB9 pin 3** (DTE) + GND pin 5
   into the opto input; read on the Pi `/dev/serial0` @19200. **Gate met (both boards): cementer ingested
   clean 14-field frames into the store (`/debug/stats` rows climbing) + the live chart painted over WiFi.**
   `Rin` **locked at 1 k** (validated at +6.35 V, the real DAQ amplitude). **Step 1 is DONE.**
   - Pi UART setup: `sudo raspi-config` → serial **hardware ON**, serial **console OFF** → reboot;
     device = `/dev/serial0` (on a Pi 4 this is the **mini-UART `ttyS0`**, NOT `ttyAMA0`; `enable_uart=1`
     from "hardware ON" locks the core clock so 19200 stays accurate — no Bluetooth-disable needed).
2. **Real wire, Pi-only.** Tap the live DAQ with no other consumer attached.
   `cementer -serial /dev/serial0 -baud 19200 -format intellisense`; watch rows climb at `/debug/stats`.
   **Gate: real-wire end-to-end on the Pi** (never yet proven — prior validation was laptop-to-USB / the
   P5 bench used the Waveshare, not the pump). ⚠ cementer sets the port baud **itself** via `go.bug.st/serial`
   (ignores `stty`); the `-baud` flag **defaults to 9600**, so passing `-baud 19200` is mandatory. The
   device path is **`-serial`**, NOT `-source` (which is a replay *file*).
3. **Coexistence.** Connect the tap **in parallel** with the existing system; verify it still reads
   perfectly with the Pi powered, unpowered, and physically yanked. **Gate: zero disturbance to the
   production path.** (P5's lower `Rin` raises the tap load to ~7-9 mA — validate here; MAX3232 buffer is
   the documented fallback if it disturbs the consumer.)

## P5 bench validation (2026-06-27) — step 1 PASSED on breadboard

Full opto path proven end-to-end on the breadboard: **PC sender → Waveshare USB→RS232 → 6N137 opto → Pi
mini-UART → cementer → SQLite → web UI over WiFi (live chart).** `/debug/stats` climbed steadily
(208 → 1079 rows ≈ 14 rows/s = ~1 line/s × 13 channels). The bench source is the Waveshare adapter run as
a transmitter (the field-DB9-adapter plan is superseded — operator acquired a Waveshare).

**The working recipe (reuse for the soldered-proto re-test):**
- **DAQ side (breadboard):** Waveshare DB9 **pin 3 (TXD)** → `Rin` → 6N137 pin 2 (anode); Waveshare DB9
  **pin 5 (GND)** → 6N137 pin 3 (cathode); 1N4148 antiparallel (**band/cathode → pin 2**, anode → pin 3).
- **Pi side:** pin 8 (Vcc) → Pi 5V (pin 2); pin 7 (VE) → tie to Vcc; pin 5 (GND) → Pi GND (pin 6);
  0.1 µF across 8↔5; pin 6 (Vo) → `Rpu` 1 k → Pi 3.3V (pin 1) **and** → Pi pin 10 (GPIO15/RXD).
  **Grounds never bridge** (opto + WiFi are the only couplings; PC and Pi share no ground).
- **PC sender:** `tools/intellisense-send.ps1 -Port COM6` (PowerShell, .NET `SerialPort`, 19200 8N1,
  CR/LF, ~1 line/s, triangle-wave so the chart moves). Needs `Set-ExecutionPolicy -Scope CurrentUser
  RemoteSigned` once. Only ONE process can hold the COM port — close the window or `$sp.Close()` to free it.
- **Pi read (eyeball gate):** `stty -F /dev/serial0 19200 raw -echo -crtscts; cat /dev/serial0` → clean
  14-field lines. (Stop the login console first if it fights the port: raspi-config console OFF.)
- **Pi ingest:** `~/cementer-arm64-new -serial /dev/serial0 -baud 19200 -format intellisense -data-dir
  ~/cementer-splittest > ~/cementer.log 2>&1 &` then `curl -s localhost:8080/debug/stats`; browser =
  `http://<pi-ip>:8080`.
- **Binary:** the Pi has no Go/Node; cross-compile on the laptop: `$env:GOOS='linux'; $env:GOARCH='arm64';
  $env:CGO_ENABLED='0'; go build -o cementer-arm64-new ./cmd/cementer` then `scp … serial123@<pi-ip>:~/`.
  **Stop the running binary before scp** (a live ELF is "text file busy" → scp "dest open Failure").
  The web UI needs a real `web/dist` (`cd web && npm run build`) embedded at build time — needs Node ≥ 20.

**Findings / gotchas (cost real time — read before the proto build):**
- **DOA optocoupler.** The first 6N137 had a **dead output stage** — LED driven correctly at ~6 mA, Vcc/VE/GND
  all good, but Vo never switched. Swapping in a spare fixed it instantly. **Test each 6N137 before trusting
  the build** (cheap optos have a real DOA rate). Diagnostic that isolated it: a continuous `0x00` flood
  holds the line ~90% positive (DMM-visible); `BreakState` on the FTDI **does not transmit** a break, so use
  the flood, not a break, to force a sustained LED-on.
- **`Rin` re-tune — RESOLVED at P6: locked at 1 k.** P5 had settled at 560 Ω while a DOA chip masked the
  margin. P6 re-tuned with the good chip using a **static +6.35 V PSU inject (the real DAQ amplitude, not the
  weak Waveshare ~+5 V)** as the gauge: at 1 k / +6.35 V the LED draws ~4.9 mA and Vo switches **solidly LOW
  (0.19 V breadboard, 0.059 V soldered)** — full clean swing 3.3 V ↔ ~0.1 V. **1 k is the field value** (lowest
  tap load → best for coexistence; the stronger real DAQ only adds margin over the bench). Lesson: gauge `Rin`
  at the *field voltage* (PSU), not the weaker bench source, to avoid over-specifying current.
- **Polarity is correct with NO inversion** on the real-RS-232 path (Waveshare or real DAQ): idle mark
  (negative) → LED off → Vo HIGH = UART idle. (The inversion + smaller `Rin` only applied to the abandoned
  ESP32-TTL "Option B".)
- **Pi 4 baud trap:** `/dev/serial0 → ttyS0` (mini-UART), and a reboot/console resets it to **9600** →
  total garbage at 19200. Fix: console OFF + set 19200 (and cementer sets its own baud anyway). The PL011
  Bluetooth-disable trick was *not* needed once `enable_uart=1` locked the clock.

## P6 soldered-proto validation (2026-06-28) — step 1 PASSED on the SOLDERED board

`Rin` re-tuned + locked at **1 k**, the protoboard soldered to match the validated breadboard, and the full
step-1 gate re-run end-to-end on the soldered board: **clean 14-field lines → cementer ingest → live chart
over WiFi.** Step 1 is DONE; the arc moves to the field.

**`Rin` re-tune method (the right way to gauge it):**
- Gauge at the **real DAQ amplitude with a bench PSU**, not the weaker Waveshare. Static inject +6.35 V across
  the input through `Rin`; DMM on Vo. At **1 k**: Vo swings **3.3 V (idle, LED off) ↔ 0.19 V (driven, LED on)**
  on the breadboard, **↔ 0.059 V** on the soldered board (harder saturation). LED current ~4.9 mA.
- Then confirm **framing** dynamically on the Waveshare (real 19200 8N1 RS-232) → `cat /dev/serial0` clean
  lines. The Waveshare's weaker ~+5 V (~3.5 mA at 1 k) is conservative: if it frames clean on the bench, the
  +6.35 V field line has more margin. **1 k locked** — lowest tap load for coexistence.

**Bench mock for the soldered board (before connecting the real Pi):** two PSUs supply the Pi side — **3.3 V**
for the Vo pull-up rail + **5 V** for Vcc, sharing a common (Pi-side) ground. Keep that mock ground isolated
from the DAQ-side ground (the Waveshare GND). This validates the proto's switching standalone, then swap to
the real Pi for the dynamic data gate.

**Findings / gotchas (soldered build — read before the v2 Amphenol build):**
- **Open joint = output stuck HIGH (the P6 time-sink).** A **gap at the DAQ-GND → cathode (pin 3)** solder
  joint left the LED's return path open, so on a positive *space* the LED couldn't conduct → Vo stuck at 3.3 V,
  no switching. Everything measured *correct at idle* (VE=5 V, Vcc=5 V, the 1N4148 clamped the mark to −0.68 V
  on pin 2) because the **mark path through the antiparallel diode was intact** — only the **space path through
  the LED** needs a closed loop the mark path doesn't. **Lesson: if a soldered opto clamps the mark correctly
  but won't switch on space, check LED-return (cathode→GND) continuity first.** Fix: bridge the gap → Vo
  immediately dropped to 0.059 V.
- **Continuity-mode red herring.** A **1 k resistor reads ~1 k in resistance mode but does NOT beep in
  continuity mode** (beep threshold ~30–50 Ω). Don't read "no beep" as "open" on any resistor ≳100 Ω — measure
  resistance. (Cost real time chasing a non-fault on the good pull-up.)
- **Diagnostic decision tree that worked** (output stuck high, line driven positive): probe **anode (pin 2)
  vs DAQ-GND** under drive → ~+1.5 V = LED conducting (fault is output-side / socket / chip); ~line-voltage,
  no drop = LED loop open (cold joint / open return); ~0.69 V = 1N4148 wired parallel. Check **VE↔Vcc tie**
  (pin 7 = 5 V) for a stuck-high before suspecting a DOA chip.

## Open items / risks

- ~~**#1 — DAQ TXD idle voltage**~~ -- DONE, **MEASURED 2026-06-25:** Intellisense **-6.35 V** (pin1=GND,
  pin2=TXD; **transmit-only 2-wire**, no handshake pins), Totco **-8.20 V** (pin5=GND, pin2=TXD). Both
  <+-10 V -> P6KE12CA adequate; `Rin` = 1 kohm / 1.5 kohm. See "DAQ behavior + measurements" above.
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
