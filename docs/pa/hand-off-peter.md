# Hand-off — Peter (live)

`as of: P1 close · 2026-06-21 · operator: peter` (machine: **Windows field laptop** — `C:\Users\pjoli\Documents\GitHub\cementer`)

> Peter's per-operator hand-off (multi-operator partition, S6). Optimize for Peter's next-session pickup.
> Peter's PA rewrites this at his wraps; Bryan does not edit it (CODEOWNERS → Peter).
>
> **Session-start order (multi-operator):** fetch → `git worktree add .coord coord` → read
> `.coord/ledger.md` tail + `.coord/claims/bryan.md` + `.coord/inbox/peter/` → THEN this file +
> `status.md`/`changelog.md` on `main`. Shared truth = `status.md` + `changelog.md`; live coordination =
> the coord branch.

## ▶ P2 in-flight (2026-06-21) — serial-split tap (hardware design)

Active arc: **`serial-split-tap`** — an isolated, listen-only serial tap so the Pi 4B can ingest a
live DAQ stream without disturbing the system that already consumes that serial. **Design is locked;
build is paused on one measurement.** Full spec:
[`docs/changes/serial-split-tap/scope.md`](../changes/serial-split-tap/scope.md).

- **Topology:** 6N137 opto front-end → **Pi GPIO UART** (bypassing the USB-serial adapter). Input
  self-powered by the line; output pulled to **3.3 V** (NOT 5 V — would fry the Pi). Polarity is
  correct without inversion.
- **Blocker = "#1":** the DAQ TXD idle voltage (multimeter). Operator gathering it "in a day or two."
  It sets the input resistor (≈±5 V→680 Ω, ±9 V→1.5 kΩ, ±12 V→2.2 kΩ) and the TVS rating.
- **Parts ordered** (6N137 ×N, DIP sockets, 1N4148, resistors, P6KE12CA TVS); rest in hand.
- **TVS caveat:** P6KE12CA clips a full ±12 V line — use P6KE15/18CA if #1 shows ≥±10 V. Not needed
  for the bench build.
- **Resume = 3 steps** (scope doc §Build & test): solder → bench replay → real-wire on Pi → coexistence.
- **Open Qs to confirm with operator:** (a) #1 value; (b) one-way link (consumer never TX's to DAQ?);
  (c) Pi-GPIO vs keep-Waveshare output path (GPIO chosen, confirm).
- Branch: `peter/p2-serial-split-scope` (this scope doc). coord: P2 claim = this arc.

## ✅ P1 result (2026-06-21)

Adopted the S6 multi-party model + stood up this laptop + verified Phase 4b. Everything that could land
**did land** (`main` `0a96095`); everything still pending is **blocked on Bryan's repo-config**, not on me.

- **Adopted** PR-flow + coord + meta-doc partition. Fast-forwarded to Bryan's PR #1 (`c952c54`).
- **Windows toolchain UP:** Go 1.26.4 + Node 24.17.0/npm 11.13.0 (winget). Commit gate installed
  (`core.hooksPath=scripts/git-hooks`). Fixed the CRLF/gofmt break (`autocrlf=false` + LF renormalize).
  Full gate green on Windows.
- **Phase 4b PA-verified E2E** (built/ran/recorded → Report tab + print render via headless Edge).
- **P1 onboarding docs landed:** PR #2 → `main` `0a96095`.
- **Bryan notified** (issue #3 + coord `inbox/bryan/` notice) of the ruleset problems.

## ⚠ OPEN — next session

1. **`.gitattributes` durable CRLF fix** — `* text=auto eol=lf` (+ maybe `*.go text eol=lf`) so no future
   Windows clone hits the gofmt break. A `peter/<arc>` PR; touches Bryan's clone → coordinate.
2. ~~**`pa.md` topology rewrite**~~ — ✅ **DONE by Bryan (PR #6 `42ef5f2`, `da33524`)**: overlay v2
   (multi-operator / PR-flow / coord) + the `hand-off-bryan.md` / `user-voice-bryan.md` rename +
   CODEOWNERS. No longer Peter's.
3. **Bryan's coord state is stale** — `claims/bryan.md` still reads `active` (B6) and the ledger has no
   B6 close block, yet all his B6 work merged to `main`. His single-operator-owned file; not mine to
   edit — surfaced only. No active contention.

_Resolved during P1 (no longer open):_ Bryan fixed the over-broad ruleset (**issue #3 closed, verified**)
→ `coord` pushes + merged-branch deletion both work now; P1 onboarding (PR #2) + wrap (PR #4) landed on
`main`; coord pushed + synced; merged `peter/*` branches cleaned up.

## Project work (none claimed by Peter; project MVP effectively reached)

- **Phase 4b is DONE** (Bryan, PR #1) — verified here. Do not rebuild.
- Parser cleanup (`internal/parser` off-path); Totco preset (unit not accessible). Both low-priority.

## Environment caveats (Windows field laptop) — IMPORTANT, reusable

- **Toolchain:** Go + Node are on the **machine PATH** → a freshly-launched shell auto-resolves them. But
  the **Bash _tool_ runs non-interactively** and a session's shell may predate that; if `go`/`npm` don't
  resolve, prepend: `export PATH="/c/Program Files/Go/bin:/c/Program Files/nodejs:$PATH"`.
- **`make` is NOT installed.** Run recipes directly: `make hooks` → `git config core.hooksPath
  scripts/git-hooks`; `make coord` → `git worktree add .coord coord`; `make web` → `cd web && npm install
  && npm run build`; `make build` → web build + `go build ./cmd/cementer`; `make demo`/`make run` → run
  `./cementer.exe -source testdata/... -format ... -data-dir <tmp> -addr :8080`.
- **`gh` is NOT installed.** Drive GitHub via the REST API with the cached token:
  `cred=$(printf "protocol=https\nhost=github.com\n\n" | git credential fill); TOKEN=$(... sed -n 's/^password=//p')`.
  (Used this for the PR #2 merge and issue #3.) External writes (issues/PRs) need explicit user OK.
- **Headless UI verify on Windows:** `cd /tmp/pw && PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1 npm i
  playwright@1.60.0`, then `chromium.launch({channel:'msedge', headless:true})` (system Edge — no browser
  download). `page.pdf()` produces the actual Save-as-PDF artifact.
- **Git config (this clone):** `core.autocrlf=false`, `core.hooksPath=scripts/git-hooks`. Remote = HTTPS,
  credential cached (credential.helper=manager; the one-time prompt is done).
- **Avoid em-dashes in `curl -d` JSON bodies** — the shell mangles them → GitHub "Problems parsing JSON".
  Use ASCII, a JSON file, or node + `JSON.stringify`.

## Coord state (the .coord worktree — RETAINED across sessions on purpose)

- `.coord` worktree on branch `coord`, **pushed + synced** (`origin/coord = local = d1028bc`).
- My `claims/peter.md` = **idle** (reset at P1 close). My P1 ledger has open + close + addendum blocks.
- **Do NOT remove the `.coord` worktree** — it's the live coordination channel (now works both ways).

## State as of P1 close

| Item | State |
|---|---|
| `main` | `a854b38` (P1 onboarding + wrap merged; synced) |
| Multi-party model | ADOPTED; full PR-flow cycles done (#2, #4) |
| Windows toolchain | ✅ Go 1.26.4 + Node 24.17.0; full gate green; CRLF fixed |
| Commit gate | ✅ installed |
| Phase 4b | ✅ DONE (Bryan PR #1), PA-verified E2E here |
| P1 onboarding + wrap docs | ✅ landed (PR #2 + PR #4) |
| coord | ✅ pushed + synced (`d1028bc`); works both ways |
| Ruleset blocks | ✅ RESOLVED + verified (issue #3 closed) |
| Tests (P1 wrap) | `go test`/`vet`/`gofmt`/`go build`/web build all ✅ |
