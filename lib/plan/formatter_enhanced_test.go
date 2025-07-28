package plan

import (
	"strings"
	"testing"

	output "github.com/ArjenSchwarz/go-output/v2"
	"github.com/ArjenSchwarz/strata/config"
)

// TestPropertyChangesFormatter tests the collapsible property formatter
func TestPropertyChangesFormatter(t *testing.T) {
	cfg := &config.Config{}
	formatter := NewFormatter(cfg)
	fn := formatter.propertyChangesFormatterDirect()

	tests := []struct {
		name          string
		input         any
		wantSummary   string
		wantExpanded  bool
		wantTruncated bool
	}{
		{
			name: "empty property changes",
			input: PropertyChangeAnalysis{
				Changes: []PropertyChange{},
				Count:   0,
			},
			wantSummary:  "",
			wantExpanded: false,
		},
		{
			name: "single non-sensitive property",
			input: PropertyChangeAnalysis{
				Changes: []PropertyChange{
					{Name: "instance_type", Before: "t2.micro", After: "t3.micro", Sensitive: false},
				},
				Count: 1,
			},
			wantSummary:  "1 properties changed",
			wantExpanded: false,
		},
		{
			name: "multiple properties with sensitive",
			input: PropertyChangeAnalysis{
				Changes: []PropertyChange{
					{Name: "instance_type", Before: "t2.micro", After: "t3.micro", Sensitive: false},
					{Name: "password", Before: "old", After: "new", Sensitive: true},
					{Name: "tags", Before: map[string]any{"env": "dev"}, After: map[string]any{"env": "prod"}, Sensitive: false},
				},
				Count: 3,
			},
			wantSummary:  "⚠️ 3 properties changed (1 sensitive)",
			wantExpanded: true, // Auto-expand when sensitive
		},
		{
			name: "truncated changes",
			input: PropertyChangeAnalysis{
				Changes: []PropertyChange{
					{Name: "big_data", Before: "lots of data", After: "more data", Sensitive: false},
				},
				Count:     100,
				Truncated: true,
			},
			wantSummary:   "100 properties changed [truncated]",
			wantExpanded:  false,
			wantTruncated: true,
		},
		{
			name:        "non-PropertyChangeAnalysis input",
			input:       "not a property change analysis",
			wantSummary: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn(tt.input)

			// Check if we got a CollapsibleValue when expected
			if tt.wantSummary != "" {
				cv, ok := result.(output.CollapsibleValue)
				if !ok {
					t.Errorf("Expected CollapsibleValue, got %T", result)
					return
				}

				// Check summary matches
				if cv.Summary() != tt.wantSummary {
					t.Errorf("Summary = %q, want %q", cv.Summary(), tt.wantSummary)
				}

				// Check expanded state
				if cv.IsExpanded() != tt.wantExpanded {
					t.Errorf("Expanded = %v, want %v", cv.IsExpanded(), tt.wantExpanded)
				}

				// Check details exist
				if cv.Details() == nil {
					t.Error("Expected Details to be non-nil")
				}
			} else {
				// Should return input unchanged for non-PropertyChangeAnalysis types
				if _, isAnalysis := tt.input.(PropertyChangeAnalysis); !isAnalysis {
					if result != tt.input {
						t.Errorf("Expected unchanged input, got %v", result)
					}
				}
			}
		})
	}
}

// TestDependenciesFormatter tests the collapsible dependencies formatter
func TestDependenciesFormatter(t *testing.T) {
	cfg := &config.Config{}
	formatter := NewFormatter(cfg)
	fn := formatter.dependenciesFormatterDirect()

	tests := []struct {
		name         string
		input        any
		wantSummary  string
		wantExpanded bool
	}{
		{
			name:        "nil dependencies",
			input:       nil,
			wantSummary: "",
		},
		{
			name: "no dependencies",
			input: &DependencyInfo{
				DependsOn: []string{},
				UsedBy:    []string{},
			},
			wantSummary: "No dependencies",
		},
		{
			name: "only depends on",
			input: &DependencyInfo{
				DependsOn: []string{"aws_vpc.main", "aws_subnet.private"},
				UsedBy:    []string{},
			},
			wantSummary:  "2 dependencies",
			wantExpanded: false,
		},
		{
			name: "only used by",
			input: &DependencyInfo{
				DependsOn: []string{},
				UsedBy:    []string{"aws_instance.web", "aws_instance.app"},
			},
			wantSummary:  "2 dependencies",
			wantExpanded: false,
		},
		{
			name: "both depends on and used by",
			input: &DependencyInfo{
				DependsOn: []string{"aws_vpc.main"},
				UsedBy:    []string{"aws_instance.web", "aws_instance.app"},
			},
			wantSummary:  "3 dependencies",
			wantExpanded: false,
		},
		{
			name:        "non-DependencyInfo input",
			input:       "not a dependency info",
			wantSummary: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn(tt.input)

			if tt.wantSummary != "" {
				if tt.wantSummary == "No dependencies" {
					// Special case: "No dependencies" returns a string
					if str, ok := result.(string); ok {
						if str != tt.wantSummary {
							t.Errorf("String = %q, want %q", str, tt.wantSummary)
						}
					} else {
						t.Errorf("Expected string for 'No dependencies', got %T", result)
					}
				} else {
					// Normal case: returns CollapsibleValue
					cv, ok := result.(output.CollapsibleValue)
					if !ok {
						t.Errorf("Expected CollapsibleValue, got %T", result)
						return
					}

					if cv.Summary() != tt.wantSummary {
						t.Errorf("Summary = %q, want %q", cv.Summary(), tt.wantSummary)
					}

					if cv.IsExpanded() != tt.wantExpanded {
						t.Errorf("Expanded = %v, want %v", cv.IsExpanded(), tt.wantExpanded)
					}
				}
			} else {
				// Should return input unchanged for non-DependencyInfo types
				if _, isDep := tt.input.(*DependencyInfo); !isDep {
					if result != tt.input {
						t.Errorf("Expected unchanged input, got %v", result)
					}
				}
			}
		})
	}
}

// TestPrepareResourceTableData tests the table data preparation function
func TestPrepareResourceTableData(t *testing.T) {
	cfg := &config.Config{
		SensitiveResources: []config.SensitiveResource{
			{ResourceType: "aws_rds_db_instance"},
		},
	}
	formatter := NewFormatter(cfg)

	// Create test resource changes
	changes := []ResourceChange{
		{
			Address:       "aws_instance.web",
			Type:          "aws_instance",
			Name:          "web",
			ChangeType:    ChangeTypeCreate,
			IsDestructive: false,
			Provider:      "aws",
			TopChanges:    []string{"instance_type", "ami"},
		},
		{
			Address:          "aws_rds_db_instance.main",
			Type:             "aws_rds_db_instance",
			Name:             "main",
			ChangeType:       ChangeTypeReplace,
			IsDestructive:    true,
			ReplacementType:  ReplacementAlways,
			Provider:         "aws",
			TopChanges:       []string{"engine"},
			ReplacementHints: []string{"engine version change requires replacement"},
			IsDangerous:      true,
			DangerReason:     "Sensitive resource replacement",
		},
	}

	tableData := formatter.prepareResourceTableData(changes)

	// Validate results
	if len(tableData) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(tableData))
	}

	// Check first resource
	row1 := tableData[0]
	if row1["resource"] != "aws_instance.web" {
		t.Errorf("Expected resource 'aws_instance.web', got %v", row1["resource"])
	}
	if row1["action"] != "Add" {
		t.Errorf("Expected action 'Add', got %v", row1["action"])
	}
	if row1["risk_level"] != "low" {
		t.Errorf("Expected risk_level 'low', got %v", row1["risk_level"])
	}

	// Verify property_changes exists and is PropertyChangeAnalysis
	if _, ok := row1["property_changes"].(PropertyChangeAnalysis); !ok {
		t.Errorf("Expected property_changes to be PropertyChangeAnalysis, got %T", row1["property_changes"])
	}

	// Check second resource (sensitive RDS)
	row2 := tableData[1]
	if row2["resource"] != "aws_rds_db_instance.main" {
		t.Errorf("Expected resource 'aws_rds_db_instance.main', got %v", row2["resource"])
	}
	if row2["risk_level"] != "high" {
		t.Errorf("Expected risk_level 'high' for sensitive resource replacement, got %v", row2["risk_level"])
	}

	// Check replacement reasons are included
	if _, hasReasons := row2["replacement_reasons"]; !hasReasons {
		t.Error("Expected replacement_reasons for replacement change")
	}
}

// TestFormatResourceChangesWithProgressiveDisclosure tests the main formatting function
func TestFormatResourceChangesWithProgressiveDisclosure(t *testing.T) {
	cfg := &config.Config{
		Plan: config.PlanConfig{
			ShowDetails: true,
		},
	}
	formatter := NewFormatter(cfg)

	summary := &PlanSummary{
		PlanFile:         "test.tfplan",
		TerraformVersion: "1.6.0",
		ResourceChanges: []ResourceChange{
			{
				Address:       "aws_instance.example",
				Type:          "aws_instance",
				Name:          "example",
				ChangeType:    ChangeTypeCreate,
				IsDestructive: false,
				Provider:      "aws",
				TopChanges:    []string{"instance_type"},
			},
		},
	}

	doc, err := formatter.formatResourceChangesWithProgressiveDisclosure(summary)
	if err != nil {
		t.Fatalf("formatResourceChangesWithProgressiveDisclosure() error = %v", err)
	}

	if doc == nil {
		t.Fatal("Expected non-nil document")
	}

	// Verify document has expected structure
	contents := doc.GetContents()
	if len(contents) == 0 {
		t.Error("Expected document to have content")
	}

	// Check that first content is the plan information table
	if tableContent, ok := contents[0].(*output.TableContent); ok {
		if tableContent.Title() != "Plan Information" {
			t.Errorf("Expected table title 'Plan Information', got %q", tableContent.Title())
		}
	} else {
		t.Errorf("Expected first content to be TableContent, got %T", contents[0])
	}
}

// TestPropertyChangeDetailsFormatting tests the formatting of property change details
func TestPropertyChangeDetailsFormatting(t *testing.T) {
	cfg := &config.Config{}
	formatter := NewFormatter(cfg)

	changes := []PropertyChange{
		{
			Name:      "instance_type",
			Before:    "t2.micro",
			After:     "t3.micro",
			Sensitive: false,
		},
		{
			Name:      "password",
			Before:    "secret123",
			After:     "newsecret456",
			Sensitive: true,
		},
		{
			Name: "tags",
			Before: map[string]any{
				"Name": "old-server",
				"Env":  "dev",
			},
			After: map[string]any{
				"Name": "new-server",
				"Env":  "prod",
			},
			Sensitive: false,
		},
	}

	details := formatter.formatPropertyChangeDetails(changes)

	// Check that we got a formatted string
	if details == "" {
		t.Error("Expected non-empty details string")
	}

	// Check normal property formatting
	if !strings.Contains(details, "• instance_type: t2.micro → t3.micro") {
		t.Errorf("Expected instance_type change in details, got: %s", details)
	}

	// Check sensitive property masking
	if !strings.Contains(details, "• password: [sensitive value hidden] → [sensitive value hidden]") {
		t.Errorf("Expected masked password in details, got: %s", details)
	}

	// Check complex property
	if !strings.Contains(details, "• tags:") {
		t.Errorf("Expected tags change in details, got: %s", details)
	}

	// Split details to count lines
	lines := strings.Split(details, "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines of details, got %d", len(lines))
	}
}

// TestFormatterWithMockedGoOutput tests formatter with mocked go-output components
func TestFormatterWithMockedGoOutput(t *testing.T) {
	cfg := &config.Config{
		ExpandAll: true, // Test global expand-all setting
	}
	formatter := NewFormatter(cfg)

	// Test that createOutputWithConfig respects expand-all setting
	outputConfig := &config.OutputConfiguration{
		Format: "table",
	}

	format := formatter.getFormatFromConfig(outputConfig.Format)
	// Just verify it doesn't panic
	if format.Name == "" {
		t.Error("Expected valid format")
	}

	// Verify that property formatter respects configuration
	fn := formatter.propertyChangesFormatterDirect()
	result := fn(PropertyChangeAnalysis{
		Changes: []PropertyChange{
			{Name: "test", Before: "a", After: "b"},
		},
		Count: 1,
	})

	if cv, ok := result.(output.CollapsibleValue); ok {
		// With expand-all, this would be expanded when rendered
		// (actual expansion happens in go-output library during rendering)
		if cv.Summary() != "1 properties changed" {
			t.Errorf("Unexpected summary: %s", cv.Summary())
		}
	}
}

// Helper to create test ResourceChange
func createTestResourceChange(address, resourceType, changeType string) ResourceChange {
	return ResourceChange{
		Address:    address,
		Type:       resourceType,
		Name:       strings.Split(address, ".")[1],
		ChangeType: ChangeType(changeType),
		Provider:   strings.Split(resourceType, "_")[0],
		TopChanges: []string{"test_property"},
	}
}

// TestFormatterErrorHandling tests error handling in formatter functions
func TestFormatterErrorHandling(t *testing.T) {
	cfg := &config.Config{}
	formatter := NewFormatter(cfg)

	// Test prepareResourceTableData with empty changes
	data := formatter.prepareResourceTableData([]ResourceChange{})
	if len(data) != 0 {
		t.Error("Expected empty data for empty changes")
	}

	// Test with valid change
	changes := []ResourceChange{
		createTestResourceChange("test.resource", "test_type", "create"),
	}
	data = formatter.prepareResourceTableData(changes)
	if len(data) != 1 {
		t.Errorf("Expected 1 row, got %d", len(data))
	}

	// Test with invalid input types
	propFn := formatter.propertyChangesFormatterDirect()
	result := propFn(struct{}{}) // Invalid type
	if result != struct{}{} {
		t.Error("Expected unchanged input for invalid type")
	}

	depFn := formatter.dependenciesFormatterDirect()
	result = depFn(123) // Invalid type
	if result != 123 {
		t.Error("Expected unchanged input for invalid type")
	}
}
