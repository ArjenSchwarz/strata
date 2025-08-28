# Output Refinements - Decision Log

## Decision History

### 2025-01-06: Feature Scope Definition

**Decision:** Group all five open issues (#17-#21) into a single "output-refinements" feature.

**Rationale:** 
- All issues relate to improving the plan summary output quality and usability
- They share common code paths in the formatter and analyzer modules
- Implementing them together ensures consistent behavior across all output improvements
- Reduces risk of conflicts between separate implementations

**Alternatives Considered:**
- Separate features for each issue (rejected: too granular, would cause merge conflicts)
- Group by type (sorting vs. display) (rejected: artificial separation, issues are interrelated)

---

### 2025-01-06: Default No-Op Behavior

**Decision:** Hide no-ops by default, provide opt-in flag to show them.

**Rationale:**
- Most users want to focus on actual changes
- Reduces noise in the output for typical use cases
- Follows principle of progressive disclosure
- Aligns with similar tools (GitHub PR diffs hide unchanged files by default)

**Alternatives Considered:**
- Show no-ops by default (rejected: too verbose for most users)
- Auto-detect based on plan size (rejected: unpredictable behavior)

---

### 2025-01-06: Sensitive Value Masking

**Decision:** Always mask sensitive values with "(sensitive value)" text.

**Rationale:**
- Security by default principle
- Consistent with Terraform's own behavior
- Prevents accidental exposure in logs, screenshots, or CI/CD outputs
- Simple, clear indication that data is hidden

**Alternatives Considered:**
- Hash the values (rejected: could still leak information through hash comparison)
- Show partial values (rejected: security risk)
- Make masking configurable (rejected: security should not be optional)

---

### 2025-01-06: Sorting Priority Order

**Decision:** Sort by: 1) Sensitivity/Danger, 2) Action Type (delete > replace > update > create), 3) Alphabetical

**Rationale:**
- Puts most critical changes first for immediate attention
- Destructive operations (delete/replace) are more risky than constructive ones
- Alphabetical provides predictable fallback ordering
- Maintains existing user expectations from before the regression

**Alternatives Considered:**
- Pure alphabetical (rejected: loses risk prioritization)
- Action type only (rejected: doesn't highlight sensitive resources)
- User-configurable order (rejected: too complex for initial implementation)

---

## Clarification Decisions - 2025-01-06

### JSON Sensitive Value Handling

**Decision:** Follow Terraform's approach in `terraform show -json` for handling sensitive values in JSON output.

**Rationale:**
- Maintains compatibility with existing Terraform tooling
- Users expect consistent behavior with Terraform
- Preserves type information where possible

---

### Provider Grouping and Sorting

**Decision:** Apply sorting within each provider group independently.

**Rationale:**
- Maintains logical grouping while preserving priority ordering
- Each provider group becomes a self-contained sorted unit
- Simpler implementation that avoids complex cross-group logic

---

### Property Sorting Scope

**Decision:** Apply alphabetical sorting only to property changes within each individual resource.

**Rationale:**
- Addresses the specific issue raised (#21)
- Doesn't disrupt other summary sections
- Maintains existing behavior for "top changes" and other summaries

---

### No-Ops and Statistics

**Decision:** No-ops continue to be counted in statistics regardless of visibility setting.

**Rationale:**
- Preserves existing behavior
- Statistics reflect the complete plan, not just visible elements
- Avoids confusion with changing counts based on display settings

---

### Sensitivity Priority for Sorting

**Decision:** Sort with priority: 1) Dangerous resources (IsDangerous), 2) Resources with sensitive properties, 3) All other resources.

**Rationale:**
- Dangerous resources (user-configured) represent highest risk
- Sensitive properties are next priority
- Clear, predictable ordering

---

### Flag Independence

**Decision:** `--show-no-ops` operates independently of `--details` flag.

**Rationale:**
- Provides maximum flexibility
- Users can see no-ops without full property details
- Follows principle of orthogonal features

---

## Stakeholder Decisions - 2025-01-06

### Compliance Requirements for Sensitive Data

**Decision:** Match Terraform's standard behavior for sensitive data handling.

**Rationale:**
- No special compliance requirements beyond Terraform's defaults
- Consistency with Terraform tooling is the primary goal
- Users already familiar with Terraform's masking approach

---

### Sorting Behavior Dependencies

**Decision:** No existing integrations depend on current sorting behavior.

**Rationale:**
- Current sorting is considered broken/regressed
- Safe to fix without compatibility concerns
- Restoring expected behavior is a bug fix, not breaking change

---

### No-Op Visibility Configuration Scope

**Decision:** Global configuration only, not per-resource-type.

**Rationale:**
- Keeps implementation simple
- Avoids configuration complexity
- Meets user needs without overcomplication

---

### Sort Order Appropriateness

**Decision:** Proposed sort order (sensitivity > action > alphabetical) is approved.

**Rationale:**
- Provides good default behavior for most use cases
- Simple, predictable ordering
- No need for user-configurable sorting

---

### No-Op Export Feature

**Decision:** Do not implement no-op export to separate files.

**Rationale:**
- Not needed for current use cases
- Adds unnecessary complexity
- Can be reconsidered if future need arises

---

### Configurable Sort Order

**Decision:** Sorting order will not be configurable.

**Rationale:**
- Single, well-designed sort order is sufficient
- Reduces configuration complexity
- Maintains consistency across installations

---

### Strict Mode for Sensitive Values

**Decision:** Do not fail commands when sensitive values are present; mask them as designed.

**Rationale:**
- Masking is the appropriate security measure
- Failing would disrupt workflows unnecessarily
- Consistent with Terraform's approach of masking rather than failing

## Final Design Principles

Based on all decisions above, the implementation will follow these principles:

1. **Simplicity First**: Avoid overcomplication in all aspects
2. **Terraform Compatibility**: Match Terraform's behavior for sensitive data
3. **Safe Defaults**: Hide no-ops, mask sensitive values, sort by risk
4. **No Breaking Changes**: Fix sorting as a bug fix, maintain compatibility
5. **Minimal Configuration**: One new flag, no complex options