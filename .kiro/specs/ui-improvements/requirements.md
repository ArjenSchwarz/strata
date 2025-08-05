# Requirements Document

## Introduction

This document outlines the requirements for improving the UI clarity of the Strata Terraform plan analysis tool. The goal is to enhance the user experience by making the output more readable, providing better visibility into high-risk changes, and adding support for markdown output format.

## Requirements

### Requirement 1: Improved Plan Information Display

**User Story:** As a Terraform user, I want a clearer plan information display with a horizontal layout, so that I can more easily scan the metadata about my plan.

#### Acceptance Criteria

1. WHEN displaying plan information THEN the system SHALL arrange items horizontally next to each other instead of vertically
2. WHEN displaying the Terraform Version THEN the system SHALL rename the key to just "Version"
3. WHEN displaying plan information THEN the system SHALL remove the dry run indicator

### Requirement 2: Enhanced Risk Visibility in Summary

**User Story:** As a Terraform user, I want to see high-risk changes highlighted in the summary, so that I can quickly identify potentially dangerous operations.

#### Acceptance Criteria

1. WHEN displaying the summary THEN the system SHALL add a "High Risk" column
2. WHEN calculating the "High Risk" column THEN the system SHALL count the number of sensitive items that have a danger flag

### Requirement 3: Always Show Sensitive Resource Changes

**User Story:** As a Terraform user, I want to always see sensitive resource changes even when detailed output is disabled, so that I never miss critical changes.

#### Acceptance Criteria

1. WHEN show-details is set to false THEN the system SHALL still display the resource changes overview
2. WHEN show-details is false THEN the system SHALL only show sensitive changes in the resource changes overview

### Requirement 4: Markdown Output Support

**User Story:** As a Terraform user, I want to generate markdown output, so that I can include plan summaries in documentation and pull requests.

#### Acceptance Criteria

1. WHEN the output format is set to markdown THEN the system SHALL use the go-output library to generate markdown
2. WHEN generating markdown output THEN the system SHALL ensure the output format is consistent with the table output