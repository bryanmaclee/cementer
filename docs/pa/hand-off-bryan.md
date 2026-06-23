# Hand-off — Bryan (live)

`as of: B6 close · 2026-06-22` (machine: Linux garage desktop — `/home/bryan-maclee/cementer`)

> Optimize for the NEXT Bryan session's pickup. **cementer is now a TWO-operator repo** (Bryan + Peter
> Oliver, co-equal). **On session OPEN, do the coord handshake FIRST** (`pa.md` v2 §4/§10): `git fetch`;
> read `.coord/ledger.md` tail + `.coord/claims/peter.md` + your `.coord/inbox/bryan/` (`make coord` if
> `.coord` isn't set up). Prior Bryan hand-offs: S5 → `archive/hand-off-2026-06-19.md` (when rotated).

## What B6 did (the biggest session)

1. **Phase 4b — printable per-job report (project MVP).** `internal/printcfg` + `/api/jobs/{id}/print-config`
   + Report tab (`web/src/report.ts`, `window.print()`/`@media print`). Merged **PR #1**.
2. **The entire multi-operator PA workflow**, when the operator brought Peter on: the DD (PR-flow ruled, no
   debate — no CI ⇒ peer review is the gate); the shared commit gate (`scripts/git-hooks`, `make hooks`);
   the **`coord` orphan branch** (ledger/claims/inbox); the meta-doc partition + **`pa.md` overlay v2**.
   Merged **PR #6**.
3. **Shared cleanup** (`.gitattributes` LF fix + removed dead `internal/parser`) — on **`bryan/cleanup`**,
   **PR open, NOT yet merged** (this is the one main-bound thing still pending).

## State as of close

| Item | State |
|---|---|
| `origin/main` | `cccb641` — `pa.md` **v2** live; 4b + multi-party foundation merged (PRs #1, #6); Peter's P1+P2 docs merged |
| `bryan/cleanup` PR | **open, awaiting merge** (`.gitattributes` + parser removal + ref updates + this wrap) |
| `coord` branch | live + synced (`make coord` → `.coord/`); my claim reset to **idle** at this wrap; Peter's notice acked |
| Remote branches | `main`, `coord`, `bryan/cleanup` only (merged/workaround branches deleted) |
| Source | `gofmt`/`vet`/`go test ./...`/`go build ./...` all green |
| Ruleset | fixed — scoped to **default-branch-only**; slash branches (`bryan/**`) + `coord` push fine |

## Open threads / next priorities

1. **Merge the `bryan/cleanup` PR** (it carries this wrap too). Then `git pull` main.
2. **nav-maps are STALE** — stamp `1465bd9` (S5); 4b + the entire multi-party system + parser removal all
   post-date them. **Regenerate** (`/map`) next session before any dev dispatch.
3. **pre-commit gate gap:** `--diff-filter=ACM` skips pure *deletions* (a delete-only commit takes the
   "no Go changes" fast path) — broaden to include `D`. Small `scripts/git-hooks/pre-commit` fix.
4. **Peter's `serial-split-tap` build** — paused on HIS field measurement #1 (DAQ TXD idle voltage) + parts.
   Not Bryan's action; track via coord. (Design + scope are landed on `main`.)
5. **Totco preset** — still TODO (unit not accessible). Same direct-laptop capture method (runbook below).
6. **Plaintext test-rig creds** in `pi4b & test db/...README` — rotate if repo ever goes public (unchanged).

## Multi-operator working notes (read before coordinating)

- **Peter is on a Windows field laptop.** `make` is NOT on PATH there — he runs Makefile recipes directly.
  The `.gitattributes` (this PR) fixes the CRLF/gofmt break he hit; he set `autocrlf=false` on his clone.
- **PR-flow:** branch `bryan/<arc>` → push → PR → merge to protected `main`. NEVER push `main` directly.
  `coord` + feature branches push direct. `gh` is NOT installed here → open PRs via the printed URL.
- **Coordination is optimistic** (claims advisory, not locks). Route any ask about Peter's in-flight arc
  into `inbox/peter/`, don't drive his files (CODEOWNERS owner-only on `*-peter`).

## Totco (deferred — unit not accessible since S3/S4)

COM6 · 9600 8N1 · Protocol 1 · 250 ms. S3 hit total silence on COM6 across all bauds → physical/electrical.
Resume: `-Loopback` self-test (jumper DB9 2↔3) → confirm null-modem cable → confirm the Totco is streaming →
capture as for Intellisense → define the Totco preset (a new `DaqFormat` value — no engine change).

## How to run / verify (reuse)

`make demo` → replays `testdata/intellisense-demo.txt` (real multi-phase capture) → `http://localhost:8080`.
`make run` = synthetic. cementer is **silent on stdout when healthy** — watch the browser / raw log /
`/debug/stats`. Headless verify: temp-install `playwright@1.60.0`, `executablePath` to the cached
`~/.cache/ms-playwright/chromium-<N>/chrome-linux64/chrome` (Linux) — see auto-memory.
