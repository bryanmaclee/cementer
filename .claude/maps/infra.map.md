# infra.map.md
# project: cementer
# updated: 2026-06-12T09:02:13-06:00  commit: ee446c3

## Deployment
Target: a single Raspberry Pi 4B (arm64) per pump — an offline "island" (no central server, no cloud). The cementer binary is the entire compute.
Artifact: one static, CGO-free binary (`make pi` → cementer-arm64) with the web client embedded via go:embed. No container, no orchestrator.
Process manager: systemd (deploy/cementer.service).

## systemd unit  [deploy/cementer.service]
Unit: cementer — "cement-pump data acquisition", After=multi-user.target.
Service: Type=simple, User/Group=cementer, WorkingDirectory=/opt/cementer.
ExecStart (operator-edited): /opt/cementer/cementer -serial /dev/serial/by-id/CHANGEME -baud 9600 -data-dir /mnt/ssd/cementer-data -addr :80
Restart=always, RestartSec=2. WantedBy=multi-user.target.
Install: copy unit to /etc/systemd/system/, daemon-reload, enable --now.

## Storage layout (runtime, on the Pi)
Data dir resolution: -data-dir → $CEMENTER_DATA_DIR → ./data.
  cementer.db (+ -wal, -shm)        SQLite WAL store
  raw-YYYYMMDD-HHMMSS.log           append-only raw capture, one per process start
Field guidance (in unit comments + README): put -data-dir on an external SSD, not the SD card (SD wear is the long-running failure mode); use a stable /dev/serial/by-id/... path.

## Cloud Resources
None. Fully offline; no cloud provider, no managed services.

## CI/CD
None configured (no .github/workflows, no other CI). Update model: build cementer-arm64, copy to Pi, restart the systemd service. Rollback = swap the binary file.

## Network
HTTP/WebSocket served on -addr (prod :80, dev :8080) over the local LAN. WS CheckOrigin accepts any origin (LAN posture).

## Collaborator bench (NOT product infra)
esp32sketches/ + "pi4b & test db/" describe a separate Python→ESP32→InfluxDB 2.9.1→Grafana 13.0.2 test rig used to prove the hardware data flow. The deep-dive (docs/deep-dives/storage-and-viz-architecture-2026-06-12.md) recommends retiring that stack to a dev/diagnostic bench. It is not part of the shipped product's infrastructure.

## Tags
#cementer #map #infra #raspberry-pi #systemd #arm64 #offline #single-binary

## Links
- [primary.map.md](./primary.map.md)
- [build.map.md](./build.map.md)
- [config.map.md](./config.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
