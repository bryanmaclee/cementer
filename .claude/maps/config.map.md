# config.map.md
# project: cementer
# updated: 2026-06-12T09:02:13-06:00  commit: ee446c3

## Environment Variables
CEMENTER_DATA_DIR — optional — directory for SQLite DB + raw logs. Fallback chain: -data-dir flag → $CEMENTER_DATA_DIR → ./data  [cmd/cementer/main.go:65]

No .env / .env.example / .env.template file exists. CEMENTER_DATA_DIR is the only env var read in source (grep `os.Getenv`).

## Command-line Flags  [cmd/cementer/main.go:45]
-addr            — HTTP listen address (default ":8080")
-serial          — serial device path, e.g. /dev/serial/by-id/... (one of -serial/-source required)
-baud            — serial baud rate (default 9600)
-source          — replay file of ASCII lines (dev source); overridden by -serial
-replay-interval — delay between replayed lines (default 250ms)
-replay-loop     — loop the replay file at EOF (default true)
-data-dir        — directory for SQLite DB + raw logs (default $CEMENTER_DATA_DIR or ./data)
-batch-interval  — SQLite commit / live-broadcast cadence (default 250ms)

## Feature Flags
None. No feature-flag system in code.

## Config Files
### web/tsconfig.json
Strict TS: target ES2022, module ESNext, moduleResolution "bundler", verbatimModuleSyntax, noEmit, isolatedModules, strict + noUnused* + noFallthroughCasesInSwitch. include: ["src"].

### web/vite.config.ts
build.outDir: "dist", emptyOutDir: true. Dev proxy: "/ws" → ws://localhost:8080 (ws:true), "/debug" → http://localhost:8080.

### deploy/cementer.service  (systemd unit template)
ExecStart flags (operator must edit): -serial /dev/serial/by-id/CHANGEME, -baud 9600, -data-dir /mnt/ssd/cementer-data, -addr :80. User/Group cementer, WorkingDirectory /opt/cementer, Restart=always RestartSec=2.

## SQLite runtime config (in DSN, not a file)  [internal/store/store.go:38]
journal_mode=WAL, synchronous=NORMAL, busy_timeout=5000, foreign_keys=ON, MaxOpenConns=1.

## Secrets note
No application secrets, API keys, or credential keys are read by the Go binary. (Plaintext test-rig credentials exist in `pi4b & test db/credetials&currentDB.README`, but they belong to the collaborator bench, not the product — flagged in non-compliance.report.md.)

## Tags
#cementer #map #config #flags #env #systemd #vite #sqlite-pragmas

## Links
- [primary.map.md](./primary.map.md)
- [build.map.md](./build.map.md)
- [infra.map.md](./infra.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
