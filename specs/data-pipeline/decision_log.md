# Data Pipeline Integration Decision Log

## Decision 1: Focus on Existing Functionality Only

**Date:** 2025-09-16
**Decision:** Limit requirements to improving current functionality instead of adding new features
**Rationale:** The user emphasized keeping it simple and not overcomplicating the implementation. Future features should be added when they are actually needed.
**Impact:** Moved advanced filtering, custom calculations, and complex configuration to future-ideas.md

## Decision 2: Remove Performance Targets

**Date:** 2025-09-16
**Decision:** Remove specific performance improvement targets (like "30% improvement for 100+ resources")
**Rationale:** User pointed out these were made up and not actually set as requirements
**Impact:** Focused on general performance improvement without specific metrics

## Decision 3: No Backward Compatibility Requirements

**Date:** 2025-09-16
**Decision:** Remove backward compatibility requirements for transformers
**Rationale:** User indicated that if the output is the same or better, users don't care about the implementation details
**Impact:** Simplified migration approach - just replace the implementation without complex compatibility layers

## Decision 4: No Custom Sorting Configuration

**Date:** 2025-09-16
**Decision:** Remove custom sorting orders and user-defined sort criteria
**Rationale:** User emphasized this is "an opinionated application" and we're not building a spreadsheet
**Impact:** Keep the existing sort order (danger > action priority > alphabetical) without customization

## Decision 5: Basic Filtering Only

**Date:** 2025-09-16
**Decision:** Limit filtering to two simple flags: --filter-dangerous and --filter-destructive
**Rationale:** User rejected complex filtering expressions and multiple filter types as overcomplicating
**Impact:** Removed resource type filtering, complex filter combinations, and custom filter expressions

## Decision 6: No Interactive Features

**Date:** 2025-09-16
**Decision:** Remove any interactive features or on-the-fly modifications
**Rationale:** User explicitly stated "No interactive features right now"
**Impact:** Focus purely on static configuration and CLI flags

## Decision 7: Delegate Format Differences to go-output

**Date:** 2025-09-16
**Decision:** Let go-output library handle format-specific behavior instead of implementing it in Strata
**Rationale:** User indicated format differences should be handled by the library, not the application
**Impact:** Simplified pipeline implementation by relying on go-output's format handling

## Decision 8: Remove Risk Scores

**Date:** 2025-09-16
**Decision:** Remove calculated risk scores from requirements
**Rationale:** User questioned what risk scores are and why they were suddenly introduced
**Impact:** Removed the entire "Calculated Fields" section focusing on risk scores and dependency analysis

## Decision 9: Focus Only on Fixing ActionSortTransformer

**Date:** 2025-09-16
**Decision:** Narrow scope to only replacing the hacky ActionSortTransformer string parsing
**Rationale:** User clarified they want to fix the ugly/hacky custom sorting, not add new features
**Impact:** Removed filtering, grouping, and other features from requirements. Focus only on replacing string parsing with data-level sorting

## Decision 10: Keep Requirements Minimal

**Date:** 2025-09-16
**Decision:** Only include what makes sense to improve about the existing hacky sorting
**Rationale:** User was clear about looking at existing functionality that needs improvement, not adding features
**Impact:** Requirements now focus solely on replacing ActionSortTransformer with pipeline sorting