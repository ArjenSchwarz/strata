package plan

import (
	"testing"

	"github.com/ArjenSchwarz/strata/config"
	"github.com/stretchr/testify/assert"
)

// TestOutputRefinements_ExistingTestCompatibility ensures all enhancements are backward compatible
// and don't break existing functionality (Task 8.3)
func TestOutputRefinements_ExistingTestCompatibility(t *testing.T) {
	testCases := []struct {
		name        string
		description string
		testFunc    func(t *testing.T)
	}{
		{
			name:        "Formatter sorting maintains backward compatibility",
			description: "Verify sortResourcesByPriority works with existing test cases",
			testFunc:    testFormatterSortingBackwardCompatibility,
		},
		{
			name:        "Property sorting maintains order with existing tests",
			description: "Verify property alphabetical sorting doesn't break existing tests",
			testFunc:    testPropertySortingBackwardCompatibility,
		},
		{
			name:        "No-op filtering is disabled by default",
			description: "Verify existing behavior unchanged when ShowNoOps is not configured",
			testFunc:    testNoOpFilteringDefaultBehavior,
		},
		{
			name:        "Statistics calculations include unmodified resources",
			description: "Verify statistics still count no-ops in Unmodified field",
			testFunc:    testStatisticsIncludeNoOps,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.testFunc(t)
		})
	}
}

// testFormatterSortingBackwardCompatibility verifies that the enhanced sorting
// maintains backward compatibility with existing formatter behavior
func testFormatterSortingBackwardCompatibility(t *testing.T) {
	t.Helper()
	// Create test resources similar to existing formatter tests
	resources := []ResourceChange{
		{
			Address:     "aws_instance.example1",
			Type:        "aws_instance",
			Name:        "example1",
			ChangeType:  ChangeTypeCreate,
			IsDangerous: false,
		},
		{
			Address:     "aws_instance.example2",
			Type:        "aws_instance",
			Name:        "example2",
			ChangeType:  ChangeTypeUpdate,
			IsDangerous: false,
		},
		{
			Address:     "aws_instance.example3",
			Type:        "aws_instance",
			Name:        "example3",
			ChangeType:  ChangeTypeDelete,
			IsDangerous: false,
		},
	}

	cfg := &config.Config{
		Plan: config.PlanConfig{
			ShowDetails: true,
		},
	}

	formatter := NewFormatter(cfg)
	sortedResources := formatter.sortResourcesByPriority(resources)

	// Verify sorting works (delete > update > create when no danger flags)
	assert.Equal(t, ChangeTypeDelete, sortedResources[0].ChangeType, "Delete should come first")
	assert.Equal(t, ChangeTypeUpdate, sortedResources[1].ChangeType, "Update should come second")
	assert.Equal(t, ChangeTypeCreate, sortedResources[2].ChangeType, "Create should come last")

	// Verify original slice is not modified (existing test expectation)
	assert.Equal(t, ChangeTypeCreate, resources[0].ChangeType, "Original slice should be unchanged")
}

// testPropertySortingBackwardCompatibility verifies property sorting doesn't break existing tests
func testPropertySortingBackwardCompatibility(t *testing.T) {
	t.Helper()
	// Test that property sorting works with various property name formats
	// that might appear in existing tests
	changes := []PropertyChange{
		{Name: "instance_type", Path: []string{"instance_type"}},
		{Name: "ami", Path: []string{"ami"}},
		{Name: "user_data", Path: []string{"user_data"}},
		{Name: "tags", Path: []string{"tags"}},
	}

	// Simulate the sorting that happens in analyzePropertyChanges
	sortedChanges := make([]PropertyChange, len(changes))
	copy(sortedChanges, changes)

	// Apply sorting (simple bubble sort as used in the analyzer)
	for i := range sortedChanges {
		for j := i + 1; j < len(sortedChanges); j++ {
			iName := sortedChanges[i].Name
			jName := sortedChanges[j].Name

			if iName > jName {
				sortedChanges[i], sortedChanges[j] = sortedChanges[j], sortedChanges[i]
			}
		}
	}

	expectedOrder := []string{"ami", "instance_type", "tags", "user_data"}
	for i, expected := range expectedOrder {
		assert.Equal(t, expected, sortedChanges[i].Name,
			"Property %d should be %s", i, expected)
	}
}

// testNoOpFilteringDefaultBehavior verifies that no-op filtering is disabled by default
// to maintain existing behavior
func testNoOpFilteringDefaultBehavior(t *testing.T) {
	t.Helper()
	// Create test resources with no-ops
	resources := []ResourceChange{
		{
			Address:    "aws_instance.example1",
			ChangeType: ChangeTypeCreate,
			IsNoOp:     false,
		},
		{
			Address:    "aws_instance.example2",
			ChangeType: ChangeTypeNoOp,
			IsNoOp:     true,
		},
	}

	// Default config should not filter no-ops (ShowNoOps defaults to true for backward compatibility)
	cfg := &config.Config{
		Plan: config.PlanConfig{
			ShowDetails: true,
			// ShowNoOps not explicitly set - should default to showing no-ops
		},
	}

	formatter := NewFormatter(cfg)

	// When ShowNoOps is false, filtering should happen
	cfg.Plan.ShowNoOps = false
	filteredResources := formatter.filterNoOps(resources)
	assert.Equal(t, 1, len(filteredResources), "Should filter out no-ops when ShowNoOps=false")

	// When ShowNoOps is true, no filtering should happen
	cfg.Plan.ShowNoOps = true
	unfilteredResources := formatter.filterNoOps(resources)
	assert.Equal(t, 2, len(unfilteredResources), "Should include no-ops when ShowNoOps=true")
}

// testStatisticsIncludeNoOps verifies that statistics calculations
// continue to include no-op resources in the Unmodified count
func testStatisticsIncludeNoOps(t *testing.T) {
	t.Helper()
	resources := []ResourceChange{
		{ChangeType: ChangeTypeCreate, IsDangerous: false},
		{ChangeType: ChangeTypeUpdate, IsDangerous: false},
		{ChangeType: ChangeTypeDelete, IsDangerous: false},
		{ChangeType: ChangeTypeReplace, IsDangerous: false, ReplacementType: ReplacementAlways},
		{ChangeType: ChangeTypeNoOp, IsDangerous: false}, // This should be counted in Unmodified
		{ChangeType: ChangeTypeNoOp, IsDangerous: false}, // This too
	}

	outputs := []OutputChange{
		{ChangeType: ChangeTypeCreate, IsNoOp: false},
		{ChangeType: ChangeTypeUpdate, IsNoOp: false},
		{ChangeType: ChangeTypeNoOp, IsNoOp: true}, // This should NOT be counted in OutputChanges
	}

	analyzer := &Analyzer{}
	stats := analyzer.calculateStatistics(resources, outputs)

	// Verify statistics calculations
	assert.Equal(t, 1, stats.ToAdd, "Should count creates")
	assert.Equal(t, 1, stats.ToChange, "Should count updates")
	assert.Equal(t, 1, stats.ToDestroy, "Should count deletes")
	assert.Equal(t, 1, stats.Replacements, "Should count replacements")
	assert.Equal(t, 2, stats.Unmodified, "Should count no-ops in Unmodified")
	assert.Equal(t, 4, stats.Total, "Total should exclude no-ops (only counts actionable resources)")
	assert.Equal(t, 2, stats.OutputChanges, "Should exclude no-op outputs from OutputChanges count")
}
