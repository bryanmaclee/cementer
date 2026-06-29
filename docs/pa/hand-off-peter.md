# Hand-off — Peter (live)

`as of: P7 close · 2026-06-29 · operator: peter` (machine: **field laptop** — hostname `P-Tech1`, user `pjoli`, clone `C:\Users\pjoli\Documents\GitHub\cementer`; Pi: **CementSerial**, user `serial123`, now reached at `cementserial.local` over the phone hotspot — IP is network-dependent, no longer a fixed `10.0.0.105`)

> Peter's per-operator hand-off (multi-operator partition, S6). Optimize for Peter's next-session pickup.
> Peter's PA rewrites this at his wraps; Bryan does not edit it (CODEOWNERS -> Peter).
>
> **Session-start order (multi-operator):** fetch -> `git worktree add .coord coord` (if absent) -> read
> `.coord/ledger.md` tail + `.coord/claims/bryan.md` + `.coord/inbox/peter/` -> THEN this file +
> `status.md`/`changelog.md` on `main`. Shared truth = `status.md` + `changelog.md`; live coordination =
> the coord branch.
>
> **Machine note (P7):** there are **three** machines — the **garage/shop desktop** (`poliv`, did P3-P6),
> **this field laptop** (`P-Tech1`/`pjoli`, doing field work), and the **cementer laptop** (the production
> DAQ consumer, runs the Intellisense job software, **never runs Claude**). Auto-ID at session start with
> `hostname`+`whoami`. The P6 hand-off mislabeled the garage desktop as "field laptop" — corrected here.

## > P7 close (2026-06-29) -- Intellisense DB9 split-off FIELD-VERIFIED (steps 2 + 3 PASSED on a real DAQ)

**RESUME POINT for P8 = the v2 Amphenol PASS-THROUGH PROTOTYPE.** The isolated tap is now proven on an
**actual Intellisense pump DAQ** (not just the bench): real wire -> opto -> Pi mini-UART -> cementer -> SQLite
-> **live chart over WiFi**, clean 14-field lines, **and zero disturbance to the production consumer** (Pi
powered + unpowered). Roadmap step 1 (DB9 field test) is DONE. Full P7 recipe + every gotcha is in
[`docs/changes/serial-split-tap/scope.md`](../changes/serial-split-tap/scope.md) "P7 field validation" — read
that first.

### Operator's roadmap (directive — Intellisense parallel-splitter MVP before Totco)
1. ~~**Field test the Intellisense DAQ unit** (DB9 split-off)~~ ✅ **DONE P7 2026-06-29.**
2. **Build the v2 hardware prototype with the Amphenol connectors** — the **permanent inline pass-through**
   form factor (data + GND pass straight through on **passive continuous conductors** to a 2nd Amphenol that
   continues the normal run; opto branches off the same node). **Stays inline + broadcasts WiFi in parallel —
   NOT a removable branch.** Prereq: **map the 6-pin Amphenol pinout** (data + GND).
3. **Garage-test the Amphenol prototype** through this same bench gate, **then field test** it.
4. **Intellisense parallel-splitter MVP done BEFORE Totco.** Totco stays deferred until then.

### What's proven (P7 field test — the real wire)
- **Real DB9 pinout (this Intellisense unit): GND = pin 5, TXD = pin 2**, idle/active **-5.66..-5.22 V** (wobble
  = actively transmitting). Differs from the earlier #1 probe ("pin 1 = GND") — field unit uses standard DB9
  pin-5 ground; trust the live reading.
- `Rin` = **1 k** frames clean at the -5.5 V field line (~4 mA) — the +6.35 V tune has margin to spare.
- **Coexistence PASSED:** consumer (cementer laptop) clean with the Pi tapping in parallel, **powered and
  unpowered**. The ~5 mA opto load is harmless; the optical barrier keeps Pi-side activity off the line; the
  consumer survives a dead Pi side (the basis for the permanent-inline v2 design).

### P7 findings that cost real time (don't re-pay — full detail in scope.md "P7 field validation")
- **DMM is the wrong instrument on a live data line.** Opto Vo "bumping 3.3 -> 3.06 V" is the **good** signature
  (mostly-idle ~1 line/s stream; DMM can't resolve the bursts), NOT a fault. The P6 open-joint fault was a
  *dead-solid* 3.3 V with zero movement. The `0x00`-flood DMM trick is **unavailable with a real DAQ** -> use
  the **UART decode (`cat /dev/serial0`)** as the gate, not the meter.
- **WiFi in the field with no editable supplicant:** pull the microSD -> mount the FAT32 `boot`/`bootfs` on
  Windows (**ignore the "format this disk" prompt for the ext4 root — never format**) -> drop a
  `wpa_supplicant.conf` in the boot root with multiple `network={}` + `priority=` + **`country=US`** (mandatory
  or the radio stays off) -> eject/reinsert/boot -> Pi auto-joins. Worked -> Pi is **not** Bookworm. Then put
  Pi + laptop on the same phone hotspot; reach via `cementserial.local`.
- **`ERR_CONNECTION_REFUSED` on `:8080` = host reachable, cementer not running** (active refusal, not a network
  failure). SSH in + start it; a reboot clears the prior instance.

## ! OPEN -- for P8
1. **v2 Amphenol pass-through prototype** is the next build (above). Prereq: map the 6-pin Amphenol pinout
   (data + GND). Then garage-gate -> field.
2. **Totco — DEFERRED** until the Intellisense parallel-splitter MVP is done (operator directive). When
   resumed: same circuit, 2nd 6N137 channel, 9600 8N1, `Rin` 1.5 k; DTR-gated so validate via coexistence;
   run the pin-4 DTR jumper confirm test; map the 6-pin Amphenol pinout.
3. **PR `peter/p3-doc-currency -> main`** carries P3+P4+P5+P6+P7 docs — **OPEN, refreshed at P7 push**; still
   **unmerged** (CODEOWNERS routes it to Peter; merge when the operator authorizes).

## Coord state (the `.coord` worktree -- RETAINED across sessions on purpose)
- `.coord` worktree on branch `coord`. Fast-forwarded to `4137f96` at P7 start, then the **P7 close block
  pushed direct** this wrap. `claims/peter.md` reset to **idle**.
- **Do NOT remove the `.coord` worktree** -- it's the live coordination channel. (Recreate on a fresh clone
  with `git worktree add .coord coord`.)
- **Bryan:** claim idle (B6 closed). New peer branch **`bryan/flobase-reconcile`** appeared (1 commit,
  `bb2f3f8`) — `/flobase` tooling only (`.claude/CLAUDE.md` fenced region + `.pa-base/profile`); **no source /
  no PA-doc change**, **no coord OPEN block** for it yet. **No contention** with the serial-split arc — not
  Peter's to drive.

## Environment caveats (field laptop `P-Tech1` + Pi) -- IMPORTANT, reusable
- **This clone path:** `C:\Users\pjoli\Documents\GitHub\cementer` (field laptop). The **garage desktop** is
  `C:\Users\poliv\...` (separate clone; sync only via `origin` — P7 started 3 commits behind, fast-forwarded).
- **Toolchain here:** Go **1.26.4**, Node **24.17.0**, `web/dist` present (built) — can cross-compile the Pi
  binary if needed. `make`/`gh` **NOT installed**; run recipes directly; drive GitHub via REST with the cached
  git token (external writes need operator OK).
- **Git config (re-check at session start):** `core.hooksPath=scripts/git-hooks` ✅ · `core.autocrlf=false` ✅
  (both correct at P7). Probe: `git config core.hooksPath` + `git config core.autocrlf`.
- **Pi reach (field):** put the Pi + laptop on the **phone hotspot**; `ssh serial123@cementserial.local` +
  browser `http://cementserial.local:8080`. If `.local` won't resolve, grab the Pi IP from the phone's
  connected-devices list. No fixed `10.0.0.105` anymore (that was the garage network).
- **Pi field data:** `~/cementer-fieldtest/` (P7). Bench data was `~/cementer-splittest/`. Binary:
  `~/cementer-arm64-new` (cross-compiled P5, has `-format`). Stop a running one before scp/restart
  (`pkill -f cementer-arm64-new`; a live ELF = "text file busy").
- **Pi UART:** `/dev/serial0 -> ttyS0` (mini-UART). console **OFF** + hardware **ON** (`enable_uart=1` locks
  the clock so 19200 is accurate). A reboot can reset it to the 9600 trap -> garbage at 19200; re-check.
  cementer sets its own port baud via the `-baud` flag (defaults **9600** -> pass `-baud 19200`).
- **Cross-compile recipe (laptop PowerShell):** `$env:GOOS='linux'; $env:GOARCH='arm64';
  $env:CGO_ENABLED='0'; go build -o cementer-arm64-new ./cmd/cementer` then reset the env vars; CGO-free
  (modernc SQLite). `scp cementer-arm64-new serial123@<pi>:~/` (note the `:~/`). Web UI needs a real
  `web/dist` (`cd web && npm run build`, Node >= 20) embedded at build time.
- **Shell confusion is the #1 time-sink:** laptop = PowerShell (`$env:`, `.\`, `scp`, drive letters); Pi = bash
  (`~/`, `kill`, `pkill`, `stty`, `cat`, `&`). Background a Pi process with `... > ~/cementer.log 2>&1 &`.

## State as of P7 close
| Item | State |
|---|---|
| `main` | `ac2dd16` (unchanged; PR from `peter/p3-doc-currency` OPEN + refreshed this wrap, unmerged) |
| Active arc | `serial-split-tap` — **Intellisense DB9 split-off FIELD-VERIFIED** (steps 2+3 PASSED on a real DAQ) |
| Next build | **v2 Amphenol pass-through prototype** (permanent inline; passive continuous through-wire) |
| Field source | a **real Intellisense DAQ** (DB9 pin 2 = TXD, pin 5 = GND, ~-5.5 V) |
| Proto board | soldered, validated bench (P6) + field (P7); `Rin` locked 1 k |
| Pi binary | `~/cementer-arm64-new`; Pi reached via `cementserial.local` over phone hotspot |
| Pi data | `~/cementer-fieldtest/` (P7 field capture) |
| Feature branch | `peter/p3-doc-currency` = P3+P4+P5+P6+P7 docs, **PUSHED**; PR to `main` OPEN + refreshed, unmerged |
| coord | P7 close block **PUSHED direct**; `claims/peter` idle |
| Tests | hardware/docs arc (zero source change) -- `go vet ./...` ok · `go test ./...` ok (api/daqformat/printcfg/store pass) |
| Bryan | idle (B6 closed); `bryan/flobase-reconcile` appeared (flobase tooling, no contention) |
| Next | **v2 Amphenol pass-through proto** -> garage gate -> field. Intellisense MVP before Totco |
