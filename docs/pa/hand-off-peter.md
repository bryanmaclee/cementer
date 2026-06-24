# Hand-off — Peter (live)

`as of: P3 close · 2026-06-23 · operator: peter` (machine: **Windows field laptop** — `C:\Users\pjoli\Documents\GitHub\cementer`)

> Peter's per-operator hand-off (multi-operator partition, S6). Optimize for Peter's next-session pickup.
> Peter's PA rewrites this at his wraps; Bryan does not edit it (CODEOWNERS → Peter).
>
> **Session-start order (multi-operator):** fetch → `git worktree add .coord coord` → read
> `.coord/ledger.md` tail + `.coord/claims/bryan.md` + `.coord/inbox/peter/` → THEN this file +
> `status.md`/`changelog.md` on `main`. Shared truth = `status.md` + `changelog.md`; live coordination =
> the coord branch.

## ▶ P3 close (2026-06-23) — doc-currency reconcile (no project work)

Short docs-only session. Caught up on Bryan's **B6/cleanup (PR #10 `ac2dd16`)** and reconciled the SoT.

- **Synced:** `main` ff `cccb641 → ac2dd16`; coord ff `04ee9c3 → 2876de7`. Both claims idle, inbox clean,
  no contention. Bryan's **B6 now closed cleanly** (claim reset + close block added) — the stale-claim nudge
  from P2 is resolved; nothing left to flag there.
- **Bryan's PR #10 closed two of my standing items:** `.gitattributes` durable LF fix (my P1 Windows CRLF
  find) + removal of the dead off-path `internal/parser`. **Neither is mine to carry anymore.**
- **Fixed the only stale note:** the Peter block in `status.md` still listed those two as "still open" →
  reconciled. Added P3 lines to `status.md` + `changelog.md` + this hand-off; bumped `last-reviewed`.
- **Off-repo (not a repo change):** Enter key was broken — `~/.claude/keybindings.json` had remapped submit
  to double-Enter (`enter`→null, `enter enter`→submit). Reset to defaults; Enter submits again. *Needs a
  Claude Code restart to take effect.*
- **Landing:** P3 docs committed to branch `peter/p3-doc-currency` (bare wrap — feature branch **NOT pushed**;
  PR to `main` pending operator auth). Coord close pushed direct.

## ▶ P2 close (2026-06-21) — serial-split tap (hardware design, BUILD PAUSED pending #1)

Arc: **`serial-split-tap`** — an isolated, listen-only serial tap so the Pi 4B can ingest a live DAQ
stream without disturbing the system that already consumes that serial. **Design locked + spec landed
on `main` (PR #7, `1b942eb`); build PAUSED pending operator measurement #1.** Full spec:
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
- Scope doc merged to `main` (PR #7); branch deleted. coord: P2 **closed**, claim reset to **idle**.
- **Resume trigger:** operator returns with #1 (and parts arrive) → new arc `peter/p2-serial-split-build`
  → solder + bench replay → real-wire on Pi → coexistence, per the scope doc's §"Build & test plan".

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

1. **`peter/p3-doc-currency` is committed but UNPUSHED** — bare wrap. Push + open a PR to `main` when
   authorized (it's docs-only, trivial). Or fold into the next arc's PR.
2. **`serial-split-tap` build (P2)** — **PAUSED** pending operator measurement **#1** (DAQ TXD idle
   voltage) + parts. Resume = new arc `peter/p2-serial-split-build` → solder → bench replay → real-wire on
   Pi → coexistence, per [`scope.md`](../changes/serial-split-tap/scope.md) §"Build & test plan". See the
   P2-close section below for the full hardware context.
3. **Totco preset** — low-priority; when a Totco unit is reachable (same direct-laptop capture method).

_Resolved (no longer open):_ `.gitattributes` durable CRLF fix + dead `internal/parser` removal (Bryan,
PR #10 `ac2dd16`); `pa.md` topology rewrite (Bryan, PR #6); Bryan's stale B6 coord claim (closed in B6);
P1 ruleset blocks (issue #3 closed, verified) — coord push-direct + branch deletion both work.

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

- `.coord` worktree on branch `coord`, **pushed + synced** (P3 close pushed direct; `origin/coord = local`).
- My `claims/peter.md` = **idle** (still idle through P3 — no project work claimed).
- **Do NOT remove the `.coord` worktree** — it's the live coordination channel (works both ways).

## State as of P3 close

| Item | State |
|---|---|
| `main` | `ac2dd16` (synced; Bryan's B6/cleanup PR #10 merged) |
| `peter/p3-doc-currency` | committed, **UNPUSHED** (bare wrap; PR pending auth) |
| Multi-party model | ADOPTED; full PR-flow cycles done; B6 closed cleanly |
| Windows toolchain | ✅ Go 1.26.4 + Node 24.17.0; full gate green; CRLF fixed |
| Commit gate | ✅ installed |
| Phase 4b / project MVP | ✅ DONE (Bryan PR #1), PA-verified E2E here |
| P1 follow-ups (.gitattributes, parser) | ✅ resolved (Bryan PR #10) |
| P2 serial-split-tap | design done; **build PAUSED** on operator measurement #1 |
| coord | ✅ pushed + synced; works both ways |
| Tests (P3 wrap) | `go vet ./...` ✅ · `go test ./...` ✅ · `gofmt -l` clean |
