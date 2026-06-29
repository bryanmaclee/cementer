# Hand-off — Peter (live)

`as of: P6 close · 2026-06-28 · operator: peter` (machine: **Windows field laptop** — `C:\Users\poliv\Documents\GitHub\cementer`; Pi: **CementSerial** @ `10.0.0.105`, user `serial123`)

> Peter's per-operator hand-off (multi-operator partition, S6). Optimize for Peter's next-session pickup.
> Peter's PA rewrites this at his wraps; Bryan does not edit it (CODEOWNERS -> Peter).
>
> **Session-start order (multi-operator):** fetch -> `git worktree add .coord coord` (if absent) -> read
> `.coord/ledger.md` tail + `.coord/claims/bryan.md` + `.coord/inbox/peter/` -> THEN this file +
> `status.md`/`changelog.md` on `main`. Shared truth = `status.md` + `changelog.md`; live coordination =
> the coord branch.

## > P6 close (2026-06-28) -- serial-split tap step-1 gate PASSED on the SOLDERED PROTO; `Rin` locked at 1 k

**RESUME POINT for P7 = FIELD TEST.** The Intellisense opto tap is **soldered, validated, and DONE on the
bench** — proven end-to-end on the *soldered* protoboard (not just breadboard): PC sender -> Waveshare RS-232
-> 6N137 opto -> Pi mini-UART -> cementer -> SQLite -> **live chart over WiFi**, clean 14-field lines. `Rin`
**locked at 1 k** (gauged at the real +6.35 V amplitude). The bench arc is complete; **the next step is the
FIELD TEST of a real Intellisense DAQ unit via the DB9 split-off.** Full recipe + every P6 gotcha is in
[`docs/changes/serial-split-tap/scope.md`](../changes/serial-split-tap/scope.md) "P6 soldered-proto
validation" — read that first.

### Operator's roadmap (directive, P6 — Intellisense parallel-splitter MVP before Totco)
1. **Field test the Intellisense DAQ unit** (DB9 split-off) — prove the soldered proto reads a real pump wire.
2. **Once the DB9 split-off is verified**, build the **v2 hardware prototype with the Amphenol connectors**
   (the pass-through splitter form factor — data + GND pass straight through to a 2nd Amphenol continuing the
   normal run; opto branches off the same node).
3. **Test the Amphenol prototype through this same process in the garage** (bench gate, same recipe), **then
   field test** it.
4. **Get the Intellisense parallel-splitter unit MVP done BEFORE moving forward with Totco.** Totco is
   explicitly deferred until the Intellisense MVP is complete.

### What's proven (step-1 bench gate, both boards)
`Rin` = **1 k** locked. Static +6.35 V PSU inject: Vo swings **3.3 V ↔ 0.059 V** (soldered; 0.19 V breadboard)
— solid saturation, ~4.9 mA tap load. Waveshare dynamic: clean 14-field lines. cementer: `/debug/stats`
climbing + live chart at `http://10.0.0.105:8080`. The full electrical + software path is validated on the
**soldered** board.

### Next actions for P7 (in order — field)
1. **Field step 2 — real wire, Pi-only.** Tap a live Intellisense DAQ DB9 (**pin 2 = TXD, pin 1 = GND**) into
   the soldered proto's TXD-IN/GND-IN leads via the Jienk breakout. `cementer -serial /dev/serial0 -baud 19200
   -format intellisense`; watch `/debug/stats` climb. **Never yet proven on a real pump wire** (bench used the
   Waveshare). ⚠ the device flag is **`-serial`** (NOT `-source`, a replay file) and **`-baud 19200` is
   mandatory** (defaults to 9600; cementer sets the port baud itself).
2. **Field step 3 — coexistence.** Tap **in parallel** with the existing consumer; verify it still reads with
   the Pi powered, unpowered, and physically yanked. **Gate: zero disturbance.** Watch the ~4.9 mA tap load;
   MAX3232 high-Z buffer is the documented fallback if it disturbs the consumer.
3. **v2 Amphenol prototype** (after DB9 split-off verified) — map the 6-pin Amphenol pinout (data + GND);
   build the pass-through splitter board; **garage-test through the same bench gate**, then field.
4. **Re-do the bench chart in ~60 s** if needed: `scp cementer-arm64-new` to the Pi (stop the running one
   first!), start it, open `http://10.0.0.105:8080`, run `tools/intellisense-send.ps1 -Port COM6`.

### P6 findings that cost real time (don't re-pay — full detail in scope.md)
- **Open joint = output stuck HIGH.** A gap at **DAQ-GND -> cathode (pin 3)** broke the LED return loop -> Vo
  stuck at 3.3 V, no switching, even though idle measured perfect (the mark path through the antiparallel
  1N4148 was intact; only the *space* path through the LED was open). **If a soldered opto clamps the mark
  (-0.68 V on pin 2) but won't switch on space, check cathode->GND continuity FIRST.**
- **Continuity-mode red herring:** a 1 k resistor reads ~1 k but does NOT beep in continuity mode (threshold
  ~30-50 Ω). Measure resistance, don't trust the beeper, on anything ≳100 Ω.
- **Gauge `Rin` at the field voltage (+6.35 V PSU), not the weaker Waveshare (~+5 V)** — else you over-spec
  the current. 1 k is right.

## ! OPEN -- for P7
1. **Field test is the next step** (above). Needs a reachable real Intellisense DAQ unit.
2. **v2 Amphenol prototype** after the DB9 split-off is field-verified.
3. **Totco — DEFERRED** until the Intellisense parallel-splitter MVP is done (operator directive P6). When
   resumed: same circuit, 2nd 6N137 channel, 9600 8N1, `Rin` 1.5 k; DTR-gated so validate via coexistence;
   run the pin-4 DTR jumper confirm test; map the 6-pin Amphenol pinout.
4. **(Cleared this wrap)** the P3+P4+P5 push backlog — branch pushed + PR to `main` opened at P6 close.

## Coord state (the `.coord` worktree -- RETAINED across sessions on purpose)
- `.coord` worktree on branch `coord`. P5 close block (was unpushed) + the **P6 close block pushed direct**
  this wrap. `claims/peter.md` reset to **idle** at this close.
- **Do NOT remove the `.coord` worktree** -- it's the live coordination channel. (Recreate on a fresh clone
  with `git worktree add .coord coord`.)
- **Bryan:** idle, B6 closed cleanly. Next Bryan arc = nav-maps regen (stale since S5) + broaden the
  pre-commit gate to catch deletions. **No contention** with the serial-split hardware arc.

## Environment caveats (Windows field laptop + Pi) -- IMPORTANT, reusable
- **This clone path:** `C:\Users\poliv\Documents\GitHub\cementer`. **Pi:** `serial123@10.0.0.105`
  (`CementSerial`); no Go/Node/repo on the Pi -> cross-compile on the laptop + `scp`.
- **Git config drifts on this clone (re-check at session start):** P5 found `core.hooksPath` unset +
  `core.autocrlf=true`; both were **correct at P6 start** (`scripts/git-hooks` / `false`). Probe:
  `git config core.hooksPath` + `git config core.autocrlf`.
- **Node:** 24.18.0 as of P5 (upgraded from 18 via `winget install OpenJS.NodeJS.LTS`); Vite needs 20+.
  `where.exe node` -> `C:\Program Files\nodejs\node.exe`.
- **PowerShell execution policy:** `RemoteSigned` (CurrentUser) set at P5 — lets the `.ps1` sender + npm run.
- **`web/dist` is gitignored** + was a stale 315-byte placeholder until P5 rebuilt it (`cd web && npm run
  build`). Rebuild before any cross-compile that needs a working web UI (the SPA is `//go:embed`-ed).
- **Cross-compile recipe (laptop PowerShell):** `$env:GOOS='linux'; $env:GOARCH='arm64';
  $env:CGO_ENABLED='0'; go build -o cementer-arm64-new ./cmd/cementer` then `$env:GOOS=''; $env:GOARCH='';
  $env:CGO_ENABLED=''`. CGO-free (modernc SQLite). **`scp cementer-arm64-new serial123@10.0.0.105:~/`** —
  note the **`:~/`** (colon!) and **stop the running binary on the Pi first** (live ELF = "text file busy").
- **Pi UART:** `/dev/serial0 -> ttyS0` (mini-UART). `raspi-config` serial **console OFF + hardware ON**
  (`enable_uart=1` locks the core clock so 19200 is accurate). A reboot can reset it to the 9600 trap ->
  garbage at 19200; re-check. cementer sets its own port baud regardless of `stty`.
- **Shell confusion is the #1 time-sink:** laptop = PowerShell (`$env:`, `.\`, `scp`, drive letters); Pi =
  bash (`~/`, `kill`, `pkill`, `chmod`, `cat`, `stty`, `&`). Background a Pi process with
  `... > ~/cementer.log 2>&1 &`; curl in a *second* shell or it blocks.
- **`make`/`gh` NOT installed** on the laptop; run recipes directly; drive GitHub via REST with the cached
  git token (external writes need operator OK).

## State as of P6 close
| Item | State |
|---|---|
| `main` | `ac2dd16` (unchanged; PR from `peter/p3-doc-currency` opened this wrap, awaiting merge) |
| Active arc | `serial-split-tap` — **soldered proto step-1 gate PASSED**; `Rin` locked 1 k; next = FIELD test |
| Bench source | **Waveshare USB->RS232** + `tools/intellisense-send.ps1`; COM6 |
| Proto board | soldered, validated; mock-Pi bench = 3.3 V (pull-up rail) + 5 V (Vcc) PSUs, shared Pi-GND |
| Pi binary | `~/cementer-arm64-new` (cross-compiled P5; embeds rebuilt SPA) |
| Pi data | `~/cementer-splittest/` (proven); cementer may still be running -> `pkill -f cementer-arm64-new` |
| Feature branch | `peter/p3-doc-currency` = P3+P4+P5+P6 docs, **PUSHED**; **PR to `main` OPENED** this wrap (auth) |
| coord | P5 + P6 close blocks **PUSHED direct**; `claims/peter` idle |
| Tests | hardware/docs arc (zero source change) -- `go vet ./...` ok · `go test ./...` ok (see status.md) |
| Bryan | idle (B6 closed); next = nav-maps regen + gate-broaden |
| Next | **FIELD: Intellisense DB9 split-off** -> v2 Amphenol proto -> garage -> field. Totco AFTER Intellisense MVP |
