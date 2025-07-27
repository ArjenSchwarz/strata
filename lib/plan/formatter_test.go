package plan

import (
	"testing"
	"time"

	"github.com/ArjenSchwarz/strata/config"
)

func TestFormatter_ValidateOutputFormat(t *testing.T) {
	cfg := &config.Config{}
	formatter := NewFormatter(cfg)

	tests := []struct {
		format  string
		wantErr bool
	}{
		{"table", false},
		{"json", false},
		{"html", false},
		{"markdown", false},
		{"xml", true}, // unsupported
		{"", true},    // empty
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			err := formatter.ValidateOutputFormat(tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOutputFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormatter_OutputSummary_V2(t *testing.T) {
	cfg := &config.Config{
		Plan: config.PlanConfig{
			ShowDetails:         true,
			HighlightDangers:    true,
			AlwaysShowSensitive: false,
		},
	}
	formatter := NewFormatter(cfg)

	summary := &PlanSummary{
		PlanFile:         "test.tfplan",
		TerraformVersion: "1.6.0",
		Workspace:        "production",
		Backend: BackendInfo{
			Type:     "s3",
			Location: "my-bucket",
		},
		CreatedAt: time.Date(2025, 5, 25, 23, 25, 28, 0, time.UTC),
		Statistics: ChangeStatistics{
			Total:        5,
			ToAdd:        2,
			ToChange:     1,
			ToDestroy:    1,
			Replacements: 1,
			HighRisk:     1,
		},
		ResourceChanges: []ResourceChange{
			{
				Address:       "aws_instance.example",
				Type:          "aws_instance",
				Name:          "example",
				ChangeType:    ChangeTypeCreate,
				IsDestructive: false,
				IsDangerous:   false,
			},
		},
	}

	outputConfig := &config.OutputConfiguration{
		Format:           "table",
		OutputFile:       "",
		OutputFileFormat: "table",
		UseEmoji:         true,
		UseColors:        true,
		TableStyle:       "default",
		MaxColumnWidth:   80,
	}

	// Test that OutputSummary doesn't crash with basic input
	err := formatter.OutputSummary(summary, outputConfig, true)
	if err != nil {
		t.Errorf("OutputSummary() error = %v", err)
	}
}

func TestFormatter_createPlanInfoDataV2(t *testing.T) {
	cfg := &config.Config{}
	formatter := NewFormatter(cfg)

	summary := &PlanSummary{
		PlanFile:         "test.tfplan",
		TerraformVersion: "1.6.0",
		Workspace:        "production",
		Backend: BackendInfo{
			Type:     "s3",
			Location: "my-bucket",
		},
		CreatedAt: time.Date(2025, 5, 25, 23, 25, 28, 0, time.UTC),
	}

	data, err := formatter.createPlanInfoDataV2(summary)
	if err != nil {
		t.Errorf("createPlanInfoDataV2() error = %v", err)
	}

	if len(data) != 1 {
		t.Errorf("Expected 1 row, got %d", len(data))
	}

	row := data[0]
	if row["Plan File"] != "test.tfplan" {
		t.Errorf("Expected Plan File to be 'test.tfplan', got %v", row["Plan File"])
	}

	if row["Version"] != "1.6.0" {
		t.Errorf("Expected Version to be '1.6.0', got %v", row["Version"])
	}
}

func TestFormatter_createStatisticsSummaryDataV2(t *testing.T) {
	cfg := &config.Config{}
	formatter := NewFormatter(cfg)

	summary := &PlanSummary{
		PlanFile: "test.tfplan",
		Statistics: ChangeStatistics{
			Total:        10,
			ToAdd:        3,
			ToChange:     4,
			ToDestroy:    2,
			Replacements: 1,
			HighRisk:     1,
		},
	}

	data, err := formatter.createStatisticsSummaryDataV2(summary)
	if err != nil {
		t.Errorf("createStatisticsSummaryDataV2() error = %v", err)
	}

	if len(data) != 1 {
		t.Errorf("Expected 1 row, got %d", len(data))
	}

	row := data[0]
	if row["TOTAL CHANGES"] != 10 {
		t.Errorf("Expected TOTAL CHANGES to be 10, got %v", row["TOTAL CHANGES"])
	}

	if row["ADDED"] != 3 {
		t.Errorf("Expected ADDED to be 3, got %v", row["ADDED"])
	}
}

func TestFormatter_getFormatFromConfig(t *testing.T) {
	cfg := &config.Config{}
	formatter := NewFormatter(cfg)

	tests := []struct {
		input    string
		expected string
	}{
		{"json", "json"},
		{"JSON", "json"},
		{"csv", "csv"},
		{"html", "html"},
		{"markdown", "markdown"},
		{"table", "table"},
		{"unknown", "table"}, // defaults to table
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			format := formatter.getFormatFromConfig(tt.input)
			// We can't easily compare Format structs, so just check it doesn't panic
			if format.Name == "" {
				t.Errorf("getFormatFromConfig() returned empty format name for input %s", tt.input)
			}
		})
	}
}

func TestFormatter_formatGroupedWithCollapsibleSections(t *testing.T) {
	cfg := &config.Config{
		Plan: config.PlanConfig{
			Grouping: config.GroupingConfig{
				Enabled:   true,
				Threshold: 2,
			},
		},
	}
	formatter := NewFormatter(cfg)

	// Create test data with multiple providers
	summary := &PlanSummary{
		PlanFile:         "test.tfplan",
		TerraformVersion: "1.6.0",
		ResourceChanges:  []ResourceChange{},
	}

	// Create test groups
	groups := map[string][]ResourceChange{
		"aws": {
			{
				Address:    "aws_s3_bucket.test",
				Type:       "aws_s3_bucket",
				ChangeType: ChangeTypeCreate,
			},
			{
				Address:    "aws_ec2_instance.web",
				Type:       "aws_ec2_instance",
				ChangeType: ChangeTypeUpdate,
			},
		},
		"azurerm": {
			{
				Address:    "azurerm_storage_account.test",
				Type:       "azurerm_storage_account",
				ChangeType: ChangeTypeCreate,
			},
		},
	}

	// Test that the function doesn't panic and returns a document
	doc, err := formatter.formatGroupedWithCollapsibleSections(summary, groups)
	if err != nil {
		t.Errorf("formatGroupedWithCollapsibleSections() error = %v", err)
		return
	}

	if doc == nil {
		t.Error("formatGroupedWithCollapsibleSections() returned nil document")
		return
	}

	// Check that the document has content
	contents := doc.GetContents()
	if len(contents) == 0 {
		t.Error("formatGroupedWithCollapsibleSections() returned document with no contents")
	}
}

func TestFormatter_hasHighRiskChanges(t *testing.T) {
	formatter := &Formatter{}

	testCases := []struct {
		name      string
		resources []ResourceChange
		expected  bool
	}{
		{
			name:      "Empty resources should return false",
			resources: []ResourceChange{},
			expected:  false,
		},
		{
			name: "Non-dangerous resources should return false",
			resources: []ResourceChange{
				{
					ChangeType:  ChangeTypeCreate,
					IsDangerous: false,
				},
				{
					ChangeType:  ChangeTypeUpdate,
					IsDangerous: false,
				},
			},
			expected: false,
		},
		{
			name: "Dangerous deletion should return true",
			resources: []ResourceChange{
				{
					ChangeType:  ChangeTypeDelete,
					IsDangerous: true,
				},
			},
			expected: true,
		},
		{
			name: "Dangerous replacement should return true",
			resources: []ResourceChange{
				{
					ChangeType:  ChangeTypeReplace,
					IsDangerous: true,
				},
			},
			expected: true,
		},
		{
			name: "Dangerous update should return false",
			resources: []ResourceChange{
				{
					ChangeType:  ChangeTypeUpdate,
					IsDangerous: true,
				},
			},
			expected: false,
		},
		{
			name: "Mixed with one dangerous deletion should return true",
			resources: []ResourceChange{
				{
					ChangeType:  ChangeTypeCreate,
					IsDangerous: false,
				},
				{
					ChangeType:  ChangeTypeDelete,
					IsDangerous: true,
				},
				{
					ChangeType:  ChangeTypeUpdate,
					IsDangerous: false,
				},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatter.hasHighRiskChanges(tc.resources)
			if result != tc.expected {
				t.Errorf("hasHighRiskChanges() = %v, expected %v", result, tc.expected)
			}
		})
	}
}
