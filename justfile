# Use the os() function to determine the binary extension [1, 2]
binary_ext := if os() == "windows" { ".exe" } else { "" }
BIN_NAME := "clipboard-tui"
BINARY := BIN_NAME + binary_ext
BIN_PATH := "./cmd/clipboard-tui"

# Default recipe
default:
    just --list

# Build the binary with the platform-specific extension
build:
    go build -o {{BINARY}} {{BIN_PATH}}
    @echo "✓ Built {{BINARY}}"

# Run the TUI
run:
    go run {{BIN_PATH}} tui

# Run the daemon
daemon:
    go run {{BIN_PATH}} daemon

# Run with help
help:
    go run {{BIN_PATH}} --help

# Run all tests
test:
    go test ./...

# Run tests with verbose output
test-verbose:
    go test -v ./...

# Run tests with coverage
test-coverage:
    go test -cover ./...

# Generate coverage report (HTML)
coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out
    @echo "✓ Coverage report generated (coverage.out)"

# Run tests for specific package
test-pkg package:
    go test -v ./internal/{{package}}

# Run specific test
test-func func:
    go test -v -run {{func}} ./...

# Format code
fmt:
    go fmt ./...
    @echo "✓ Code formatted"

# Lint code
lint:
    golangci-lint run ./...

# Run fmt + vet + tests
check: fmt
    go vet ./...
    go test ./...
    @echo "✓ All checks passed"

# Clean build artifacts - rm works on Windows if sh is available [3]
clean:
    rm -f {{BINARY}} coverage.out
    @echo "✓ Cleaned artifacts"

# Install dependencies
deps:
    go mod download
    go mod tidy
    @echo "✓ Dependencies installed"

# Build and run the binary using the correct extension
run-binary: build
    ./{{BINARY}} tui

# Full CI pipeline
ci: deps check build
    @echo "✓ CI pipeline passed"