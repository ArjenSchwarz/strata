# Decision Log - Simplified Plan Rendering

## Context
This log captures key decisions made during the requirements gathering and design phases for simplifying the plan rendering architecture and fixing the multi-table markdown rendering bug.

## Original Requirements Phase Decisions

### 1. Use of builder.Table() vs output.NewTableContent()
**Date**: 2025-07-28
**Decision**: Prefer output.NewTableContent() over builder.Table()
**Rationale**: 
- The working example that successfully renders multiple tables uses NewTableContent exclusively
- builder.Table() appears to be a convenience method that may not properly handle multiple tables in certain formats
- NewTableContent provides more control over table schema and formatting
**Trade-offs**: May require more verbose code but provides better control

### 2. Renderer Architecture
**Date**: 2025-07-28
**Decision**: Use standard renderers with collapsible content support, not separate collapsible renderers
**Rationale**:
- The working example shows collapsible content works with standard markdown renderer
- Separate renderer paths create unnecessary complexity
- The go-output v2 library supports collapsible content through data preparation, not renderer selection
**Clarification**: This doesn't mean removing collapsible functionality - it means using the standard renderer that already supports collapsible content when properly formatted

### 3. Format-Specific Logic
**Date**: 2025-07-28
**Decision**: Remove format-specific logic from Strata code - let go-output handle format differences
**Rationale**:
- Format handling is go-output's responsibility, not Strata's
- Strata should prepare data consistently and let go-output handle format-specific rendering
- Configuration differences should be handled through go-output's format capabilities metadata
**Implementation**: Single code path for all formats, with go-output managing format-specific behavior

### 4. Scope of Simplification
**Date**: 2025-07-28
**Decision**: Simplify architecture AND fix bug as a single effort
**Rationale**:
- Previous attempts to fix the bug within the current complex architecture have failed
- The complexity itself is preventing the bug fix
- The architectural issues are the root cause of the rendering problems
**Approach**: Complete architectural simplification following the working example pattern, which will inherently fix the multi-table rendering issue

### 5. Performance Considerations
**Date**: 2025-07-28
**Decision**: Maintain current performance levels - no regression
**Rationale**:
- Performance is acceptable in current implementation
- Focus is on correctness and maintainability
- Any performance optimizations are bonus, not requirement
**Approach**: Ensure no performance degradation through testing

### 6. Migration Strategy
**Date**: 2025-07-28
**Decision**: No migration strategy - clean replacement
**Rationale**:
- The configuration format is not changing
- The user interface remains the same
- Clean break is simpler than maintaining compatibility code
- Previous attempts to work within existing architecture failed
**Implementation**: Replace implementation entirely

### 7. Provider Grouping Threshold
**Date**: 2025-07-28
**Decision**: Maintain existing threshold check behavior
**Rationale**:
- Current behavior is well-understood by users
- Threshold provides control over when grouping occurs
- Removing it would be a behavior change beyond bug fixing
**Implementation**: Keep threshold check in simplified implementation

### 8. Testing Strategy
**Date**: 2025-07-28
**Decision**: Comprehensive test suite required before implementation
**Rationale**:
- Multi-table rendering is core functionality
- Need to prevent regression
- Tests will validate the fix works across all formats
**Approach**: Create tests for all formats, edge cases, and performance

### 9. Collapsible Content in Non-Supporting Formats
**Date**: 2025-07-28
**Decision**: Delegate format capability handling to go-output
**Rationale**:
- go-output already handles format differences correctly
- Strata should not duplicate this logic
- Keeps Strata code format-agnostic
**Implementation**: Trust go-output to handle format capabilities

### 10. ActionSortTransformer Location
**Date**: 2025-07-28
**Decision**: Keep ActionSortTransformer as post-processing transformer
**Rationale**:
- Current location works correctly
- Moving it would expand scope beyond bug fix
- Can be reconsidered in future refactoring
**Implementation**: No change to transformer architecture

## Design Phase Decisions

### 11. Root Cause Analysis Method
**Date**: 2025-07-28
**Decision**: Investigate actual cause through code analysis and proof-of-concept testing
**Rationale**:
- Design-critic challenged assumption about library bug
- Need evidence-based approach rather than assumptions
- Must validate go-output v2 pattern actually works for our use case
**Evidence**: Testing confirmed multi-table rendering works correctly with proper API usage

### 12. Architectural Approach: Targeted Fix vs Complete Rewrite
**Date**: 2025-07-28
**Decision**: Use targeted fix approach instead of complete architectural overhaul
**Rationale**:
- Root cause analysis revealed the issue is artificially disabled tables, not fundamental architecture problem
- Targeted fix has much lower risk while achieving same primary goal
- Existing functionality (collapsible content, provider grouping) already works correctly
- Design-critic feedback exposed risks of over-engineering the solution
**Implementation**: Focus on re-enabling tables and using consistent NewTableContent pattern

### 13. Problem Classification: Bug Fix vs Architecture Simplification
**Date**: 2025-07-28
**Decision**: Treat as primarily a bug fix with secondary simplification benefits
**Rationale**:
- Multi-table rendering "bug" was actually just disabled code (lines 189-191 in formatter.go)
- Comment claimed "go-output v2 multi-table rendering bug" but this doesn't exist
- Proof-of-concept confirmed three tables render correctly in markdown when using proper pattern
**Impact**: Reduced scope and risk while maintaining all benefits

### 14. Validation Requirements
**Date**: 2025-07-28
**Decision**: Require working proof-of-concept before proceeding with design
**Rationale**:
- Design-critic correctly identified that entire strategy depended on unvalidated assumption
- Must prove go-output v2 collapsible-tables example pattern works for Strata's data structures
- Evidence-based design decisions are more reliable than theoretical ones
**Evidence**: Created and tested proof-of-concept confirming multi-table markdown rendering works

### 15. Error Handling Strategy: Conservative vs Aggressive
**Date**: 2025-07-28
**Decision**: Use conservative error handling approach
**Rationale**:
- Targeted fix should maintain existing error handling patterns where possible
- For critical tables (Plan Information, Statistics), log warnings but continue operation
- Avoid changing too many behaviors simultaneously
- Preserve existing robustness while fixing the core issue
**Implementation**: Graceful degradation for individual table failures

### 16. Code Consolidation Priority
**Date**: 2025-07-28
**Decision**: Prioritize functionality over code cleanup
**Rationale**:
- Primary goal is fixing multi-table rendering bug
- Code cleanup (removing duplicate functions) is secondary benefit
- Keep working duplicate functions if they provide stable functionality
- Remove only proven unnecessary duplicates (e.g., keep "Direct" formatter versions)
**Approach**: Conservative cleanup focused on clear duplicates

### 17. Testing Focus: Regression Prevention vs New Features
**Date**: 2025-07-28
**Decision**: Emphasize regression testing over comprehensive new feature testing
**Rationale**:
- This is primarily a bug fix, not new feature development
- Must ensure all existing functionality continues working exactly as before
- New functionality is limited to re-enabling disabled tables
- Performance and behavior must remain identical except for the fix
**Implementation**: Comprehensive regression test suite with specific multi-table markdown test

### 18. Implementation Phases: Sequential vs Parallel
**Date**: 2025-07-28
**Decision**: Use sequential phased approach with validation at each step
**Rationale**:
- Allows validation that each change works before proceeding
- Enables rollback to specific working state if issues arise
- Reduces debugging complexity by isolating changes
- Aligns with conservative, low-risk approach
**Phases**: Enable tables → Unify format handling → Code cleanup → Testing

### 19. Documentation Strategy: Comprehensive vs Targeted
**Date**: 2025-07-28
**Decision**: Focus documentation on specific changes made and rationale
**Rationale**:
- Targeted fix doesn't require comprehensive architectural documentation
- Should document the root cause discovery and solution validation
- Important to record why the "library bug" assumption was incorrect
- Help prevent future similar issues
**Approach**: Clear decision log, specific implementation notes, validation evidence