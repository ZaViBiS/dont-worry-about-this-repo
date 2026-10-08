.PHONY: test build lint fmt test-all

test:
	go test -v -race ./...

build:
	go build ./cmd/bot

test-build:
	go build ./...

lint:
	golangci-lint run

fmt:
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

test-all: test test-build lint fmt
	@echo "success"

