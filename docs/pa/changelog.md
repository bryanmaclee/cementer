# Changelog — cementer PA sessions

Cross-session audit trail (distinct from the git log — per-commit detail belongs in the log; this is
the human-discoverable session narrative). Newest block on top.

---

## 2026-06-21 — Peter P1 · adopt multi-party model + stand up Windows toolchain + verify Phase 4b

First session by the **second co-equal operator (Peter)** on the Windows field laptop. Coord id **P1**.

- **Caught up + adopted (S6 machinery):** opened on the stale single-operator contract; found Bryan's
  S6 work and (mid-session) his **PR #1** merge → `main` `c952c54` (Phase 4b printable report + commit
  gate + multi-party DD + `coord` branch + meta-doc partition). On the operator's ruling **"adopt it"**:
  fast-forwarded to `c952c54`, installed the commit gate (`core.hooksPath=scripts/git-hooks`), onboarded
  to coord (P1 ledger + claim), and reconciled stray single-operator edits to the partition.
- **Windows toolchain stood up:** installed **Go 1.26.4** + **Node 24.17.0 / npm 11.13.0** (winget) +
  `~/.bashrc`/`~/.bash_profile`. Fixed a Windows-only **CRLF/gofmt break** (no `.gitattributes` +
  `autocrlf=true` → gate rejects every Go change): set `autocrlf=false`, renormalized the tree to LF.
  **Full gate validated green on Windows** (gofmt/vet/build/test + web tsc+vite). `make` is absent →
  Makefile recipes run directly.
- **Phase 4b PA-verified E2E:** built + ran the binary on the demo stream, created a job, recorded a
  segment, and rendered the **Report tab** + print-media output via headless **Edge** (`channel:'msedge'`,
  no browser download). Confirmed D-pdf (browser Save-as-PDF only) works.
- **Landed P1 onboarding docs** via **PR #2 → `main` `0a96095`** (first full PR-flow cycle by Peter:
  branch → push → gate → PR → merge). Confirmed `peter/*` feature-branch pushes are allowed.
- **Surfaced two ruleset problems to Bryan** (GitHub issue **#3** + a `coord` `inbox/bryan/` notice):
  the require-PR rule wrongly covers `coord` (should be push-direct) and restrict-deletions blocks
  merged-branch cleanup. Both are repo-owner config.
- **Blocked, pending Bryan:** `coord` pushes (`13c695a`, `b5d0089` stuck local); merged-branch deletion.
- **Tests:** `go test ./...` ✅ · `go vet ./...` ✅ · `gofmt -l` clean · `go build ./...` ✅ · web build ✅.

## 2026-06-18/19 — Session 5 · Phases 2, 3a, 3b, 4a shipped (DAQ engine → self-describing pump → jobs/recording → charting)

- **Machine:** Linux garage desktop. FULL profile read (pa.md + pa-base.md + data-model.md + README +
  status + hand-off + full user-voice + package-doc reads + git-sync). Start-state: S4's "push pending
  (GitHub Desktop)" had in fact landed — `origin/main = local = 2d28a3d`, verified by git STATE.
- **Six commits landed + pushed** (each: canonical `cementer-go-engineer`, worktree-isolated, model opus;
  PA independent E2E at landing; one PA-authored commit; worktree removed):
  - `83f036a` **Phase 2** — `internal/daqformat` generic config-driven engine + **Intellisense preset**
    built from the *live wire* (14-col, no header, `HH:MM:SS`-uptime → server-stamped), NOT the
    superseded 15-col CSV export. Verified: 13 channels, `agg.pressure == unit1.pressure` sum proven E2E.
  - `cd71beb` **Phase 3a** — self-describing pump: `pump_profiles`/`profile_channels` (seeded from the
    Phase-2 vocab), per-connection **hello/profile** WS frame, `GET/PUT /api/profile` + reset,
    scope-grouped vanilla-TS readout (enabled-only). Verified: WS greeting lists enabled-only after a PUT;
    seed idempotent across restart.
  - `cf46ab3` **Phase 3b** — `jobs` + `recording_segments`, `/api/jobs*` + `/api/recording/*`, minimal
    client controls. **Axiom #1 PROVEN**: samples climb while recording STOPPED, RECORDING, and after STOP
    — recording is a pure marker. Axiom #5 held (no stage reset).
  - `5c69e07` **Phase 4a** — charting core: `store.Series`/`JobSeries` (spike-preserving min/max
    decimation) + `GET /api/samples` + `GET /api/jobs/{id}/series` (read-only, single conn); **uPlot**
    live rolling chart (replaces the value grid; role-grouped axes; legend keeps latest values) + job
    history chart with segment shading; personal live-view config in localStorage.
  - `1f65c13` **Collaborator quickstart** — `make demo` (real capture, correct format) + fixed `make run`
    (`-format synthetic`, was silently dropping every line) + README "Quick demo" for Peter; Go 1.26+ /
    `docs/the plan` currency fixes.
  - `1465bd9` **Phase 4a fix-ups** — uPlot time axis fed **ms→seconds** (labels were off 1000×; line
    shapes were already right) + `testdata/intellisense-demo.txt` (ten real captures concatenated → varied
    multi-phase demo, no more sawtooth loop).
- **Decisions (locked, in the scope docs):** Phase-2 D1–D4; Phase-3 D1–D10 (D2 single-conn CRUD = store
  sole DB owner; D4 auth deferred; D8 job fields); Phase-4 X=time, all-enabled role-grouped, replace-
  readout, **PDF = browser Save-as-PDF only**. Adopted the **landing discipline**: fold each realized
  contract into `data-model.md` at the landing that ships it (applied every arc this session).
- **Headless verification unlocked:** Playwright browsers are cached locally; temp-installed
  `playwright@1.60.0` and **screenshotted the live chart** — confirmed the seconds-fix renders correct
  2026 timestamps + varied traces + no stacking. The "stacking" the user saw was the old single-capture
  demo loop, not a chart bug. (Saved to auto-memory.)
- **Scope artifacts written:** `docs/changes/phase3-jobs-recording-profiles/scope.md`,
  `docs/changes/phase4-charting-printing/scope.md`; six dispatch briefs archived under `docs/pa/briefs/`.
- **Wrap:** `go test ./...` ✅ (api/daqformat/parser/store; others no-test) · `go vet` ✅ · `gofmt -l` clean
  · `make build` ✅ (CGO-free, uPlot bundled offline). Nav-maps regenerated (were 5 phases stale).
- **Left:** Phase **4b** (print template + per-job overrides + print-CSS/PDF) — not started; minor
  `controls.ts` new-job-form-renders-expanded cosmetic; **3c** retention (deferred by design).

---

## 2026-06-16 — Session 4 · Intellisense DAQ live wire captured (D4 CLOSED for Intellisense)

- **Machine:** field LAPTOP (Windows, this checkout). FULL profile read (pa.md + pa-base.md +
  data-model.md + README + status + hand-off + full user-voice + git-sync). Skipped the deep `internal/*`
  package-doc reads — not load-bearing for a direct-laptop serial arc (no Go pipeline involved).
- **Start-state correction:** the Session-3 hand-off header said "PAUSED, uncommitted, nothing committed
  this session," but git showed commit `04ba031` had in fact committed + pushed Session-3's work (origin
  = local, 0/0). The only dirty file was an uncommitted Session-3 changelog block (accurate; folded in
  here). Verified against git STATE, not the hand-off narrative.
- **Goal (user):** capture real raw-data from the **Intellisense** DAQ (Totco deferred — not accessible).
  Different rig from the Session-3 Totco attempt: **Prolific PL2303GT adapter on COM7**.
- **Found the wire contract empirically:** first read @ 9600 8N1 → 43% printable (garbage); a printable-
  ratio baud sweep found **19200 8N1 → 100%**. Format = comma-delimited, **no header**, **14 fields**,
  `<CR><LF>`, ~1 line/s, timestamp = **`HH:MM:SS` uptime** (resets to `00:00:00` on power-up).
- **No header from the unit** — confirmed on both a record pause/unpause and a full power-cycle. Mapping
  is therefore empirical.
- **Mapped columns by actuating the rig** (per-action captures, comparing which column moved):
  - rate → cols 3 & 7; volume totals → cols 4 (job) & 12 (stage); both pairs track together (1 unit).
  - pressure → col 5 (unit 1, 0→1306 on slow valve close); **col 2 = aggregate = sum(col 5, col 6)**
    proven (col 6 flat, no 2nd unit) — confirms the user's hypothesis + the data-model aggregate concept.
  - **density → col 1 = 8.21, matching the unit's own interface readout** (ground truth).
  - water rate (col 9) never moved — this pump has no flow meter (user-confirmed).
  - 6 flat columns (6, 8, 9, 10, 11, 13) all explained by this 1-unit / no-backup-density / no-flow-meter
    rig; identities fixed by the column order, which matches the earlier-decoded Enbridge CSV.
- **Corrected an earlier call:** the live wire differs from the 15-col Enbridge **CSV** in *framing*
  (HH:MM:SS vs Excel-serial, no header, 14 vs 15 cols) but the **column order/semantics match** — the CSV
  is a valid identity guide; the preset to build is the live 14-col one.
- **Parser note for Phase 2:** power interruption produced a torn `?,,,,...` fragment → the parser must
  skip non-14-field lines (raw log keeps the bytes; structured store drops the bad line).
- **Banked:** new findings doc `docs/changes/phase2-intellisense-daqformat/intellisense-wire-capture-2026-06-16.md`
  (full map + Phase-2-ready DaqFormat preset); updated `status.md` (D4 closed for Intellisense, phase
  board, real-format section), `data-model.md` (preset now from real wire), `user-voice.md` (Session 4),
  hand-off. **10 raw `.bin` captures committed under `captures/`** (not gitignored → travel as provenance).
- **No source code changed** (docs + captures only). **D4 wire-contract gate CLOSED for Intellisense;**
  Phase 2 (`internal/daqformat` engine + Intellisense preset) is now fully unblocked. **Totco preset
  still TODO** (unit not accessible this trip). **Push pending** — user pushes via GitHub Desktop (GCM
  hang).

---

## 2026-06-14 — Session 3 · direct-laptop serial capture pivot + Totco diagnostic (BLOCKED, physical)

- **Machine:** field LAPTOP (Windows, this checkout). FULL profile read (pa.md + pa-base.md +
  data-model.md + README + status + hand-off + full user-voice + package docs + git-sync). Found a prior
  Session-3 start (2026-06-13) had rotated the hand-off (S2 snapshot → `archive/hand-off-2026-06-13.md`)
  but never committed it; verified the archive == committed S2 hand-off (pure CRLF/LF diff). Continued as
  Session 3 rather than re-rotating.
- **Goal (user):** capture the real DAQ raw-data feed. **Pivot:** plug the RS-232→USB adapter **directly
  into the laptop** and read the wire there — no Pi, no Go build, no Node (D4 = wire contract; a raw byte
  capture is the purest form). **Two DAQs to capture: Totco first, then Intellisense** → defines BOTH
  Phase-2 presets.
- **Verified the field runbook's serial claims against live code** before driving: `serialreader.
  DefaultConfig` = 9600 8N1; `buildSource` overrides only `BaudRate`; `bufio.ScanLines` handles `\n`/`\r\n`
  but not bare-`\r`. The `Config` struct already has DataBits/Parity/StopBits fields (unwired) → exposing
  flags is small, not a refactor.
- **Tooling:** added `tools/serial-read.ps1` (PowerShell `System.IO.Ports.SerialPort`; `-Sweep`,
  `-Loopback`, normal read → hex/ASCII dump + `captures/*.bin`).
- **Totco settings (from the DAQ config screen):** COM6 · 9600 8N1 (parity off) · Protocol 1 · 250 ms.
- **⛔ BLOCKED — total silence on COM6.** 0 bytes at 9600 8N1 (±DTR/RTS), and **0 bytes across a full baud
  sweep 2400→115200**. Wrong baud yields *garbage*, not silence → diagnosed as **physical/electrical, not
  settings**: straight-through cable between two DTE ends (needs null-modem), or DAQ not transmitting, or
  adapter fault. Modem lines CD/CTS/DSR all low throughout. Active loopback returned nothing (adapter not
  jumpered → inconclusive about the adapter itself).
- **Resume steps recorded** in hand-off (loopback self-test → null-modem cable → confirm Totco actually
  streaming → capture Totco then Intellisense). **D4 still OPEN.**
- **No source code changed** (docs + the PS tool only). **Committed + pushed** `04ba031` to `origin/main`
  (user authorized "commit and push it"). Push hung on Git Credential Manager (interactive auth); user
  completed it via the **GitHub Desktop app**; stuck `git-credential-manager`/`git-remote-https` procs
  killed. Recorded the GCM-hang as a persistent memory.

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
