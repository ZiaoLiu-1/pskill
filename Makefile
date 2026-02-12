APP     = pskill
BIN     = bin/$(APP)
VERSION = $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  = $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    = $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS = -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: build run tidy test lint fmt clean install release snapshot start dev

build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/pskill

run: build
	./$(BIN)

# Quick dev cycle: clean old binary + build fresh + install + launch
start: clean build
	@if [ "$$(id -u)" -eq 0 ]; then echo "Do not run 'sudo make start'. Run 'make start' (it will ask sudo only for install)."; exit 1; fi
	@sudo rm -f /usr/local/bin/$(APP)
	@sudo cp $(BIN) /usr/local/bin/$(APP)
	@echo "$(APP) $(VERSION) ($(COMMIT)) → /usr/local/bin/$(APP)"
	@/usr/local/bin/$(APP) version
	@/usr/local/bin/$(APP)

# Same as start but skip launching (just build + install)
dev: build
	@if [ "$$(id -u)" -eq 0 ]; then echo "Do not run 'sudo make dev'. Run 'make dev'."; exit 1; fi
	@echo "Installing $(APP) $(VERSION) ($(COMMIT))..."
	@sudo cp $(BIN) /usr/local/bin/$(APP) && echo "  → /usr/local/bin/$(APP)" || echo "  ⚠ Run: sudo cp $(BIN) /usr/local/bin/$(APP)"

tidy:
	go mod tidy

test:
	go test ./... -v

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Install golangci-lint: https://golangci-lint.run/usage/install/"; exit 1; }
	golangci-lint run ./...

fmt:
	gofmt -w -s .
	go vet ./...

clean:
	rm -rf bin/ dist/ coverage/

install: build
	@if [ "$$(id -u)" -eq 0 ]; then echo "Do not run 'sudo make install'. Run 'make install'."; exit 1; fi
	@sudo cp $(BIN) /usr/local/bin/$(APP)
	@echo "Installed $(APP) $(VERSION) to /usr/local/bin/$(APP)"

snapshot:
	goreleaser release --snapshot --clean

release:
	goreleaser release --clean
