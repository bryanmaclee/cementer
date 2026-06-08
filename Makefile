# cementer build. The web client is built first (vite -> web/dist) and then
# embedded into a single Go binary.
#
# Go is installed user-local in this environment; prepend it to PATH so `make`
# finds it. Harmless if go is already on PATH.
export PATH := $(HOME)/.local/go/bin:$(PATH)

GO ?= go
BIN ?= cementer

.PHONY: all build web server run tidy clean

all: build

# Build the web client (installs deps on first run) then the server binary.
build: web server

web:
	cd web && npm install && npm run build

server:
	CGO_ENABLED=0 $(GO) build -o $(BIN) ./cmd/cementer

# Cross-compile for a 64-bit Raspberry Pi (pure-Go SQLite => no C toolchain).
pi: web
	cd web >/dev/null
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build -o $(BIN)-arm64 ./cmd/cementer

# Run against the synthetic replay stream (no pump needed).
run: build
	./$(BIN) -source testdata/sample-stream.txt

tidy:
	$(GO) mod tidy

clean:
	rm -rf $(BIN) $(BIN)-arm64 web/dist data
