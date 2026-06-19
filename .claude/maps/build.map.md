# build.map.md
# project: cementer
# updated: 2026-06-19T23:05:55Z  commit: 1465bd9

Toolchain note: Go is user-local at `~/.local/go/bin` (the Makefile prepends it to PATH via `export PATH := $(HOME)/.local/go/bin:$(PATH)`). Node/npm required for the web build.

## Makefile Targets  [Makefile]
| Target       | What it does                                                                      |
|--------------|-----------------------------------------------------------------------------------|
| make / all   | alias for `build`                                                                 |
| make build   | `web` then `server` (web client built first, then embedded in Go binary)          |
| make web     | `cd web && npm install && npm run build` → web/dist                               |
| make server  | `CGO_ENABLED=0 go build -o cementer ./cmd/cementer`                               |
| make pi      | `web` then `CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o cementer-arm64 ./cmd/cementer` |
| make run     | `build` then `./cementer -source testdata/sample-stream.txt -format synthetic`    |
| make demo    | `build` then `./cementer -source testdata/intellisense-demo.txt -format intellisense -replay-interval 200ms` |
| make tidy    | `go mod tidy`                                                                     |
| make clean   | `rm -rf cementer cementer-arm64 web/dist data`                                    |

## Web npm scripts  [web/package.json]
| Command        | What it does                                                |
|----------------|-------------------------------------------------------------|
| npm run dev    | vite dev server (HMR, WS + API + debug proxy to :8080)     |
| npm run build  | `tsc && vite build` (typecheck then bundle to web/dist)     |
| npm run preview| `vite preview`                                              |

## Build pipeline
1. `npm run build` → `web/dist/` (tsc typecheck + vite bundle with uPlot included).
2. `//go:embed all:web/dist` in `assets.go` bakes `web/dist` into the Go binary.
3. `go build` **requires** `web/dist` to exist — always run `make web` (or `make build`) first.
4. CGO is never used; the pure-Go SQLite driver makes a fully static binary possible for arm64.

## Development run commands
```
# Synthetic replay (4-channel, no pump):
./cementer -source testdata/sample-stream.txt -format synthetic

# Intellisense demo replay (14-column multi-phase live capture):
./cementer -source testdata/intellisense-demo.txt -format intellisense -replay-interval 200ms

# Production (Pi, real pump):
./cementer -serial /dev/serial/by-id/XXXX -baud 9600 -data-dir /mnt/ssd/cementer-data -addr :80
```

## CI/CD Pipeline
No CI/CD configured. No `.github/workflows`, `.gitlab-ci.yml`, or `Jenkinsfile` present.

## Docker
No Dockerfile or docker-compose. Deployment is a single static binary + systemd unit (deploy/cementer.service), not a container.

## Tags
#cementer #map #build #makefile #vite #cross-compile #go-embed #systemd #uplot

## Links
- [primary.map.md](./primary.map.md)
- [dependencies.map.md](./dependencies.map.md)
- [test.map.md](./test.map.md)
- [config.map.md](./config.map.md)
- [master-list.md](../../master-list.md)
- [pa.md](../../pa.md)
