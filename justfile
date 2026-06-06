# Use the os() function to determine the binary extension [1, 2]
binary_ext := if os() == "windows" { ".exe" } else { "" }
BIN_NAME := "quick-agent"
BINARY := BIN_NAME + binary_ext
BIN_PATH := "./cmd/quick-agent"

# Default recipe
default:
    just --list

# Build the binary with the platform-specific extension
build version="dev":
    go build -ldflags "-s -w -X github.com/mberetvas/quick-agent/internal/version.Version={{version}}" -o {{BINARY}} {{BIN_PATH}}
    @echo "✓ Built {{BINARY}}"

# Build all arch binaries for the current OS into dist/
build-dist version="dev":
    VERSION={{version}} bash ./scripts/build.sh

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

# Fail if total statement coverage is below 80%
cover-check min="80":
    go test -coverprofile=coverage.out ./...
    @bash -eu -c 'pct=$(go tool cover -func=coverage.out | awk "/^total:/ {gsub(/%/,\"\",\$3); print \$3}"); echo "Total coverage: ${pct}%"; awk "BEGIN {exit !(${pct} >= {{min}})}" || { echo "FAIL: coverage ${pct}% < {{min}}%"; exit 1; }; echo "✓ Coverage gate passed"'

# Full CI pipeline
ci: deps check build
    @echo "✓ CI pipeline passed"