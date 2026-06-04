.PHONY: build test test-race test-short test-integration test-cover bench e2e manual

BINARY := clipboard-tui
PKG := ./cmd/clipboard-tui

build:
	go build -o $(BINARY) $(PKG)

test:
	go test -v ./...

test-short:
	go test -v -short ./...

test-race:
	go test -race -v ./...

test-integration:
	go test -v -tags=integration ./internal/...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

bench:
	go test -bench=. -benchmem ./internal/clipboard/... ./internal/llm/ollama/...

e2e:
	bash ./scripts/test-e2e.sh

manual:
	bash ./scripts/manual-test.sh
