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
