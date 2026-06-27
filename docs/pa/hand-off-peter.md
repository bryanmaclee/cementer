# Hand-off — Peter (live)

`as of: P5 close · 2026-06-27 · operator: peter` (machine: **Windows field laptop** — `C:\Users\poliv\Documents\GitHub\cementer`; Pi: **CementSerial** @ `10.0.0.105`, user `serial123`)

> Peter's per-operator hand-off (multi-operator partition, S6). Optimize for Peter's next-session pickup.
> Peter's PA rewrites this at his wraps; Bryan does not edit it (CODEOWNERS -> Peter).
>
> **Session-start order (multi-operator):** fetch -> `git worktree add .coord coord` (if absent) -> read
> `.coord/ledger.md` tail + `.coord/claims/bryan.md` + `.coord/inbox/peter/` -> THEN this file +
> `status.md`/`changelog.md` on `main`. Shared truth = `status.md` + `changelog.md`; live coordination =
> the coord branch.

## > P5 close (2026-06-27) -- serial-split tap PROVEN end-to-end on breadboard (step-1 bench gate PASSED)

**RESUME POINT for P6:** the Intellisense opto tap **works end-to-end on the breadboard** — proven all the
way: PC sender -> Waveshare RS-232 -> 6N137 opto -> Pi mini-UART -> cementer -> SQLite -> **live chart over
WiFi**. The build is still on **breadboard**; the remaining arc is (a) re-tune `Rin`, (b) solder the proto +
re-run the bench gate, (c) take it to the field (steps 2-3). **Full working recipe + every gotcha is in
[`docs/changes/serial-split-tap/scope.md`](../changes/serial-split-tap/scope.md) "P5 bench validation" —
read that first.**

### What's proven (step-1 bench gate)
`/debug/stats` climbed 208 -> 1079 rows (~14 rows/s = ~1 line/s x 13 channels); live chart painted at
`http://10.0.0.105:8080`. The whole electrical + software path is validated on breadboard with a **Waveshare
USB->RS232** as the bench DAQ source (operator acquired one mid-session; this **supersedes** the field-DB9
-adapter plan and a briefly-considered ESP32-TTL "Option B"). Real-RS-232 path => `Rin`~1k-class, **no
inversion**, 1N4148 active.

### Next actions for P6 (in order)
1. **Re-tune `Rin` UP with the good chip.** Bench settled at **560 Ω**, but that was reached while a **DOA
   6N137 was masking the real margin**. The Waveshare is a *weaker* driver (~+5 V space) than the real DAQ
   (+6.35 V), so size `Rin` on the weak bench source then the field has margin. Step `Rin` up (680 -> 820 ->
   1 k) to the highest value that still switches Vo solidly -> minimizes field tap load (coexistence). Verify
   with the `0x00` flood (DMM-visible) or the live chart.
2. **Solder the protoboard** to match the validated breadboard, then **re-run step 1** (same recipe).
   Terminate the input as two labeled leads ("TXD-IN"/"GND-IN") — bench lands them on the Waveshare DB9
   pin3/pin5; field lands them on the DAQ DB9 **pin2(TXD)/pin1(GND)** via the Jienk breakout.
3. **Field steps 2-3:** step 2 = real wire, Pi-only (`cementer -serial /dev/serial0 -baud 19200 -format
   intellisense`; never yet proven on a real pump wire). Step 3 = coexistence (tap in parallel with the live
   consumer; Pi powered/unpowered/yanked -> zero disturbance). Watch the higher tap-load (~7-9 mA) here.
4. **The chart is a 60-sec re-do if you want it again:** `scp cementer-arm64-new` to the Pi (stop the
   running one first!), start it, open `http://10.0.0.105:8080`, run `tools/intellisense-send.ps1 -Port COM6`.

### Findings that cost real time (don't re-pay)
- **DOA 6N137.** First chip's output stage was dead — LED driven ~6 mA, Vcc/VE/GND all good, Vo stuck at
  3.3 V. A spare fixed it instantly. **Test each opto.** Diagnostic: a continuous `0x00` flood holds the
  line ~90% positive (DMM-visible); FTDI **`BreakState` does NOT transmit a break** — use the flood.
- **1N4148 orientation.** Parallel-with-LED clamps the anode at 0.69 V (LED never lights). Antiparallel
  (band/cathode -> pin 2) clamps the idle negative mark to -0.68 V — correct.
- **Pi 4 baud trap.** `/dev/serial0 -> ttyS0` (mini-UART); a reboot/console resets it to **9600** -> garbage
  at 19200. Fix: `sudo raspi-config` serial **console OFF + hardware ON** (the latter sets `enable_uart=1`,
  locking the core clock so 19200 is accurate — Bluetooth-disable/PL011 trick NOT needed). cementer sets its
  own port baud regardless (ignores stty).

## ! OPEN -- for P6
1. **ALL PUSHES DEFERRED (operator instruction P5 close: "wrap local-only, leave the push for tomorrow").**
   - **Feature branch `peter/p3-doc-currency`** now carries **P3 + P4 + P5** docs (committed locally this
     wrap, **UNPUSHED**). Next: `git push` the branch + open **ONE PR -> `main`** (needs operator auth).
     No source code — PA docs + `serial-split-tap/scope.md` updates + `tools/intellisense-send.ps1` + a
     `.gitignore` line.
   - **Coord close UNPUSHED too.** The P5 open+close block is appended to `.coord/ledger.md` and committed
     **locally** on the `coord` branch, but **NOT pushed** (operator deferred coord push as well —
     unusual; normally coord pushes direct even on a no-push wrap). **Next session: `git push origin coord`
     first thing** so Bryan sees P5. No contention tonight (Bryan idle; hardware/docs arc).
2. **Serial-split build** — on breadboard, working. Resume per "Next actions for P6" above.
3. **Totco** — second channel after Intellisense proto proves out; run the pin-4 DTR jumper confirm test;
   map the 6-pin Amphenol pinout for v2.

## Coord state (the `.coord` worktree -- RETAINED across sessions on purpose)
- `.coord` worktree on branch `coord`. P5 open+close block appended + committed **LOCALLY**; **NOT pushed**
  (see OPEN #1). `claims/peter.md` reset to **idle** at this close.
- **Do NOT remove the `.coord` worktree** -- it's the live coordination channel. (Recreate on a fresh clone
  with `git worktree add .coord coord`.)
- **Bryan:** idle, B6 closed cleanly. Next Bryan arc = nav-maps regen (stale since S5) + broaden the
  pre-commit gate to catch deletions. **No contention** with the serial-split hardware arc.

## Environment caveats (Windows field laptop + Pi) -- IMPORTANT, reusable
- **This clone path:** `C:\Users\poliv\Documents\GitHub\cementer`. **Pi:** `serial123@10.0.0.105`
  (`CementSerial`); no Go/Node/repo on the Pi -> cross-compile on the laptop + `scp`.
- **Git config drifts on this clone (re-check at session start):** P5 found `core.hooksPath` **unset** and
  `core.autocrlf=true` (both drifted from the documented `scripts/git-hooks` / `false`). Restored at P5
  start. Probe: `git config core.hooksPath` + `git config core.autocrlf`.
- **Node:** was **18.12.1** at P5 start (P1's "Node 24" had not stuck); Vite needs 20+. Upgraded via
  `winget install OpenJS.NodeJS.LTS` -> **24.18.0**. `where.exe node` -> `C:\Program Files\nodejs\node.exe`.
- **PowerShell execution policy:** set **`Set-ExecutionPolicy -Scope CurrentUser RemoteSigned`** at P5 (was
  blocking `.ps1` scripts incl. npm + the sender every new window). Should persist now.
- **`web/dist` is gitignored + was a stale 315-byte placeholder** until P5 rebuilt it (`cd web && npm run
  build`). Rebuild it before any cross-compile that needs a working web UI (the SPA is `//go:embed`-ed).
- **Cross-compile recipe (laptop PowerShell):** `$env:GOOS='linux'; $env:GOARCH='arm64';
  $env:CGO_ENABLED='0'; go build -o cementer-arm64-new ./cmd/cementer` then `$env:GOOS=''; $env:GOARCH='';
  $env:CGO_ENABLED=''`. CGO-free (modernc SQLite). **`scp cementer-arm64-new serial123@10.0.0.105:~/`** —
  note the **`:~/`** (colon!) and **stop the running binary on the Pi first** (live ELF = "text file busy").
- **Shell confusion is the #1 time-sink this session:** laptop = PowerShell (`$env:`, `.\`, `scp`, drive
  letters); Pi = bash (`~/`, `kill`, `pkill`, `chmod`, `cat`, `stty`, `&`). Background a Pi process with
  `... > ~/cementer.log 2>&1 &`; curl in a *second* shell or it blocks.
- **`make`/`gh` NOT installed** on the laptop; run recipes directly; drive GitHub via REST with the cached
  git token (external writes need operator OK).
- **Avoid em-dashes in `curl -d` JSON** (shell mangles them). Toolchain (Go/Node) on machine PATH; the Bash
  _tool_ is non-interactive (prepend PATH export if go/npm don't resolve).

## State as of P5 close
| Item | State |
|---|---|
| `main` | `ac2dd16` (unchanged this session) |
| Active arc | `serial-split-tap` — **step-1 bench gate PASSED on breadboard**; next = Rin re-tune + solder + field |
| Bench source | **Waveshare USB->RS232** (real RS-232) + `tools/intellisense-send.ps1`; COM6 |
| Pi binary | `~/cementer-arm64-new` (cross-compiled P5; old `~/cementer-arm64` was stale, no `-format`) |
| Pi data | `~/cementer-splittest/` (1079+ rows proven); cementer may still be running -> `pkill -f cementer-arm64-new` |
| Feature branch | `peter/p3-doc-currency` = P3+P4+P5 docs, **committed LOCAL / UNPUSHED** -> push + ONE PR next (auth) |
| coord | P5 open+close committed **LOCAL / UNPUSHED** -> `git push origin coord` first thing P6; `claims/peter` idle |
| New tool | `tools/intellisense-send.ps1` (committed); `cementer-arm64*` now gitignored |
| Tests | docs+tooling arc (zero source change) -- `go vet`/`go test ./...` recorded (see status.md) |
| Bryan | idle (B6 closed); next = nav-maps regen + gate-broaden |
