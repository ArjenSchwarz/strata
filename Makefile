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

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -f strata
	rm -rf dist/

# Install the application
install:
	go install .

# Update v1 tag to latest commit and push to GitHub
update-v1-tag:
	git tag -f v1
	git push origin v1 --force

# Show help
help:
	@echo "Available targets:"
	@echo "  build                 - Build the Strata application with version info"
	@echo "  build-release         - Build release version (requires VERSION=x.y.z)"
	@echo "  test                  - Run Go unit tests"
	@echo "  test-action-unit      - Run GitHub Action unit tests"
	@echo "  test-action-integration - Run GitHub Action integration tests"
	@echo "  test-action           - Run all GitHub Action tests"
	@echo "  run-sample            - Run sample with SAMPLE=<filename>"
	@echo "  run-sample-details    - Run sample with details SAMPLE=<filename>"
	@echo "  fmt                   - Format Go code"
	@echo "  lint                  - Run linter (requires golangci-lint)"
	@echo "  clean                 - Clean build artifacts"
	@echo "  install               - Install the application"
	@echo "  update-v1-tag         - Update v1 tag to latest commit and push to GitHub"
	@echo "  help                  - Show this help message"
	@echo ""
	@echo "Build examples:"
	@echo "  make build                    - Build with dev version"
	@echo "  make build VERSION=1.2.3     - Build with specific version"
	@echo "  make build-release VERSION=1.2.3 - Build release version"

.PHONY: build build-release test test-action-unit test-action-integration test-action run-sample run-sample-details fmt lint clean install update-v1-tag help
