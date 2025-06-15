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
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("formatPlanInfo() output missing expected string: %s", expected)
		}
	}

	// Check that "Terraform Version" is now just "Version"
	if strings.Contains(output, "Terraform Version") {
		t.Errorf("formatPlanInfo() output should not contain 'Terraform Version', it should be renamed to 'Version'")
	}

	// Check that "Dry Run" is not present
	if strings.Contains(output, "Dry Run") {
		t.Errorf("formatPlanInfo() output should not contain 'Dry Run'")
	}

	// Check for horizontal layout (keys in first row, values in second)
	// This is harder to test directly, but we can check for the presence of both rows
	if !strings.Contains(output, "Plan File") && !strings.Contains(output, "Version") &&
		!strings.Contains(output, "Workspace") && !strings.Contains(output, "Backend") &&
		!strings.Contains(output, "Created") {
		t.Errorf("formatPlanInfo() output should contain keys in the first row")
	}
}

func TestFormatter_formatPlanInfo_HorizontalLayout(t *testing.T) {
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

	// Check for horizontal layout by verifying the keys appear in the same row
	// and values appear in a separate row
	lines := strings.Split(output, "\n")

	// Find the line with the keys
	keyLineIndex := -1
	for i, line := range lines {
		if strings.Contains(line, "Plan File") && strings.Contains(line, "Version") {
			keyLineIndex = i
			break
		}
	}

	if keyLineIndex == -1 {
		t.Error("formatPlanInfo() should have a row with all keys")
		return
	}

	// Check if the next line contains the values
	if keyLineIndex+1 >= len(lines) {
		t.Error("formatPlanInfo() missing values row after keys row")
		return
	}

	valuesLine := lines[keyLineIndex+1]
	if !strings.Contains(valuesLine, "test.tfplan") || !strings.Contains(valuesLine, "1.6.0") {
		t.Error("formatPlanInfo() values should be in a separate row from keys")
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
			HighRisk:     1,
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
		"HIGH RISK",
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

func TestFormatter_OutputSummary_WithSensitiveOnly(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := &config.Config{
		Plan: config.PlanConfig{
			AlwaysShowSensitive: true,
		},
	}
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
				IsDangerous:   false,
			},
			{
				Address:       "aws_rds_instance.db",
				Type:          "aws_rds_instance",
				Name:          "db",
				ChangeType:    ChangeTypeReplace,
				IsDestructive: true,
				IsDangerous:   true,
				DangerReason:  "Sensitive resource replacement",
			},
		},
		Statistics: ChangeStatistics{
			ToAdd:        1,
			ToChange:     0,
			ToDestroy:    0,
			Replacements: 1,
			Conditionals: 0,
			HighRisk:     1,
			Total:        2,
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

	// Check that only sensitive resource details are shown
	expectedStrings := []string{
		"Sensitive Resource Changes",
		"aws_rds_instance.db",
		"Replace",
		"Sensitive resource replacement",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("OutputSummary() with sensitive only missing expected string: %s", expected)
		}
	}

	// Check that non-sensitive resources are not shown
	if strings.Contains(output, "aws_instance.web") {
		t.Errorf("OutputSummary() with sensitive only should not show non-sensitive resources")
	}
}

func TestFormatter_formatSensitiveResourceChanges(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := &config.Config{}
	formatter := NewFormatter(cfg)

	summary := &PlanSummary{
		ResourceChanges: []ResourceChange{
			{
				Address:       "aws_instance.web",
				Type:          "aws_instance",
				Name:          "web",
				ChangeType:    ChangeTypeCreate,
				IsDestructive: false,
				IsDangerous:   false,
			},
			{
				Address:       "aws_rds_instance.db",
				Type:          "aws_rds_instance",
				Name:          "db",
				ChangeType:    ChangeTypeReplace,
				IsDestructive: true,
				IsDangerous:   true,
				DangerReason:  "Sensitive resource replacement",
			},
			{
				Address:          "aws_s3_bucket.data",
				Type:             "aws_s3_bucket",
				Name:             "data",
				ChangeType:       ChangeTypeUpdate,
				IsDestructive:    false,
				IsDangerous:      true,
				DangerReason:     "Sensitive property change",
				DangerProperties: []string{"acl", "versioning"},
			},
		},
	}

	err := formatter.formatSensitiveResourceChanges(summary, "table")
	if err != nil {
		t.Errorf("formatSensitiveResourceChanges() error = %v", err)
	}

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that only sensitive resources are shown
	expectedStrings := []string{
		"Sensitive Resource Changes",
		"aws_rds_instance.db",
		"Replace",
		"Sensitive resource replacement",
		"aws_s3_bucket.data",
		"Modify",
		"Sensitive property change",
		"acl, versioning",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("formatSensitiveResourceChanges() missing expected string: %s", expected)
		}
	}

	// Check that non-sensitive resources are not shown
	if strings.Contains(output, "aws_instance.web") {
		t.Errorf("formatSensitiveResourceChanges() should not show non-sensitive resources")
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
				HighRisk:     2,
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
				HighRisk:     0,
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
				HighRisk:     1,
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
				HighRisk:     0,
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
			expectedHeaders := []string{"TOTAL", "ADDED", "REMOVED", "MODIFIED", "REPLACEMENTS", "CONDITIONALS", "HIGH RISK"}
			for _, header := range expectedHeaders {
				if !strings.Contains(output, header) {
					t.Errorf("Output should contain header %s", header)
				}
			}
		})
	}
}

func TestFormatter_OutputFormat_Compatibility(t *testing.T) {
	testFormats := []string{"table", "json", "markdown"}

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
					HighRisk:     1,
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

			// For markdown format, verify it contains markdown table syntax
			if format == "markdown" {
				if !strings.Contains(output, "|") {
					t.Errorf("Markdown format should contain table syntax with pipe characters")
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
		"aws_instance.web":                  "Add",
		"module.vpc.aws_vpc.main":           "Modify",
		"aws_rds_instance.database":         "Replace",
		"aws_s3_bucket.logs":                "Remove",
		"vpc-123456":                        "vpc-123456",
		"db-789012":                         "db-789012",
		"logs-bucket-345":                   "logs-bucket-345",
		"vpc":                               "vpc",
		"Always":                            "Always",
		"N/A":                               "N/A", // For delete operation
		"Never":                             "Never",
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

func TestFormatter_ValidateOutputFormat(t *testing.T) {
	cfg := &config.Config{}
	formatter := NewFormatter(cfg)

	testCases := []struct {
		name        string
		format      string
		shouldError bool
	}{
		{
			name:        "valid table format",
			format:      "table",
			shouldError: false,
		},
		{
			name:        "valid json format",
			format:      "json",
			shouldError: false,
		},
		{
			name:        "valid html format",
			format:      "html",
			shouldError: false,
		},
		{
			name:        "valid markdown format",
			format:      "markdown",
			shouldError: false,
		},
		{
			name:        "valid format with uppercase",
			format:      "TABLE",
			shouldError: false,
		},
		{
			name:        "valid format with mixed case",
			format:      "Json",
			shouldError: false,
		},
		{
			name:        "invalid format",
			format:      "xml",
			shouldError: true,
		},
		{
			name:        "empty format",
			format:      "",
			shouldError: true,
		},
		{
			name:        "invalid format with special chars",
			format:      "table@#$",
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := formatter.ValidateOutputFormat(tc.format)

			if tc.shouldError && err == nil {
				t.Errorf("Expected error for format '%s' but got none", tc.format)
			}
			if !tc.shouldError && err != nil {
				t.Errorf("Unexpected error for format '%s': %v", tc.format, err)
			}

			// Check error message format for invalid formats
			if tc.shouldError && err != nil {
				expectedSubstring := "unsupported output format"
				if !strings.Contains(err.Error(), expectedSubstring) {
					t.Errorf("Error message should contain '%s', got: %s", expectedSubstring, err.Error())
				}
			}
		})
	}
}

func TestFormatter_ErrorHandling(t *testing.T) {
	cfg := &config.Config{}
	formatter := NewFormatter(cfg)

	testCases := []struct {
		name         string
		methodName   string
		summary      *PlanSummary
		outputFormat string
		shouldError  bool
		errorMessage string
	}{
		{
			name:         "formatStatisticsSummary with nil summary",
			methodName:   "formatStatisticsSummary",
			summary:      nil,
			outputFormat: "table",
			shouldError:  true,
			errorMessage: "summary cannot be nil",
		},
		{
			name:         "formatStatisticsSummary with empty plan file",
			methodName:   "formatStatisticsSummary",
			summary:      &PlanSummary{PlanFile: ""},
			outputFormat: "table",
			shouldError:  true,
			errorMessage: "plan file name is required",
		},
		{
			name:         "formatPlanInfo with nil summary",
			methodName:   "formatPlanInfo",
			summary:      nil,
			outputFormat: "table",
			shouldError:  true,
			errorMessage: "summary cannot be nil",
		},
		{
			name:         "formatSensitiveResourceChanges with nil summary",
			methodName:   "formatSensitiveResourceChanges",
			summary:      nil,
			outputFormat: "table",
			shouldError:  true,
			errorMessage: "summary cannot be nil",
		},
		{
			name:         "formatResourceChangesTable with nil summary",
			methodName:   "formatResourceChangesTable",
			summary:      nil,
			outputFormat: "table",
			shouldError:  true,
			errorMessage: "summary cannot be nil",
		},
		{
			name:         "valid summary should not error",
			methodName:   "formatStatisticsSummary",
			summary:      &PlanSummary{PlanFile: "test.tfplan", Statistics: ChangeStatistics{}},
			outputFormat: "table",
			shouldError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout to prevent test output pollution
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			var err error
			switch tc.methodName {
			case "formatStatisticsSummary":
				err = formatter.formatStatisticsSummary(tc.summary, tc.outputFormat)
			case "formatPlanInfo":
				err = formatter.formatPlanInfo(tc.summary, tc.outputFormat)
			case "formatSensitiveResourceChanges":
				err = formatter.formatSensitiveResourceChanges(tc.summary, tc.outputFormat)
			case "formatResourceChangesTable":
				err = formatter.formatResourceChangesTable(tc.summary, tc.outputFormat)
			}

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Consume the output to prevent pipe blocking
			var buf bytes.Buffer
			io.Copy(&buf, r)

			if tc.shouldError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tc.shouldError && err != nil && tc.errorMessage != "" {
				if !strings.Contains(err.Error(), tc.errorMessage) {
					t.Errorf("Error message should contain '%s', got: %s", tc.errorMessage, err.Error())
				}
			}
		})
	}
}
