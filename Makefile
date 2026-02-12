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

# Quick dev cycle: build + install to PATH + launch TUI
start: build
	@cp $(BIN) /usr/local/bin/$(APP) 2>/dev/null || { mkdir -p $(HOME)/bin && cp $(BIN) $(HOME)/bin/$(APP) && echo "Installed $(APP) $(VERSION) ($(COMMIT)) to $(HOME)/bin/$(APP)"; } || true
	@command -v $(APP) >/dev/null 2>&1 && $(APP) || ./$(BIN)

# Same as start but skip launching (just build + install)
dev: build
	@cp $(BIN) /usr/local/bin/$(APP) 2>/dev/null || { mkdir -p $(HOME)/bin && cp $(BIN) $(HOME)/bin/$(APP); } || true
	@echo "Installed $(APP) $(VERSION) ($(COMMIT))"

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
	@cp $(BIN) /usr/local/bin/$(APP) 2>/dev/null || { mkdir -p $(HOME)/bin && cp $(BIN) $(HOME)/bin/$(APP); }
	@echo "Installed $(APP) $(VERSION)"

snapshot:
	goreleaser release --snapshot --clean

release:
	goreleaser release --clean
