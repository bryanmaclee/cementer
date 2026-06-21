# Hand-off — Peter (live)

`as of: Session P1 (in progress) · 2026-06-21 · operator: peter` (machine: **Windows field laptop** — `C:\Users\pjoli\Documents\GitHub\cementer`)

> Peter's per-operator session hand-off (multi-operator partition, S6). Optimize for Peter's next-session
> pickup. Peter's PA rewrites this at his session wraps; Bryan does not edit it (CODEOWNERS → Peter).
>
> **Shared truth** (phase board, code state, audit trail) lives in `status.md` / `changelog.md` on `main`.
> **Live cross-operator coordination** lives on the **coord branch** — set up the worktree with
> `git worktree add .coord coord` (the `make coord` target needs `make`, which is **absent on this box**),
> then read `.coord/ledger.md` + `.coord/claims/bryan.md` + `.coord/inbox/peter/` at session start.

## ✅ This Windows laptop is now a full dev box (P1)

- **Toolchain installed (winget):** Go **1.26.4**, Node **24.17.0**, npm **11.13.0**. Added to a new
  `~/.bashrc` + `~/.bash_profile`; the Windows **machine PATH** already includes them, so a
  freshly-launched shell (next session) auto-resolves the toolchain. **NOTE:** the Bash *tool* runs
  non-interactively and this session's shell predates the install — within a single session, prepend
  `export PATH="/c/Program Files/Go/bin:/c/Program Files/nodejs:$PATH"` if `go`/`npm` don't resolve.
- **`make` is NOT installed.** Run Makefile recipes directly: `make hooks` = `git config core.hooksPath
  scripts/git-hooks`; `make coord` = `git worktree add .coord coord`; `make web` = `cd web && npm install
  && npm run build`; `make build` = web build + `go build ./cmd/cementer`.
- **Commit gate installed:** `core.hooksPath=scripts/git-hooks`.
- **CRLF fix (Windows-only bug):** the repo has **no `.gitattributes`** and Git-for-Windows defaults to
  `core.autocrlf=true`, so the whole tree checked out CRLF and `gofmt` (LF-only) flagged every `.go` file
  → the pre-commit gate would reject any Go change here. Fixed this clone: **`core.autocrlf=false`** +
  renormalized the working tree to LF. **Durable cross-clone fix still TODO** (see OPEN #3).
- **Full gate validated green on Windows:** `gofmt` clean · `go vet ./...` · `go build ./...` (embed) ·
  `go test ./...` all pass · web build (tsc strict + vite) ✓.

## ⚠ OPEN — decisions / blockers (do these next)

1. **`coord` direct-push is blocked by a repo ruleset** ("Changes must be made through a pull request",
   hitting `refs/heads/coord`). This **contradicts the multi-party design** (coord = unprotected /
   push-direct for low latency). **Bryan's repo-config call:** either exempt `coord` from the ruleset
   (intended) or accept coord-via-PR. My P1 coord onboarding (`13c695a` on the `coord` branch) is
   stuck behind this. Raised in `.coord/ledger.md` (committed locally, unpushed).
2. **Get the P1 meta-docs to `main`.** `user-voice-peter.md` + this file are updated but **uncommitted**.
   Per PR-flow → a `peter/<arc>` branch → PR (CODEOWNERS routes `*-peter` to Peter, self-merge fine).
   Push auth + the gate both work now, so this is unblocked once we decide to do it. **Untested:** whether
   the repo ruleset also blocks pushing a `peter/*` feature branch (the coord rejection may be branch-wide
   or main+coord-only) — find out by pushing the first `peter/<arc>` branch.
3. **Durable CRLF fix → propose `.gitattributes`** (`* text=auto eol=lf`, maybe `*.go text eol=lf`) via a
   small PR so no future Windows clone hits the gofmt break. Touches Bryan's clone normalization too →
   coordinate. (This clone is already safe via `autocrlf=false`.)
4. **`pa.md` topology rewrite** — still declares "standalone single-operator", STALE since S6. The DD
   names the §4/§10 rewrite + the symmetric `hand-off-bryan.md`/`user-voice-bryan.md` rename as the
   pending follow-up. **Whose arc (Bryan or Peter)?** — raised in the coord ledger; coordinate first.

## ▶ Project work (none claimed by Peter yet)

- **Phase 4b is DONE** (Bryan, PR #1) — do NOT rebuild. Project MVP effectively reached. Not yet *run*
  end-to-end on this box (toolchain's ready now — `cd web && npm i && npm run build` then
  `go run ./cmd/cementer -source testdata/sample-stream.txt -format synthetic`, open :8080).
- **Parser cleanup** — `internal/parser` off the main path; delete or fold into a daqformat test.
- **Totco preset** — blocked (unit not accessible); direct-laptop capture method in the S5 archive hand-off.

## Per-clone git config set this session

`core.autocrlf=false` · `core.hooksPath=scripts/git-hooks` · remote = HTTPS (cred cached). `gh` not installed.

## State as of close (P1, in progress)

| Item | State |
|---|---|
| Local main | `c952c54` (synced; commit gate installed) |
| Multi-party model | ADOPTED |
| Windows toolchain | ✅ Go 1.26.4 + Node 24.17.0; full gate green; CRLF fixed (this clone) |
| Push auth | ✅ cached (no more hang) |
| Coord onboarding | committed locally `13c695a`, **push-pending** (repo ruleset, OPEN #1) |
| P1 meta-docs | updated, **uncommitted** (OPEN #2) |
| Phase 4b | DONE (Bryan, PR #1); not yet run on this box |
