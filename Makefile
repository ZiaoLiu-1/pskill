APP=pskill

.PHONY: build run tidy

build:
	go build -o bin/$(APP) ./cmd/pskill

run:
	go run ./cmd/pskill

tidy:
	go mod tidy
