APP     = pskill
BIN     = bin/$(APP)
VERSION = $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build run tidy test lint fmt clean install release snapshot

build:
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(BIN) ./cmd/pskill

run: build
	./$(BIN)

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
	install -m 755 $(BIN) /usr/local/bin/$(APP)

snapshot:
	goreleaser release --snapshot --clean

release:
	goreleaser release --clean
