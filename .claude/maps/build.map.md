# build.map.md
# project: cementer
# updated: 2026-06-12T09:02:13-06:00  commit: ee446c3

Toolchain note: Go is user-local at ~/.local/go/bin (the Makefile prepends it to PATH). Node/npm required for the web build.

## Makefile Targets  [Makefile]
make / make all  — alias for `build`
make build       — `web` then `server` (web client built first, then embedded Go binary)
make web         — `cd web && npm install && npm run build`  → web/dist
make server      — `CGO_ENABLED=0 go build -o cementer ./cmd/cementer`
make pi          — `web`, then `CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o cementer-arm64 ./cmd/cementer` (Raspberry Pi cross-compile, no C toolchain)
make run         — `build` then `./cementer -source testdata/sample-stream.txt` (synthetic replay, no pump)
make tidy        — `go mod tidy`
make clean       — rm -rf cementer cementer-arm64 web/dist data

## Web npm scripts  [web/package.json]
npm run dev      — `vite` (dev server with WS/debug proxy to :8080, hot reload)
npm run build    — `tsc && vite build` (typecheck then bundle to web/dist)
npm run preview  — `vite preview`

## Build pipeline
1. web/dist is produced by `npm run build` (tsc + vite).
2. `go:embed all:web/dist` (assets.go) bakes web/dist into the Go binary.
3. `go build` requires web/dist to exist; it is git-ignored and rebuilt each build.
=> The Go server build will FAIL if web/dist is absent — always run `make web` (or `make build`) first.

## Run commands
Dev (no pump):  ./cementer -source testdata/sample-stream.txt -replay-interval 250ms   (then open http://localhost:8080)
Prod (Pi):      ./cementer -serial /dev/serial/by-id/XXXX -baud 9600 -data-dir /mnt/ssd/cementer-data -addr :80

## CI/CD Pipeline
No CI/CD detected. No .github/workflows, .gitlab-ci.yml, or Jenkinsfile present.

## Commit gate
No pre-commit hook installed (core.hooksPath unset per docs/pa/status.md). Planned baseline (not yet active): gofmt -l + go vet ./... + go build ./... + go test ./...; make build pre-push when web/ changed.

## Docker
No Dockerfile or docker-compose present. Deployment is a single static binary + systemd unit (deploy/cementer.service), not a container.

## Tags
#cementer #map #build #makefile #vite #cross-compile #go-embed #systemd

## Links
- [primary.map.md](./primary.map.md)
- [dependencies.map.md](./dependencies.map.md)
- [test.map.md](./test.map.md)
- [config.map.md](./config.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
