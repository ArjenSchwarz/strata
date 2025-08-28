# Task Completion Checklist for Strata

## MANDATORY: Before Completing Any Task

### 1. Code Validation (Run in this order)
```bash
# ALWAYS run the full validation suite
make check  # This runs fmt, vet, lint, and test

# OR run individually:
make fmt    # Format Go code
make vet    # Static analysis
make lint   # Code quality checks
make test   # Run all tests
```

### 2. Build Verification
```bash
# Confirm the project builds successfully
make build
```

### 3. For Go Code Changes
- ✅ Code is formatted with `make fmt`
- ✅ No vet issues with `make vet`
- ✅ No linting errors with `make lint`
- ✅ All tests pass with `make test`
- ✅ Project builds with `make build`
- ✅ Follow Go naming conventions (CamelCase for exported, camelCase for unexported)
- ✅ Errors are properly wrapped with context
- ✅ All exported types/functions are documented
- ✅ Use `any` instead of `interface{}`

### 4. For GitHub Action Changes
```bash
# Run action-specific tests
make test-action-unit        # Test shell modules
make test-action-integration # End-to-end tests
```

### 5. Sample Testing (if relevant)
```bash
# Test with sample files to verify functionality
make run-sample SAMPLE=basic-plan.json
make run-all-samples  # Quick validation of all samples
```

### 6. Documentation Updates
- Update README.md if adding new user-facing features
- Add implementation notes in docs/implementation/ for complex features
- Update CLAUDE.md if changing development workflow

### 7. Alternative Direct Commands (if Makefile unavailable)
```bash
gofmt ./...
go test ./... -v
golangci-lint run
go build -o strata
```

## Critical Rules
1. **NEVER** skip the validation steps
2. **ALWAYS** run `make check` before marking task complete
3. **DO NOT** commit if tests are failing
4. **DO NOT** commit if linting has errors
5. **NEVER** add comments unless explicitly requested
6. **ALWAYS** follow existing code patterns in the codebase

## Common Issues to Check
- Unused variables or imports (caught by linting)
- Improper error handling (wrap errors with context)
- Missing documentation on exported functions
- Non-idiomatic Go code (use modern patterns)
- Failing tests after changes

## Final Verification
Before considering task complete, confirm:
✅ make check passes without errors
✅ make build creates binary successfully
✅ No new warnings or errors introduced
✅ Code follows project conventions
✅ Tests cover new functionality