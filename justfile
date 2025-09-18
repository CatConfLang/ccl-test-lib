# CCL Test Library Justfile

# Show available commands
default:
    @just --list

# Core aliases
alias t := test
alias b := build
alias l := lint
alias f := format
alias c := clean

# === BUILD ===

# Sync schemas from ccl-test-data repository
sync-schemas:
    go run cmd/schema-sync/main.go schemas

# Generate Go types from schemas
generate: sync-schemas
    go generate ./...

# Build all packages
build: generate
    go build ./...

# Build examples
build-examples:
    go build -o bin/basic-example examples/basic/basic_usage.go
    go build -o bin/ccl-data-example examples/ccl-test-data/ccl-test-data_usage.go

# Install dependencies
deps:
    go mod download
    go mod tidy

# === DEVELOPMENT WORKFLOW ===

# Quick development check: format, lint, build, test
dev:
    just format
    just lint
    just generate
    just build
    just test

# Complete CI pipeline: all checks and validation
ci:
    just deps
    just format-check
    just lint
    just generate
    just build
    just test
    just vet

# === TESTING ===

# Run all tests
test:
    go test ./...

# Run tests with verbose output
test-verbose:
    go test -v ./...

# Run tests with coverage
test-coverage:
    go test -cover ./...
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
    go test -bench=. ./...

# === CODE QUALITY ===

# Format code
format:
    go fmt ./...
    gofmt -w .

# Check if code is formatted
format-check:
    #!/usr/bin/env bash
    if [ -n "$(gofmt -l .)" ]; then
        echo "Code is not formatted. Run 'just format' to fix."
        gofmt -l .
        exit 1
    fi

# Lint code
lint:
    go mod tidy
    go fmt ./...
    go vet ./...

# Advanced static analysis
vet:
    go vet ./...

# === EXAMPLES ===

# Run basic usage example
run-basic:
    go run examples/basic/basic_usage.go

# Run ccl-test-data usage example
run-ccl-data:
    go run examples/ccl-test-data/ccl-test-data_usage.go

# Run all examples
run-examples:
    @echo "=== Running Basic Example ==="
    just run-basic
    @echo "=== Running CCL Test Data Example ==="
    just run-ccl-data

# === UTILITIES ===

# Clean build artifacts and temporary files
clean:
    go clean ./...
    rm -rf bin/
    rm -f coverage.out coverage.html
    rm -f test-basic test-ccl-data

# Show module information
mod-info:
    go mod graph
    go list -m all

# Show package dependencies
deps-graph:
    go list -deps ./...

# Verify module integrity
mod-verify:
    go mod verify

# Update dependencies
deps-update:
    go get -u ./...
    go mod tidy

# === DOCUMENTATION ===

# Generate Go documentation
docs:
    godoc -http=:6060

# Show package documentation
docs-show PACKAGE="":
    #!/usr/bin/env bash
    if [ -z "{{PACKAGE}}" ]; then
        go doc github.com/tylerbu/ccl-test-lib
    else
        go doc github.com/tylerbu/ccl-test-lib/{{PACKAGE}}
    fi

# === RELEASE ===

# Prepare for release: full validation
release-check:
    just clean
    just deps
    just format-check
    just lint
    just vet
    just build
    just build-examples
    just test-coverage
    just run-examples
    @echo "✅ Release ready!"

# Tag a new version
tag VERSION:
    git tag -a {{VERSION}} -m "Release {{VERSION}}"
    git push origin {{VERSION}}

# === PROJECT HEALTH ===

# Show project statistics
stats:
    @echo "=== Lines of Code ==="
    find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -1
    @echo "=== Package Count ==="
    go list ./... | wc -l
    @echo "=== Test Coverage ==="
    go test -cover ./... | grep coverage

# Security scan
security:
    go list -json -m all | nancy sleuth

# Performance profile
profile:
    go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...
    @echo "Use 'go tool pprof cpu.prof' or 'go tool pprof mem.prof' to analyze"