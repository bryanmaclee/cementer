# Changelog — cementer PA sessions

Cross-session audit trail (distinct from the git log — per-commit detail belongs in the log; this is
the human-discoverable session narrative). Newest block on top.

---

## 2026-06-13 — Session 2 · bench-top hardware validation (Go+SQLite Pi stack)

- **Machine:** Peter's garage desktop (Windows). FULL profile read (pa.md + pa-base.md + data-model.md +
  README + status + hand-off + user-voice + git-sync); rotated S1 hand-off → `archive/hand-off-2026-06-12.md`.
- **Goal (user):** deploy Bryan's Go binary to the Pi 4B to bench-test the hardware before driving to the
  real DAQ unit.
- **Built the Pi binary on the desktop:** installed **Go 1.26.4** (`winget GoLang.Go`); Node 18.12.1 too
  old for Vite ^8 → **stubbed `web/dist/index.html`** (gitignored) to satisfy `//go:embed all:web/dist`
  (Rule-3 shortcut, surfaced — UI placeholder is fine for a hardware test). Cross-compiled
  `CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o cementer-arm64 ./cmd/cementer` → 14.8 MB static
  aarch64 ELF (verified via `file`). `scp`'d to `serial123@10.0.0.105`.
- **Verified against CODE, not README:** `-addr` default `:8080` (README said `:80`); raw log =
  `<data-dir>/raw-<ts>.log` (appended before parse); endpoints `/ws/live`, `/debug/stats`, `/`; Go 1.26.4
  required (README said 1.22+); `serialreader` hard-wired **8N1**, only baud is a flag.
- **✅ BENCH-TOP PASS (both serial paths)** on `CementSerial` (10.0.0.105), topology laptop `send_csv.py`
  → ESP32 (`csvToSerialSend`) → [path] → Pi `cementer`, **simulated transport** (recorded Enbridge CSV):
  - **GPIO UART** (`/dev/serial0`→`ttyS0` @115200): raw log filled (15-col lines), SQLite WAL 2.24 MB,
    `/debug/stats` → **2,812 rows**, HTTP 200 from the laptop.
  - **CP2102 USB adapter** (`/dev/ttyUSB0`, by-id @115200): fresh `~/cementer-usbtest` db → **4,404 rows**.
    This is the exact Pi-side path the real RS-232→USB adapter will use.
- **Debug recovered:** "Pi not capturing" was a false alarm — `cementer` is silent on stdout, capture
  shows in the raw log / `/debug/stats` (logged under hand-off anomalies).
- **Scope boundaries stated honestly (user-aligned):** transport/plumbing/columns = REAL; **wire contract
  + channel semantics still UNPROVEN** → real DAQ (Phase 2 D4 / no-code mapping). User held logging until
  results were verified ("don't convolute the logs with maybes") — honored; logged only after both passes.
- **Authored the ⚡ FIELD RUNBOOK** in `hand-off.md` (cold-start DAQ procedure + gotchas: get DAQ serial
  settings before driving, 8N1-only limitation, by-id paths, silent-stdout) so the field LAPTOP (different
  machine) can execute via `git pull`.
- **Hygiene:** added `/cementer-arm64` to `.gitignore` (build artifact was untracked-but-not-ignored).
- **No source code changed** (docs + a cross-compile only). **Committed + pushed** to `origin/main` (user
  authorized: "go ahead and push it").

---

## 2026-06-12 — Session 1 · PA workflow init

- Read `pa-base.md` (`pa-base v1`) in full + surveyed the real project state.
- **Decisions** (AskUserQuestion): scope = full contract + scaffolding; topology = standalone island.
- **Authored** `pa.md` — the cementer PA overlay v1: a pointer to the vendored base + fills for all
  ~32 base slots (Go 1.26 / Vite-vanilla-TS / modernc-SQLite / Raspberry-Pi target) + 5 project axioms
  (raw≠live≠recording independence; no-code DAQ-format adaptation; standalone self-describing Pi;
  layered durability; recording-segments-as-markers).
- **Created scaffolding:** `docs/pa/{status,hand-off,user-voice,design-insights,changelog,anti-patterns}.md`
  + `docs/pa/archive/`, `docs/pa/briefs/`, `docs/deep-dives/`, `docs/changes/`.
- **Verified state (not narrative):** store schema = `samples` only (recording/jobs/profiles unbuilt);
  `assets.go` `//go:embed all:web/dist` (fresh-worktree build gotcha); no git commit gate installed;
  only `internal/parser/parser_test.go` exists.
- **Recorded** INS-001 (verify code state, not commit-message narrative).
- **Flagged debts:** stale `docs/plan` reference in `main.go`/README; nav-maps ungenerated; no commit
  gate.
- **git-sync (session-start):** fetched; local `main` was 1 behind; clean fast-forward to `ddf8ada`
  (collaborator Peter Oliver). That commit added: 3 real Enbridge DAQ CSVs (~25k rows, 15-column
  format), an ESP32 test rig (`csvToSerialSend.ino`, `send_csv.py`: laptop CSV→USB→ESP32→UART2→Pi),
  and a `pi4b & test db/...README`.
- **Findings surfaced (not resolved):**
  (1) ⚠ **MAJOR FORK** — the README documents a working **Python→InfluxDB→Grafana** prototype vs this
  repo's **Go→SQLite→custom-vanilla-TS** stack; DB/serve choice left open. Logged to status.md as an
  axiom-level deliberation point.
  (2) Phase 2 **unblocked** — real DAQ format decoded into status.md.
  (3) ⚠ Plaintext credentials committed in the README — hygiene flag logged.
  (4) Parser-vs-real-format handling unverified.
- **Committed + pushed** the workflow init (`e290b8d`).
- **Deep-dive ran (R2)** on the storage+viz fork: 3 parallel sourced research agents (InfluxDB-on-Pi,
  Grafana-printable-charts, SQLite-TS + single-binary-ops) → `docs/deep-dives/storage-and-viz-architecture-2026-06-12.md`.
  Recommendation: **adopt (A) Go+SQLite+uPlot, retire (B) Influx/Grafana to dev bench** — pending user
  ratification. Recorded INS-002.
- **nav-maps:** full cold-start → `.claude/maps/` (13 maps + non-compliance report), pushed (`b0fef5f`).
  Surfaced: `parser.DefaultConfig` is synthetic 4-channel (≠ real 15-col); README "Go 1.22+" vs 1.26.4;
  empty placeholder dirs `internal/api/`, `web/src/chart/`.
- **✅ Fork RATIFIED (user):** adopt (A) Go+SQLite+uPlot; retire (B) Influx/Grafana to dev bench. DAQ
  format = **Intellisense**. Retention/downsampling → Phase 3/4. Marked the deep-dive RATIFIED; updated
  status.md (fork resolved, Phase 2 = Intellisense), hand-off, user-voice.
- **Tests:** not run this session (no source changed — docs/maps only).
- **Phase 2 scoped** → `docs/changes/phase2-intellisense-daqformat/scope.md` (curated from real CSVs;
  model/store already fit; 8-step work breakdown; E2E-verify strategy). Pushed `e395623`.
- **Phase 2 decisions locked (user):** D1 new `internal/daqformat` pkg · D2 embedded LOGTIME (Excel-serial)
  + server fallback · D3 map `meta.*` channels now, semantics → Phase 3 · **D4: get a live-serial capture
  before "done"** (build GATED) → drafted `live-serial-capture-request.md` for the collaborator.
- **Forged the canonical dev agent** `~/.claude/agents/cementer-go-engineer.md` (effective NEXT session;
  Go/SQLite/Pi/vanilla-TS, axiom-aware, modeled on scrml-js-codegen-engineer). Per §5, dispatch Phase 2
  through it next session.
- **Git:** `e290b8d` (init) → `ee446c3` (deep-dive) → `b0fef5f` (maps) → `92113fc` (ratify) → `e395623`
  (scope) → `f44b41a` (Phase-2 decisions/agent). All on `origin/main`.
- **WRAP (wrap and push):** tests `go build ./...` ✅ / `go vet ./...` ✅ / `go test ./...` ✅ (parser
  passes; other pkgs no test files; `web/dist` present so embed compiled). Worktrees: only main (none to
  clean). Maps: current (no source changed since stamp `b0fef5f` — refresh no-op). Inbox/outbox: N/A
  (standalone). State-doc regen: N/A. Tree clean; pushed; `origin/main` synced 0/0. Hand-off finalized
  (next session rotates it to archive). **Session 1 closed.**
