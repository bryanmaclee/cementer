# cementer build. The web client is built first (vite -> web/dist) and then
# embedded into a single Go binary.
#
# Go is installed user-local in this environment; prepend it to PATH so `make`
# finds it. Harmless if go is already on PATH.
export PATH := $(HOME)/.local/go/bin:$(PATH)

GO ?= go
BIN ?= cementer

.PHONY: all build web server run demo tidy clean hooks

all: build

# Install the source-controlled git hooks (run ONCE per clone). Points git at the tracked
# scripts/git-hooks so every operator runs the identical pre-commit (gofmt+vet+build+test)
# and pre-push (make build when web/ changed) gate. See docs/pa/ for the multi-operator flow.
hooks:
	@chmod +x scripts/git-hooks/*
	@git config core.hooksPath scripts/git-hooks
	@echo "git hooks installed (core.hooksPath=scripts/git-hooks)"

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

# Run against the synthetic replay stream (no pump needed). -format synthetic matches
# the 4-channel testdata; the DEFAULT -format is intellisense (a 14-col shape), so this
# flag is required or every synthetic line is dropped by the field-count guard.
run: build
	./$(BIN) -source testdata/sample-stream.txt -format synthetic

# Demo the real Intellisense wire with NO pump: replays a multi-phase capture (the ten
# 19200-8N1 field captures concatenated chronologically) into a fully populated chart, so
# the loop shows real variety (idle -> rate -> density -> pressure to 1306 -> density),
# not one repeating ramp. After it starts, open http://localhost:8080.
demo: build
	./$(BIN) -source testdata/intellisense-demo.txt -format intellisense -replay-interval 200ms

tidy:
	$(GO) mod tidy

clean:
	rm -rf $(BIN) $(BIN)-arm64 web/dist data
