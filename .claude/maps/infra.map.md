# infra.map.md
# project: cementer
# updated: 2026-06-19T23:05:55Z  commit: 1465bd9

## Deployment
Target: a single Raspberry Pi 4B (arm64) per pump — fully offline "island" (no central server, no cloud). The cementer binary is the entire compute.
Artifact: one static, CGO-free binary (`make pi` → `cementer-arm64`) with the web client embedded via `go:embed`. No container, no orchestrator.
Process manager: systemd (`deploy/cementer.service`).

## systemd unit  [deploy/cementer.service]
Unit: cementer — "cement-pump data acquisition", After=multi-user.target.
Service: Type=simple, User/Group=cementer, WorkingDirectory=/opt/cementer.
ExecStart (operator-edited):
```
/opt/cementer/cementer \
    -serial /dev/serial/by-id/CHANGEME \
    -baud 9600 \
    -data-dir /mnt/ssd/cementer-data \
    -addr :80
```
Restart=always, RestartSec=2. WantedBy=multi-user.target.
Install: `sudo cp deploy/cementer.service /etc/systemd/system/ && sudo systemctl daemon-reload && sudo systemctl enable --now cementer`.

## Storage layout (runtime, on the Pi)
Data dir resolution: `-data-dir` → `$CEMENTER_DATA_DIR` → `./data`.
```
<data-dir>/
  cementer.db          SQLite WAL store (5 tables: samples, pump_profiles, profile_channels, jobs, recording_segments)
  cementer.db-wal      WAL write-ahead log (auto-managed)
  cementer.db-shm      shared memory file (auto-managed)
  raw-YYYYMMDD-HHMMSS.log   append-only raw capture, one file per process start
```
Field guidance: put `-data-dir` on an external SSD, NOT the SD card (SD wear is the long-running failure mode); use a stable `/dev/serial/by-id/...` path.

## Cloud Resources
None. Fully offline; no cloud provider, no managed services.

## Network
HTTP + WebSocket served on `-addr` (prod :80, dev :8080) over the local LAN.
WS `CheckOrigin` accepts any origin (LAN deployment; no CORS enforcement).

## CI/CD
None configured. Update model: `make pi` → copy binary to Pi → `sudo systemctl restart cementer`. Rollback = swap the binary file.

## Tags
#cementer #map #infra #raspberry-pi #systemd #arm64 #offline #single-binary #sqlite #ssd

## Links
- [primary.map.md](./primary.map.md)
- [build.map.md](./build.map.md)
- [config.map.md](./config.map.md)
- [schema.map.md](./schema.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
