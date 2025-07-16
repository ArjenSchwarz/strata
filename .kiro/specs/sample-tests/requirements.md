# Requirements Document

## Introduction

This document outlines the requirements for reorganizing sample tests and creating a Makefile to streamline development workflows for the Strata project.

## Requirements

### Requirement 1: Sample Tests Organization

**User Story:** As a developer, I want sample tests organized in a dedicated directory, so that test files are better structured and easier to locate.

#### Acceptance Criteria

1. WHEN the project is built THEN the system SHALL have sample test files moved to a directory called 'samples'

### Requirement 2: Makefile Creation

**User Story:** As a developer, I want a Makefile with standard targets, so that I can execute common development tasks with simple commands.

#### Acceptance Criteria

1. WHEN running 'make build' THEN the system SHALL build the Strata application (equivalent to 'go build .')
2. WHEN running 'make test' THEN the system SHALL execute all tests (equivalent to 'go test .')
3. WHEN running 'make run-sample SAMPLE=<filename>' THEN the system SHALL execute the sample file with the plan summary command (equivalent to 'go run . plan summary ${sample file}')
4. WHEN running 'make run-sample-verbose SAMPLE=<filename>' THEN the system SHALL execute the sample with verbose output (equivalent to 'go run . plan summary -v ${sample file}')
5. WHEN running 'make run-sample-debug SAMPLE=<filename>' THEN the system SHALL execute the sample with verbose and debug output (equivalent to 'go run . plan summary -v -d ${sample file}')