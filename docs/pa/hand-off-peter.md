# Hand-off — Peter (live)

`as of: P4 close · 2026-06-25 · operator: peter` (machine: **Windows field laptop** — `C:\Users\poliv\Documents\GitHub\cementer`)

> Peter's per-operator hand-off (multi-operator partition, S6). Optimize for Peter's next-session pickup.
> Peter's PA rewrites this at his wraps; Bryan does not edit it (CODEOWNERS -> Peter).
>
> **Session-start order (multi-operator):** fetch -> `git worktree add .coord coord` (if absent) -> read
> `.coord/ledger.md` tail + `.coord/claims/bryan.md` + `.coord/inbox/peter/` -> THEN this file +
> `status.md`/`changelog.md` on `main`. Shared truth = `status.md` + `changelog.md`; live coordination =
> the coord branch.

## > P4 close (2026-06-25) -- serial-split BUILD resumed; `#1` measured both DAQs; Intellisense channel ready to solder

**RESUME POINT for P5 (tomorrow):** the operator is physically building **Intellisense channel 1** and
running the **bench gate (step 1)**. Pick up by asking what the Pi saw at the gate -- clean 14-col lines /
garbage / silence -- then proceed to **step 2** (real wire on the Pi). Full design + new findings are folded
into [`docs/changes/serial-split-tap/scope.md`](../changes/serial-split-tap/scope.md).

### `#1` MEASURED -- the blocker is cleared
Operator measured both DAQ TXD idle voltages (multimeter, TXD vs GND; reads negative = RS-232 mark):

| Unit | GND / TXD pins | Idle (mark) | `Rin`=(V-1.5)/5mA | Pick | Read baud |
|---|---|---|---|---|---|
| **Intellisense** | pin1 / pin2 -- **transmit-only, 2-wire** (no handshake pins active) | **-6.35 V** | 970 ohm | **1 kohm** | 19200 8N1 |
| **Totco** | pin5 / pin2 | **-8.20 V** | 1.34 kohm | **1.5 kohm** | 9600 8N1 |

- Pull-up `Rpu` = **1 kohm -> 3.3 V** (both). To 3.3 V NOT 5 V (Pi not 5 V-tolerant).
- **TVS P6KE12CA covers BOTH** (both lines <+-10 V) -- field hardening only, skip on the bench.
- Resistors in hand: BOJACK 1000-pc 25-value kit (1ohm-1Mohm) + a "ja90002x300" kit -> 1k & 1.5k stocked.

### NEW FINDING -- Totco TX is **DTR-gated** (not command-polled)
Evidence: pin 2 (Totco TXD) sits at -8.2 V mark whenever the unit is powered (even USB unplugged) -> its
transmitter is **always alive**. DATA appears on pin 2 only when the consumer software runs, and exactly
then **pin 4 -> +9.25 V (DTR asserted)** while **pin 3 (RXD) stays idle mark -- no command bytes ever go
in.** So the Totco streams **only while the consumer asserts DTR**, not in response to a command.
- **Listen-tap implication:** perfect in **coexistence** (existing consumer holds DTR -> Totco streams ->
  we listen). But a **Pi-only standalone read sees silence** unless the Pi asserts DTR. -> **Totco validates
  via the COEXISTENCE test (step 3), not the Pi-only step 2.** (Intellisense, transmit-only, is standalone.)
- **Decisive confirm test (operator):** disconnect consumer, jumper **pin 4 -> +5..9.25 V**, watch pin 2.
  Streams = confirmed; silence = theory wrong, dig further. Also likely explains the **S3 "total silence on
  COM6"** (nothing was asserting DTR).

### Build plan -- Intellisense channel FIRST
Build/validate Intellisense single-channel before adding Totco (Intellisense = the sure thing; Totco has the
unconfirmed DTR behavior; separate input domains = cleaner isolation). A "2-in-1" (two opto channels on one
board) is electrically just **2x the identical circuit** and buildable with parts in hand -- but de-risk by
proving channel 1 first.
- **Wiring + the inviolable rule (DAQ-GND != Pi-GND, gap down the board): scope.md "The circuit" + build
  sheet there.** `Rin` 1k -> 6N137 pin2(anode); pin3(cathode)->DAQ GND; 1N4148 antiparallel; Pi side
  pin8(Vcc)->Pi 5V, **pin7(VE)->pin8** (or output disabled), pin5->Pi GND, 0.1uF pin8->pin5, pin6(Vo)->`Rpu`
  1k->3.3V and pin6->Pi pin10 (GPIO15/RXD).
- **Bench fake-DAQ = the field DB9->USB adapter run as a TRANSMITTER** (operator has NO Waveshare). Laptop
  replays a captured Intellisense `.bin` out the adapter COM port @19200; pick it off the **Jienk DB9
  terminal breakout** at the adapter's **TXD = DB9 pin 3** (NOT pin 2 -- pin 2 was the *read* side in the
  field) + GND pin 5 -> opto input. Read on Pi `/dev/serial0` @19200. **Gate = clean 14-col ASCII.**
- **v2 final form factor:** 6-pin Amphenol -> splitter protoboard (data+GND **pass straight through** to a
  2nd Amphenol that continues the normal run) -> opto branch off the same node -> Pi. Pass-through =
  continuous wire, so the consumer's line is electrically unchanged except the opto's ~5 mA tap load (= the
  step-3 coexistence test). **v2 prereq: map the 6-pin Amphenol pinout (data + GND) before cutover.**

### Resume = scope.md "Build & test plan" -- 3 go/no-go gates
1. solder + bench replay (above) -> 2. real-wire on Pi (`cementer -source /dev/serial0 -format intellisense`,
watch `/debug/stats` rows climb -- **never yet proven on real wire**) -> 3. coexistence (tap in parallel with
the live consumer; Pi powered / unpowered / physically yanked -> **zero disturbance** to production).
Pi UART: `raspi-config` -> serial hardware **ON**, console **OFF**; device `/dev/serial0` (`ttyAMA0`).

## ! OPEN -- for P5
1. **Pending doc PR (unmerged).** This wrap's docs sit on **`peter/p3-doc-currency`** (`b66010b` P3 + the
   P4 commit), **LOCAL / UNPUSHED** (bare wrap). The P3 PR was deferred by the operator "to fold in more
   progress" -- P4 is that progress. **Next: push the branch + open ONE PR -> `main`** (needs operator auth).
   Carries Peter's P3+P4 session bookkeeping + the scope.md update; **no source code.**
2. **Serial-split build** -- in the operator's hands (soldering + bench gate). Resume per above.
3. **Totco** -- second channel after Intellisense proves out; run the **pin-4 DTR jumper confirm test**; map
   the 6-pin Amphenol pinout for v2.

## Coord state (the `.coord` worktree -- RETAINED across sessions on purpose)
- `.coord` worktree on branch `coord`, pushed/synced (P4 **open** + **close** blocks appended this session).
- `claims/peter.md` reset to **idle** at this close.
- **Do NOT remove the `.coord` worktree** -- it's the live coordination channel. (Recreate on a fresh clone
  with `git worktree add .coord coord`.)
- **Bryan:** idle, B6 closed cleanly. Next arc = nav-maps regen (stale since S5) + broaden the pre-commit
  gate to catch deletions. **No contention** with the serial-split hardware arc.

## Environment caveats (Windows field laptop) -- IMPORTANT, reusable
- **This clone path:** `C:\Users\poliv\Documents\GitHub\cementer` (prior hand-offs cited `C:\Users\pjoli\...`
  -- same operator; account/path label differs, not load-bearing).
- **Toolchain:** Go + Node on the machine PATH. The **Bash _tool_ runs non-interactively** -- if `go`/`npm`
  don't resolve, prepend: `export PATH="/c/Program Files/Go/bin:/c/Program Files/nodejs:$PATH"`.
- **`make` is NOT installed** -- run recipes directly: `make hooks`->`git config core.hooksPath
  scripts/git-hooks`; `make coord`->`git worktree add .coord coord`; `make web`->`cd web && npm install &&
  npm run build`; `make build`->web build + `go build ./cmd/cementer`; `make demo`/`run`->
  `./cementer.exe -source testdata/... -format ... -data-dir <tmp> -addr :8080`.
- **`gh` is NOT installed** -- drive GitHub via REST with the cached git token (`git credential fill`).
  External writes (issues/PRs) need explicit user OK.
- **Headless UI verify:** `cd /tmp/pw && PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1 npm i playwright@1.60.0`, then
  `chromium.launch({channel:'msedge', headless:true})` (system Edge -- no download); `page.pdf()` = real PDF.
- **Git (this clone):** `core.autocrlf=false`, `core.hooksPath=scripts/git-hooks`, remote HTTPS, credential
  cached. **`.gitattributes` (`* text=auto eol=lf`) now on `main`** (Bryan PR #10) -- coord commits still
  warn "LF->CRLF"; benign for `.md`.
- **Avoid em-dashes in `curl -d` JSON** -- the shell mangles them -> GitHub "Problems parsing JSON". Use
  ASCII / a JSON file / node `JSON.stringify`.

## State as of P4 close
| Item | State |
|---|---|
| `main` | `ac2dd16` (synced; this laptop ff'd from **22 behind** at session open) |
| Active arc | `serial-split-tap` **BUILD** -- Intellisense channel, in operator's hands |
| `#1` (the blocker) | DONE -- **MEASURED** both DAQs (Intellisense -6.35 V / Totco -8.20 V) |
| Pending docs | `peter/p3-doc-currency` (P3+P4), **LOCAL/UNPUSHED** -> push+PR next (operator auth) |
| coord | P4 open+close pushed; `claims/peter` idle |
| Phase 4b / MVP | DONE (Bryan PR #1) |
| Tests | docs-only session (zero source change) -- `go vet`/`test ./internal/...` green (see status.md) |
| Bryan | idle (B6 closed); next = nav-maps regen + gate-broaden |
