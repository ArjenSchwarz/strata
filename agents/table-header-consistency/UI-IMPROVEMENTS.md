# UI/UX Improvements

## Summary
This analysis reviews table header consistency in Strata's Terraform plan summary output. The current implementation shows significant inconsistency across four different tables, using three different header capitalization styles (Title Case, ALL UPPERCASE, and all lowercase). This inconsistency creates a fragmented user experience that reduces readability and professional appearance across terminal, web, and documentation contexts.

## Critical Issues

### Issue: Inconsistent Table Header Capitalization Across Output
**Current State**: Four different tables use three different header capitalization styles:
- Plan Information: "Plan File | Version | Workspace | Backend | Created" (Title Case)
- Summary Statistics: "TOTAL CHANGES | ADDED | REMOVED | MODIFIED | REPLACEMENTS | HIGH RISK | UNMODIFIED" (ALL UPPERCASE)
- Resource Changes: "action | resource | type | id | replacement | module | danger | property_changes" (all lowercase)
- Output Changes: "NAME | ACTION | CURRENT | PLANNED | SENSITIVE" (ALL UPPERCASE)

**Problem**: This inconsistency severely impacts usability by:
- Creating cognitive dissonance as users scan between tables
- Appearing unprofessional and inconsistent
- Making the interface feel fragmented rather than cohesive
- Reducing scannability as users must mentally adjust to different formatting patterns
- Breaking visual hierarchy consistency across the document

**Recommendation**: Standardize all table headers to **Title Case** formatting
- Plan Information: "Plan File | Version | Workspace | Backend | Created" (no change)
- Summary Statistics: "Total Changes | Added | Removed | Modified | Replacements | High Risk | Unmodified"
- Resource Changes: "Action | Resource | Type | ID | Replacement | Module | Danger | Property Changes"
- Output Changes: "Name | Action | Current | Planned | Sensitive"

**Impact**: Users will experience consistent visual hierarchy, improved scannability, and a more professional, cohesive interface across all output formats.

**Implementation Notes**:
- Update header definitions in `/Users/arjenschwarz/projects/personal/strata/lib/plan/formatter.go` lines 276, 894, 1297, and 1374
- Ensure multi-word headers use proper Title Case (e.g., "Property Changes" not "Property_Changes")
- Test across all output formats (table, HTML, Markdown, JSON) to ensure consistency

## High Priority Improvements

### Issue: Inconsistent Multi-Word Header Formatting
**Current State**: Multi-word headers are inconsistently formatted:
- "property_changes" (snake_case)
- "HIGH RISK" (UPPERCASE with space)
- "TOTAL CHANGES" (UPPERCASE with space)

**Problem**: Mixed formatting conventions within the same interface reduce professional appearance and make headers harder to parse visually.

**Recommendation**: Use consistent Title Case with proper spacing for all multi-word headers:
- "Property Changes" instead of "property_changes"
- "High Risk" instead of "HIGH RISK"
- "Total Changes" instead of "TOTAL CHANGES"

**Impact**: Improved readability and professional appearance with consistent formatting conventions.

### Issue: Header Capitalization Choice Analysis
**Current State**: Three different capitalization approaches are used inconsistently.

**Problem**: Each approach has specific UX implications:
- **ALL UPPERCASE**: More visually prominent but can feel aggressive/shouting in text interfaces; harder to read for extended periods
- **all lowercase**: Modern/minimalist but can appear unprofessional in business contexts; harder to distinguish from data
- **Title Case**: Professional, readable, follows standard UI conventions; works well across all contexts

**Recommendation**: Title Case provides the optimal balance for this CLI tool because:
- Professional appearance suitable for business/infrastructure contexts
- Excellent readability in terminal environments with limited typography
- Consistent with standard UI design patterns
- Works well across all output formats (terminal, HTML, Markdown, CSV)
- Maintains clear hierarchy without being visually aggressive
- Accessible and familiar to users across different platforms

**Impact**: Headers will be more readable, professional, and consistent with user expectations across different viewing contexts.

## Medium Priority Enhancements

### Issue: Header Semantic Clarity
**Current State**: Some headers could be more descriptive:
- "ID" vs "Resource ID"
- "Danger" vs "Risk Level"
- "Action" could be more specific in context

**Problem**: Abbreviated headers may be unclear to new users or in wide tables where context is lost.

**Recommendation**: Consider slightly more descriptive headers where space allows:
- "Resource ID" instead of "ID"
- "Risk Level" instead of "Danger"
- Keep "Action" as it's clear in Terraform context

**Impact**: Improved clarity for new users while maintaining compact display for experienced users.

## Positive Observations

### Well-Executed Header Content
The header content choices are generally appropriate:
- Plan Information headers capture essential metadata clearly
- Summary Statistics headers provide clear numerical categories
- Resource Changes headers cover all necessary change dimensions
- Output Changes headers follow Terraform conventions effectively

### Functional Column Organization
The column organization within each table is logical and follows user mental models:
- Plan Information flows from specific to general context
- Summary Statistics presents metrics in logical order
- Resource Changes prioritizes action type and resource identification
- Output Changes follows a clear before/after pattern

### Comprehensive Information Architecture
The table structure successfully captures all necessary information for Terraform plan analysis without overwhelming users with excessive detail in the header level.

## Implementation Priority
1. **Critical**: Standardize all headers to Title Case (immediate impact on user experience)
2. **High**: Fix multi-word header formatting consistency
3. **Medium**: Verify cross-format consistency
4. **Low**: Consider semantic clarity improvements and visual weight enhancements

This consolidation will significantly improve the professional appearance and usability of Strata's table output across all supported formats.