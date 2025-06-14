package plan

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ArjenSchwarz/strata/config"
)

func TestFormatter_formatPlanInfo(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

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
		IsDryRun:  false,
	}

	err := formatter.formatPlanInfo(summary, "table")
	if err != nil {
		t.Errorf("formatPlanInfo() error = %v", err)
	}

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that key information is present
	expectedStrings := []string{
		"Plan Information",
		"test.tfplan",
		"1.6.0",
		"production",
		"s3 (my-bucket)",
		"2025-05-25 23:25:28",
		"No",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("formatPlanInfo() output missing expected string: %s", expected)
		}
	}
}

func TestFormatter_formatPlanInfo_DryRun(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := &config.Config{}
	formatter := NewFormatter(cfg)

	summary := &PlanSummary{
		PlanFile:         "test.tfplan",
		TerraformVersion: "1.6.0",
		Workspace:        "default",
		Backend: BackendInfo{
			Type:     "local",
			Location: "terraform.tfstate",
		},
		CreatedAt: time.Now(),
		IsDryRun:  true,
	}

	err := formatter.formatPlanInfo(summary, "table")
	if err != nil {
		t.Errorf("formatPlanInfo() error = %v", err)
	}

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that dry run is shown as "Yes"
	if !strings.Contains(output, "Yes") {
		t.Error("formatPlanInfo() should show 'Yes' for dry run")
	}
}

func TestFormatter_formatStatisticsSummary(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := &config.Config{}
	formatter := NewFormatter(cfg)

	summary := &PlanSummary{
		PlanFile: "test.tfplan",
		Statistics: ChangeStatistics{
			ToAdd:        2,
			ToChange:     1,
			ToDestroy:    1,
			Replacements: 1,
			Conditionals: 0,
			Total:        5,
		},
	}

	err := formatter.formatStatisticsSummary(summary, "table")
	if err != nil {
		t.Errorf("formatStatisticsSummary() error = %v", err)
	}

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that key information is present
	expectedStrings := []string{
		"Summary for test.tfplan",
		"TOTAL",
		"ADDED",
		"REMOVED",
		"MODIFIED",
		"REPLACEMENTS",
		"CONDITIONALS",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("formatStatisticsSummary() output missing expected string: %s", expected)
		}
	}
}

func TestFormatter_OutputSummary(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := &config.Config{}
	formatter := NewFormatter(cfg)

	summary := &PlanSummary{
		PlanFile:         "test.tfplan",
		TerraformVersion: "1.6.0",
		Workspace:        "default",
		Backend: BackendInfo{
			Type:     "local",
			Location: "terraform.tfstate",
		},
		CreatedAt: time.Now(),
		IsDryRun:  false,
		ResourceChanges: []ResourceChange{
			{
				Address:       "aws_instance.web",
				Type:          "aws_instance",
				Name:          "web",
				ChangeType:    ChangeTypeCreate,
				IsDestructive: false,
			},
		},
		Statistics: ChangeStatistics{
			ToAdd:        1,
			ToChange:     0,
			ToDestroy:    0,
			Replacements: 0,
			Conditionals: 0,
			Total:        1,
		},
	}

	err := formatter.OutputSummary(summary, "table", false)
	if err != nil {
		t.Errorf("OutputSummary() error = %v", err)
	}

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that both plan info and summary are present
	expectedStrings := []string{
		"Plan Information",
		"test.tfplan",
		"Summary for test.tfplan",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("OutputSummary() output missing expected string: %s", expected)
		}
	}
}

func TestFormatter_OutputSummary_WithDetails(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := &config.Config{}
	formatter := NewFormatter(cfg)

	summary := &PlanSummary{
		PlanFile:         "test.tfplan",
		TerraformVersion: "1.6.0",
		Workspace:        "default",
		Backend: BackendInfo{
			Type:     "local",
			Location: "terraform.tfstate",
		},
		CreatedAt: time.Now(),
		IsDryRun:  false,
		ResourceChanges: []ResourceChange{
			{
				Address:       "aws_instance.web",
				Type:          "aws_instance",
				Name:          "web",
				ChangeType:    ChangeTypeCreate,
				IsDestructive: false,
			},
			{
				Address:       "aws_instance.old",
				Type:          "aws_instance",
				Name:          "old",
				ChangeType:    ChangeTypeDelete,
				IsDestructive: true,
			},
		},
		Statistics: ChangeStatistics{
			ToAdd:        1,
			ToChange:     0,
			ToDestroy:    1,
			Replacements: 0,
			Conditionals: 0,
			Total:        2,
		},
	}

	err := formatter.OutputSummary(summary, "table", true)
	if err != nil {
		t.Errorf("OutputSummary() error = %v", err)
	}

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that resource details are shown
	expectedStrings := []string{
		"aws_instance.web",
		"aws_instance.old",
		"Resource Changes",
		"ACTION",
		"RESOURCE",
		"TYPE",
		"ID",
		"REPLACEMENT",
		"MODULE",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("OutputSummary() with details missing expected string: %s", expected)
		}
	}
}

func TestGetChangeIcon(t *testing.T) {
	tests := []struct {
		changeType ChangeType
		expected   string
	}{
		{ChangeTypeCreate, "+"},
		{ChangeTypeUpdate, "~"},
		{ChangeTypeDelete, "-"},
		{ChangeTypeReplace, "±"},
		{ChangeTypeNoOp, " "},
	}

	for _, tt := range tests {
		t.Run(string(tt.changeType), func(t *testing.T) {
			result := getChangeIcon(tt.changeType)
			if result != tt.expected {
				t.Errorf("getChangeIcon(%v) = %v, want %v", tt.changeType, result, tt.expected)
			}
		})
	}
}

func TestNewFormatter(t *testing.T) {
	cfg := &config.Config{}
	formatter := NewFormatter(cfg)

	if formatter.config != cfg {
		t.Error("NewFormatter() should set config correctly")
	}
}

func TestFormatter_StatisticsSummary_VariousChangeCombinations(t *testing.T) {
	testCases := []struct {
		name        string
		statistics  ChangeStatistics
		planFile    string
		shouldError bool
	}{
		{
			name: "All change types present",
			statistics: ChangeStatistics{
				ToAdd:        5,
				ToChange:     3,
				ToDestroy:    2,
				Replacements: 1,
				Conditionals: 1,
				Total:        12,
			},
			planFile:    "complex.tfplan",
			shouldError: false,
		},
		{
			name: "Only creates",
			statistics: ChangeStatistics{
				ToAdd:        10,
				ToChange:     0,
				ToDestroy:    0,
				Replacements: 0,
				Conditionals: 0,
				Total:        10,
			},
			planFile:    "create-only.tfplan",
			shouldError: false,
		},
		{
			name: "Only replacements",
			statistics: ChangeStatistics{
				ToAdd:        0,
				ToChange:     0,
				ToDestroy:    0,
				Replacements: 3,
				Conditionals: 2,
				Total:        5,
			},
			planFile:    "replace-only.tfplan",
			shouldError: false,
		},
		{
			name: "No changes",
			statistics: ChangeStatistics{
				ToAdd:        0,
				ToChange:     0,
				ToDestroy:    0,
				Replacements: 0,
				Conditionals: 0,
				Total:        0,
			},
			planFile:    "no-changes.tfplan",
			shouldError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			cfg := &config.Config{}
			formatter := NewFormatter(cfg)

			summary := &PlanSummary{
				PlanFile:   tc.planFile,
				Statistics: tc.statistics,
			}

			err := formatter.formatStatisticsSummary(summary, "table")

			// Restore stdout and read output
			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if tc.shouldError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify plan file name appears in output
			if !strings.Contains(output, tc.planFile) {
				t.Errorf("Output should contain plan file name %s", tc.planFile)
			}

			// Verify horizontal format headers are present
			expectedHeaders := []string{"TOTAL", "ADDED", "REMOVED", "MODIFIED", "REPLACEMENTS", "CONDITIONALS"}
			for _, header := range expectedHeaders {
				if !strings.Contains(output, header) {
					t.Errorf("Output should contain header %s", header)
				}
			}
		})
	}
}

func TestFormatter_OutputFormat_Compatibility(t *testing.T) {
	testFormats := []string{"table", "json"}

	for _, format := range testFormats {
		t.Run(format, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			cfg := &config.Config{}
			formatter := NewFormatter(cfg)

			summary := &PlanSummary{
				PlanFile:         "test.tfplan",
				TerraformVersion: "1.6.0",
				Workspace:        "default",
				Backend: BackendInfo{
					Type:     "local",
					Location: "terraform.tfstate",
				},
				CreatedAt: time.Now(),
				IsDryRun:  false,
				ResourceChanges: []ResourceChange{
					{
						Address:         "aws_instance.web",
						Type:            "aws_instance",
						Name:            "web",
						ChangeType:      ChangeTypeReplace,
						IsDestructive:   true,
						ReplacementType: ReplacementAlways,
						PhysicalID:      "i-1234567890",
						ModulePath:      "web/compute",
					},
					{
						Address:         "aws_s3_bucket.data",
						Type:            "aws_s3_bucket",
						Name:            "data",
						ChangeType:      ChangeTypeReplace,
						IsDestructive:   true,
						ReplacementType: ReplacementConditional,
						PhysicalID:      "bucket-abcdef",
						ModulePath:      "-",
					},
				},
				Statistics: ChangeStatistics{
					ToAdd:        0,
					ToChange:     0,
					ToDestroy:    0,
					Replacements: 1,
					Conditionals: 1,
					Total:        2,
				},
			}

			err := formatter.OutputSummary(summary, format, true)

			// Restore stdout and read output
			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if err != nil {
				t.Errorf("OutputSummary() with format %s error = %v", format, err)
			}

			// Output should not be empty
			if strings.TrimSpace(output) == "" {
				t.Errorf("OutputSummary() with format %s produced empty output", format)
			}

			// For table format, verify specific content
			if format == "table" {
				expectedContent := []string{
					"Plan Information",
					"Summary for test.tfplan",
					"Resource Changes",
					"aws_instance.web",
					"aws_s3_bucket.data",
					"Always",
					"Conditional",
					"i-1234567890",
					"bucket-abcdef",
				}

				for _, expected := range expectedContent {
					if !strings.Contains(output, expected) {
						t.Errorf("Table format output missing expected content: %s", expected)
					}
				}
			}
		})
	}
}

func TestFormatter_ResourceChangesTable_WithDifferentResourceTypes(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := &config.Config{}
	formatter := NewFormatter(cfg)

	summary := &PlanSummary{
		ResourceChanges: []ResourceChange{
			{
				Address:         "aws_instance.web",
				Type:            "aws_instance",
				Name:            "web",
				ChangeType:      ChangeTypeCreate,
				IsDestructive:   false,
				ReplacementType: ReplacementNever,
				PhysicalID:      "-",
				ModulePath:      "-",
				IsDangerous:     false,
			},
			{
				Address:         "module.vpc.aws_vpc.main",
				Type:            "aws_vpc",
				Name:            "main",
				ChangeType:      ChangeTypeUpdate,
				IsDestructive:   false,
				ReplacementType: ReplacementNever,
				PhysicalID:      "vpc-123456",
				ModulePath:      "vpc",
				IsDangerous:     false,
			},
			{
				Address:         "aws_rds_instance.database",
				Type:            "aws_rds_instance",
				Name:            "database",
				ChangeType:      ChangeTypeReplace,
				IsDestructive:   true,
				ReplacementType: ReplacementAlways,
				PhysicalID:      "db-789012",
				ModulePath:      "-",
				IsDangerous:     true,
				DangerReason:    "Sensitive resource replacement",
			},
			{
				Address:         "aws_s3_bucket.logs",
				Type:            "aws_s3_bucket",
				Name:            "logs",
				ChangeType:      ChangeTypeDelete,
				IsDestructive:   true,
				ReplacementType: ReplacementNever,
				PhysicalID:      "logs-bucket-345",
				ModulePath:      "-",
				IsDangerous:     false,
			},
		},
	}

	err := formatter.formatResourceChangesTable(summary, "table")
	if err != nil {
		t.Errorf("formatResourceChangesTable() error = %v", err)
	}

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that all resource information is displayed correctly
	expectedContent := map[string]string{
		"aws_instance.web":        "Add",
		"module.vpc.aws_vpc.main": "Modify",
		"aws_rds_instance.database": "Replace",
		"aws_s3_bucket.logs":      "Remove",
		"vpc-123456":              "vpc-123456",
		"db-789012":               "db-789012",
		"logs-bucket-345":         "logs-bucket-345",
		"vpc":                     "vpc",
		"Always":                  "Always",
		"N/A":                     "N/A", // For delete operation
		"Never":                   "Never",
		"⚠️ Sensitive resource replacement": "⚠️ Sensitive resource replacement",
	}

	for content, description := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("formatResourceChangesTable() output missing %s: %s", description, content)
		}
	}

	// Verify table headers
	headers := []string{"ACTION", "RESOURCE", "TYPE", "ID", "REPLACEMENT", "MODULE", "DANGER"}
	for _, header := range headers {
		if !strings.Contains(output, header) {
			t.Errorf("formatResourceChangesTable() output missing header: %s", header)
		}
	}
}
