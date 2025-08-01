package plan

import (
	"strings"
	"testing"

	"github.com/ArjenSchwarz/strata/config"
)

func TestFormatterCreateActionDisplay(t *testing.T) {
	cfg := &config.Config{
		Plan: config.PlanConfig{
			ExpandableSections: config.ExpandableSectionsConfig{
				Enabled:             true,
				AutoExpandDangerous: false,
			},
		},
		ExpandAll: true, // Expand all to see the details
	}
	formatter := NewFormatter(cfg)

	// Create a test summary with a create action
	summary := &PlanSummary{
		ResourceChanges: []ResourceChange{
			{
				Address:    "aws_iam_role.web_profile",
				Type:       "aws_iam_role",
				ChangeType: ChangeTypeCreate,
				PropertyChanges: PropertyChangeAnalysis{
					Changes: []PropertyChange{
						{
							Name:   "name",
							Action: "add",
							After:  "web-server-profile",
						},
						{
							Name:   "path",
							Action: "add",
							After:  "/",
						},
						{
							Name:   "tags",
							Action: "add",
							After: map[string]any{
								"Environment": "production",
								"ManagedBy":   "terraform",
							},
						},
					},
					Count: 3,
				},
			},
		},
	}

	// Format the property changes
	tableData := formatter.prepareResourceTableData(summary.ResourceChanges)
	if len(tableData) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(tableData))
	}

	// Get the property changes formatter
	propFormatter := formatter.propertyChangesFormatterTerraform()

	// Get the property changes from the table data
	propChanges := tableData[0]["property_changes"]

	// Apply the formatter
	result := propFormatter(propChanges)

	// Convert to string to check the content
	resultStr := ""
	if collapsible, ok := result.(interface{ Details() any }); ok {
		details := collapsible.Details()
		if str, ok := details.(string); ok {
			resultStr = str
		}
	}

	// Verify the formatting
	// For create actions, we should see the new values with + prefix
	expectedPatterns := []string{
		"+ name = \"web-server-profile\"",
		"+ path = \"/\"",
		"+ tags = { Environment = \"production\", ManagedBy = \"terraform\" }",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(resultStr, pattern) {
			t.Errorf("Expected to find pattern %q in output, but got:\n%s", pattern, resultStr)
		}
	}

	// Ensure we don't have comparison arrows
	if strings.Contains(resultStr, "->") {
		t.Errorf("Create action should not contain comparison arrows (->), but got:\n%s", resultStr)
	}

	// Ensure we have the + prefix for new resources (Terraform diff-style)
	if !strings.Contains(resultStr, "  +") {
		t.Errorf("Create action should contain diff-style + prefix, but got:\n%s", resultStr)
	}
}

func TestFormatterUpdateActionDisplay(t *testing.T) {
	cfg := &config.Config{
		Plan: config.PlanConfig{
			ExpandableSections: config.ExpandableSectionsConfig{
				Enabled:             true,
				AutoExpandDangerous: false,
			},
		},
		ExpandAll: true,
	}
	formatter := NewFormatter(cfg)

	// Create a test summary with an update action
	summary := &PlanSummary{
		ResourceChanges: []ResourceChange{
			{
				Address:    "aws_instance.web",
				Type:       "aws_instance",
				ChangeType: ChangeTypeUpdate,
				PropertyChanges: PropertyChangeAnalysis{
					Changes: []PropertyChange{
						{
							Name:   "instance_type",
							Action: "update",
							Before: "t2.micro",
							After:  "t2.small",
						},
					},
					Count: 1,
				},
			},
		},
	}

	// Format the property changes
	tableData := formatter.prepareResourceTableData(summary.ResourceChanges)
	propFormatter := formatter.propertyChangesFormatterTerraform()
	propChanges := tableData[0]["property_changes"]
	result := propFormatter(propChanges)

	// Convert to string to check the content
	resultStr := ""
	if collapsible, ok := result.(interface{ Details() any }); ok {
		details := collapsible.Details()
		if str, ok := details.(string); ok {
			resultStr = str
		}
	}

	// For update actions, we should see the comparison format
	expectedPattern := "~ instance_type = \"t2.micro\" -> \"t2.small\""
	if !strings.Contains(resultStr, expectedPattern) {
		t.Errorf("Expected to find pattern %q in output for update action, but got:\n%s", expectedPattern, resultStr)
	}
}
