# Design Document: Sample Tests Organization & Makefile

## Overview

This design document outlines the approach for reorganizing sample test files and implementing a Makefile to streamline development workflows for the Strata project. The implementation will focus on two main components:

1. Creating a dedicated `samples` directory to better organize test files
2. Implementing a Makefile with standard targets for common development tasks

These changes will improve the project structure and provide developers with consistent, simplified commands for building, testing, and running sample files.

## Architecture

The implementation will follow the existing project structure while introducing the following changes:

1. **New Directory Structure**:
   ```
   strata/
   ├── ...
   ├── samples/           # New directory for sample test files
   │   ├── danger-sample.json
   │   ├── k8ssample.json
   │   ├── websample.json
   │   └── ...
   ├── Makefile           # New Makefile for development workflows
   └── ...
   ```

2. **Makefile Integration**:
   - The Makefile will be placed at the root of the project
   - It will provide standardized commands that wrap Go commands and CLI operations
   - Sample execution commands will reference files in the new samples directory

## Components and Interfaces

### 1. Sample Directory

The `samples` directory will serve as a container for all sample test files used during development and testing. This component:

- Provides a dedicated location for sample files
- Improves project organization by separating test data from application code
- Makes it easier to locate and manage test files

### 2. Makefile

The Makefile will provide a standardized interface for common development tasks. It will include the following targets:

1. **build**: Compiles the Strata application
   - Command: `make build`
   - Implementation: `go build .`

2. **test**: Runs all tests in the project
   - Command: `make test`
   - Implementation: `go test ./...`

3. **run-sample**: Executes a sample file with the plan summary command
   - Command: `make run-sample SAMPLE=<filename>`
   - Implementation: `go run . plan summary samples/${SAMPLE}`

4. **run-sample-verbose**: Executes a sample with verbose output
   - Command: `make run-sample-verbose SAMPLE=<filename>`
   - Implementation: `go run . plan summary -v samples/${SAMPLE}`

5. **run-sample-debug**: Executes a sample with verbose and debug output
   - Command: `make run-sample-verbose SAMPLE=<filename>`
   - Implementation: `go run . plan summary -v -d samples/${SAMPLE}`

## Data Models

This implementation does not introduce new data models or modify existing ones. It focuses on project organization and build process improvements.

## Error Handling

The Makefile will include appropriate error handling:

1. **Sample File Validation**:
   - Check if the specified sample file exists before execution
   - Provide meaningful error messages when a sample file is not found

2. **Command Execution**:
   - Ensure that command failures are properly reported
   - Use appropriate exit codes to indicate success or failure

## Testing Strategy

The implementation will be tested using the following approach:

1. **Manual Verification**:
   - Verify that all sample files are correctly moved to the samples directory
   - Confirm that relative paths in documentation and code are updated

2. **Makefile Testing**:
   - Test each Makefile target to ensure it executes the correct command
   - Verify that the Makefile handles errors appropriately
   - Test sample execution with various input files

3. **Integration Testing**:
   - Ensure that the reorganization does not break existing functionality
   - Verify that sample files can still be accessed and processed correctly

## Implementation Plan

1. Create the `samples` directory at the project root
2. Move all sample JSON files to the new directory
3. Create the Makefile with the required targets
4. Update any documentation or code references to sample files
5. Test all Makefile targets with various sample files

## Design Decisions and Rationale

1. **Samples Directory Location**:
   - Placing the samples directory at the project root provides easy access
   - This location is consistent with common Go project organization practices
   - It clearly separates test data from application code

2. **Makefile vs. Shell Scripts**:
   - A Makefile was chosen over shell scripts for cross-platform compatibility
   - Makefiles are a standard tool in many development environments
   - The target-based approach provides a clear, self-documenting interface

3. **Sample File Parameter Approach**:
   - Using a SAMPLE parameter allows for flexible file selection
   - This approach avoids hardcoding sample filenames in the Makefile
   - It provides a consistent interface for all sample execution commands