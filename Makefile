# Strata Makefile
# Provides standard targets for common development tasks

# Build the Strata application
build:
	go build .

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

# Show help
help:
	@echo "Available targets:"
	@echo "  build                 - Build the Strata application"
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
	@echo "  help                  - Show this help message"

.PHONY: build test test-action-unit test-action-integration test-action run-sample run-sample-details fmt lint clean install help
