.PHONY: build test lint install clean

build:
	go build -o bin/scihub-pp-cli ./cmd/scihub-pp-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/scihub-pp-cli

clean:
	rm -rf bin/

build-mcp:
	go build -o bin/scihub-pp-mcp ./cmd/scihub-pp-mcp

install-mcp:
	go install ./cmd/scihub-pp-mcp

build-all: build build-mcp
