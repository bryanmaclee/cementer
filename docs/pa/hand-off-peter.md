# Hand-off — Peter (live)

`as of: P1 close · 2026-06-21 · operator: peter` (machine: **Windows field laptop** — `C:\Users\pjoli\Documents\GitHub\cementer`)

> Peter's per-operator hand-off (multi-operator partition, S6). Optimize for Peter's next-session pickup.
> Peter's PA rewrites this at his wraps; Bryan does not edit it (CODEOWNERS → Peter).
>
> **Session-start order (multi-operator):** fetch → `git worktree add .coord coord` → read
> `.coord/ledger.md` tail + `.coord/claims/bryan.md` + `.coord/inbox/peter/` → THEN this file +
> `status.md`/`changelog.md` on `main`. Shared truth = `status.md` + `changelog.md`; live coordination =
> the coord branch.

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

## ⚠ OPEN — next session (most are Bryan's to unblock)

1. **Push the blocked coord commits** once Bryan exempts `coord` from the require-PR rule:
   `13c695a` (P1 ledger open + claim) and `b5d0089` (the inbox/bryan ruleset notice). They sit on the
   local `coord` branch in the `.coord` worktree. `cd .coord && git push origin coord`.
2. **Land the P1 WRAP docs.** `status.md`, `changelog.md`, `user-voice-peter.md`, and this file were
   updated at wrap and committed on branch **`peter/p1-wrap`** — **PUSH + PR + MERGE still pending**
   (the wrap was bare "wrap P1", not "wrap and push"). `git push -u origin peter/p1-wrap` → open PR →
   merge (CODEOWNERS routes `*-peter` to Peter; the shared `status.md`/`changelog.md` may need Bryan
   review). _Until merged, `main`'s status/changelog lag this file._
3. **`.gitattributes` durable CRLF fix** — `* text=auto eol=lf` (+ maybe `*.go text eol=lf`) so no future
   Windows clone hits the gofmt break. A `peter/<arc>` PR; touches Bryan's clone → coordinate.
4. **`pa.md` topology rewrite** — still says "standalone single-operator" (STALE since S6). DD names the
   §4/§10 rewrite + the symmetric `hand-off-bryan.md`/`user-voice-bryan.md` rename. **Whose arc?** Raise
   on the coord ledger before claiming.
5. **Bryan is active** — his `bryan/s6-phase4b-multiparty` advanced to `5a2a5d5` after PR #1. Fetch +
   read the coord ledger before starting anything to avoid overlap.

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

- `.coord` worktree on branch `coord`, local tip `b5d0089` (2 commits ahead of `origin/coord`, **push-blocked**).
- My `claims/peter.md` = **idle** (reset at P1 close). My P1 ledger has open + close blocks.
- **Do NOT remove the `.coord` worktree** — it holds unpushed coordination commits.

## State as of P1 close

| Item | State |
|---|---|
| `main` | `0a96095` (P1 docs merged; synced) |
| Multi-party model | ADOPTED |
| Windows toolchain | ✅ Go 1.26.4 + Node 24.17.0; full gate green; CRLF fixed |
| Commit gate | ✅ installed |
| Phase 4b | ✅ DONE (Bryan PR #1), PA-verified E2E here |
| P1 wrap docs | committed `peter/p1-wrap`, **push/PR/merge pending** (OPEN #2) |
| coord onboarding (`13c695a`,`b5d0089`) | committed local, **push-blocked** (OPEN #1) |
| Bryan notified of ruleset | ✅ issue #3 + coord notice |
| Tests (P1 wrap) | `go test`/`vet`/`gofmt`/`go build`/web build all ✅ |
