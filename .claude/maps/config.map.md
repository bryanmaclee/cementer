# config.map.md
# project: cementer
# updated: 2026-06-19T23:05:55Z  commit: 1465bd9

## Environment Variables
| Key                | Required | Notes                                                     |
|--------------------|----------|-----------------------------------------------------------|
| CEMENTER_DATA_DIR  | optional | Directory for SQLite DB + raw logs. Fallback chain: -data-dir flag → $CEMENTER_DATA_DIR → ./data. [cmd/cementer/main.go] |

No `.env` / `.env.example` / `.env.template` file exists. `CEMENTER_DATA_DIR` is the only env var read by the Go binary.

## Command-line Flags  [cmd/cementer/main.go]
| Flag               | Default     | Notes                                                          |
|--------------------|-------------|----------------------------------------------------------------|
| -addr              | :8080       | HTTP listen address                                            |
| -serial            | ""          | serial device path, e.g. /dev/serial/by-id/... (one of -serial/-source required) |
| -baud              | 9600        | serial baud rate                                               |
| -source            | ""          | replay file of ASCII lines (dev source); overridden by -serial |
| -replay-interval   | 250ms       | delay between replayed lines                                   |
| -replay-loop       | true        | loop the replay file at EOF                                    |
| -data-dir          | ""          | directory for SQLite DB + raw logs                             |
| -batch-interval    | 250ms       | SQLite commit + live-broadcast cadence                         |
| -format            | intellisense| DAQ format preset: "intellisense" or "synthetic"               |

## Feature Flags
None. No feature-flag system in code.

## Config Files

### web/tsconfig.json
Target ES2022, module ESNext, moduleResolution "bundler", verbatimModuleSyntax, noEmit, isolatedModules, strict + noUnused* + noFallthroughCasesInSwitch. include: ["src"].

### web/vite.config.ts
build.outDir: "dist", emptyOutDir: true. Dev proxy: "/ws" → ws://localhost:8080 (ws:true), "/debug" → http://localhost:8080, "/api" → http://localhost:8080.

### deploy/cementer.service  (systemd unit template)
ExecStart flags (operator must edit): `-serial /dev/serial/by-id/CHANGEME`, `-baud 9600`, `-data-dir /mnt/ssd/cementer-data`, `-addr :80`. User/Group cementer, WorkingDirectory /opt/cementer, Restart=always RestartSec=2.

## SQLite runtime config (in DSN, not a file)  [internal/store/store.go]
`journal_mode=WAL`, `synchronous=NORMAL`, `busy_timeout=5000`, `foreign_keys=ON`, `MaxOpenConns=1`.

## localStorage keys (browser, per-laptop personal config)
| Key                | Owner                       | Notes                                        |
|--------------------|-----------------------------|----------------------------------------------|
| cementer.theme     | web/src/theme.ts            | "dark" (default) or "light"                  |
| cementer.liveview  | web/src/chart/config.ts     | JSON blob: {hidden, colors, windowSec}       |

## Secrets note
No application secrets, API keys, or credential keys are read by the Go binary. No `.env` file exists in the repo.

## Tags
#cementer #map #config #flags #env #systemd #vite #sqlite-pragmas #localstorage

## Links
- [primary.map.md](./primary.map.md)
- [build.map.md](./build.map.md)
- [infra.map.md](./infra.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
