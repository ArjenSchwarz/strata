# Data Pipeline Integration Requirements

## Introduction

The Data Pipeline Integration feature replaces the hacky ActionSortTransformer that currently parses rendered table strings to re-sort them. By using go-output v2's data transformation pipeline, we can sort data at the structural level before rendering, eliminating fragile regex patterns and string manipulation.

## Requirements

### 1. Replace ActionSortTransformer with Pipeline Sorting

**User Story:** As a developer, I want to eliminate the hacky string parsing in ActionSortTransformer by sorting data before it's rendered, so that the code is more maintainable and reliable.

**Acceptance Criteria:**
1.1. The system SHALL replace ActionSortTransformer's string parsing with pipeline-based data sorting
1.2. The system SHALL apply sorting at the data level using go-output v2's SortWith operation
1.3. The sorting logic SHALL remain unchanged: danger indicators first, then action priority, then alphabetically
1.4. The system SHALL produce identical output to the current implementation
1.5. The system SHALL remove all regex patterns for identifying table rows and columns
1.6. The system SHALL eliminate the need to parse "| action" and "| resource" patterns

### 2. Pipeline Integration in Formatter

**User Story:** As a maintainer, I want the pipeline operations integrated cleanly into the existing formatter flow, so that the change is transparent to users.

**Acceptance Criteria:**
2.1. The system SHALL apply pipeline sorting within prepareResourceTableData or similar data preparation methods
2.2. The system SHALL use the pipeline's SortWith method with a custom comparator function
2.3. The system SHALL maintain compatibility with all existing output formats
2.4. The system SHALL work with the existing collapsible content system
2.5. The system SHALL preserve the existing provider grouping functionality

### 3. Remove Byte-Level Transformer Infrastructure

**User Story:** As a developer, I want to remove the now-unnecessary ActionSortTransformer class and its associated infrastructure, so that the codebase is cleaner.

**Acceptance Criteria:**
3.1. The system SHALL remove the ActionSortTransformer struct and all its methods
3.2. The system SHALL remove the registration of ActionSortTransformer in the output pipeline
3.3. The system SHALL remove cached regex patterns used for string parsing
3.4. The system SHALL remove the hasDangerIndicator string parsing function if it's only used by ActionSortTransformer
3.5. The system SHALL keep other transformers (EmojiTransformer, ColorTransformer) unchanged

### 4. Testing and Validation

**User Story:** As a maintainer, I want comprehensive tests proving the pipeline implementation produces identical results, so that I can be confident nothing breaks.

**Acceptance Criteria:**
4.1. The system SHALL include tests comparing pipeline-sorted output with current sorted output
4.2. The system SHALL verify sorting works correctly for danger indicators, actions, and resource names
4.3. The system SHALL test with all sample plan files to ensure compatibility
4.4. The system SHALL include benchmarks showing improved performance over string parsing