# Changelog — cementer PA sessions

Cross-session audit trail (distinct from the git log — per-commit detail belongs in the log; this is
the human-discoverable session narrative). Newest block on top.

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
- **Tests:** not run this session (no source changed). **Git:** synced to `ddf8ada`; nothing committed
  (no authorization); `pa.md` + `pa-base.md` + `docs/pa/**` + `docs/{deep-dives,changes}/` untracked.
