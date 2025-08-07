# Strata Makefile
# Provides standard targets for common development tasks

# Version information
VERSION ?= dev
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags for version injection
LDFLAGS := -X github.com/ArjenSchwarz/strata/cmd.Version=$(VERSION) \
           -X github.com/ArjenSchwarz/strata/cmd.BuildTime=$(BUILD_TIME) \
           -X github.com/ArjenSchwarz/strata/cmd.GitCommit=$(GIT_COMMIT)

# Build the Strata application
build:
	go build -ldflags "$(LDFLAGS)" .

# Build the Strata application with version information
build-release:
	@if [ "$(VERSION)" = "dev" ]; then \
		echo "Error: VERSION must be set for release builds. Usage: make build-release VERSION=1.2.3"; \
		exit 1; \
	fi
	go build -ldflags "$(LDFLAGS)" -o strata .

# Run all tests
test:
	go test ./...

# Run tests with verbose output and coverage
test-verbose:
	go test -v -cover ./...

# Generate test coverage report
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

# Run benchmark tests
benchmarks:
	go test -bench=. ./...

# Run a sample file with the plan summary command
run-sample:
	@if [ -z "$(SAMPLE)" ]; then \
		echo "Error: SAMPLE parameter is required. Usage: make run-sample SAMPLE=<filename>"; \
		exit 1; \
	fi
	@if [ ! -f "samples/$(SAMPLE)" ]; then \
		echo "Error: Sample file samples/$(SAMPLE) not found"; \
		exit 1; \
	fi
	go run . plan summary samples/$(SAMPLE)

# Run a sample file with verbose output
run-sample-details:
	@if [ -z "$(SAMPLE)" ]; then \
		echo "Error: SAMPLE parameter is required. Usage: make run-sample-details SAMPLE=<filename>"; \
		exit 1; \
	fi
	@if [ ! -f "samples/$(SAMPLE)" ]; then \
		echo "Error: Sample file samples/$(SAMPLE) not found"; \
		exit 1; \
	fi
	go run . plan summary --details samples/$(SAMPLE)

# List available sample files
list-samples:
	@echo "Available sample files:"
	@ls -1 samples/ | sed 's/^/  /'

# Run all sample files quickly (table output)
run-all-samples:
	@echo "Testing all sample files..."
	@for sample in samples/*.json; do \
		echo "Testing $$(basename $$sample)..."; \
		go run . plan summary $$sample > /dev/null && echo "✓ $$(basename $$sample)" || echo "✗ $$(basename $$sample)"; \
	done

# Run unit tests for GitHub Action components
test-action-unit:
	./test/action_test.sh

# Run integration tests for GitHub Action
test-action-integration:
	./test/integration_test.sh

# Run all action tests
test-action: test-action-unit test-action-integration

# Format Go code
fmt:
	go fmt ./...

# Run go vet for static analysis
vet:
	go vet ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Run full validation suite
check: fmt vet lint test

# Clean build artifacts
clean:
	rm -f strata
	rm -rf dist/
	rm -f coverage.out coverage.html

# Install the application
install:
	go install .

# Clean up go.mod and go.sum
deps-tidy:
	go mod tidy

# Update dependencies to latest versions
deps-update:
	go get -u ./...
	go mod tidy

# Run security scan (requires gosec)
security-scan:
	@which gosec > /dev/null || (echo "gosec not installed. Run: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest" && exit 1)
	gosec ./...

go-functions:
	@echo "Finding all functions in the project..."
	@grep -r "^func " . --include="*.go" | grep -v vendor/

# Update v1 tag to latest commit and push to GitHub
update-v1-tag:
	git tag -f v1
	git push origin v1 --force

# Show help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  build                 - Build the Strata application with version info"
	@echo "  build-release         - Build release version (requires VERSION=x.y.z)"
	@echo "  install               - Install the application"
	@echo "  clean                 - Clean build artifacts and coverage files"
	@echo ""
	@echo "Testing targets:"
	@echo "  test                  - Run Go unit tests"
	@echo "  test-verbose          - Run tests with verbose output and coverage"
	@echo "  test-coverage         - Generate test coverage report (HTML)"
	@echo "  benchmarks            - Run benchmark tests"
	@echo "  test-action-unit      - Run GitHub Action unit tests"
	@echo "  test-action-integration - Run GitHub Action integration tests"
	@echo "  test-action           - Run all GitHub Action tests"
	@echo ""
	@echo "Code quality targets:"
	@echo "  fmt                   - Format Go code"
	@echo "  vet                   - Run go vet for static analysis"
	@echo "  lint                  - Run linter (requires golangci-lint)"
	@echo "  check                 - Run full validation suite (fmt, vet, lint, test)"
	@echo "  security-scan         - Run security analysis (requires gosec)"
	@echo ""
	@echo "Sample testing targets:"
	@echo "  list-samples          - List available sample files"
	@echo "  run-sample            - Run sample with SAMPLE=<filename>"
	@echo "  run-sample-details    - Run sample with details SAMPLE=<filename>"
	@echo "  run-all-samples       - Test all sample files quickly"
	@echo ""
	@echo "Dependency management:"
	@echo "  deps-tidy             - Clean up go.mod and go.sum"
	@echo "  deps-update           - Update dependencies to latest versions"
	@echo ""
	@echo "Development utilities:"
	@echo "  go-functions          - List all Go functions in the project"
	@echo "  update-v1-tag         - Update v1 tag to latest commit and push to GitHub"
	@echo "  help                  - Show this help message"
	@echo ""
	@echo "Build examples:"
	@echo "  make build                    - Build with dev version"
	@echo "  make build VERSION=1.2.3     - Build with specific version"
	@echo "  make build-release VERSION=1.2.3 - Build release version"

.PHONY: build build-release test test-verbose test-coverage benchmarks test-action-unit test-action-integration test-action run-sample run-sample-details list-samples run-all-samples fmt vet lint check clean install deps-tidy deps-update security-scan go-functions update-v1-tag help
