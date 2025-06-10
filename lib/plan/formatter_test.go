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
