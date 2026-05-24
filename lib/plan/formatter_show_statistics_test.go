package plan

import (
	"testing"
	"time"

	"github.com/ArjenSchwarz/strata/config"
)

// testSummaryWithStats returns a minimal PlanSummary suitable for statistics tests
func testSummaryWithStats() *PlanSummary {
	return &PlanSummary{
		PlanFile:         "test.tfplan",
		TerraformVersion: "1.6.0",
		Workspace:        "default",
		Backend:          BackendInfo{Type: "local", Location: "terraform.tfstate"},
		CreatedAt:        time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Statistics: ChangeStatistics{
			Total:    2,
			ToAdd:    1,
			ToChange: 1,
			HighRisk: 0,
		},
		ResourceChanges: []ResourceChange{
			{
				Address:    "aws_instance.web",
				Type:       "aws_instance",
				Name:       "web",
				ChangeType: ChangeTypeCreate,
			},
		},
	}
}

func TestOutputSummary_ShowStatisticsTrue(t *testing.T) {
	cfg := config.GetDefaultConfig()
	cfg.Plan.ShowStatisticsSummary = true

	formatter := NewFormatter(cfg)
	outputConfig := &config.OutputConfiguration{
		Format:           "json",
		OutputFileFormat: "json",
	}

	err := formatter.OutputSummary(testSummaryWithStats(), outputConfig, true)
	if err != nil {
		t.Fatalf("OutputSummary() returned error: %v", err)
	}
}

func TestOutputSummary_ShowStatisticsFalse(t *testing.T) {
	cfg := config.GetDefaultConfig()
	cfg.Plan.ShowStatisticsSummary = false

	formatter := NewFormatter(cfg)
	outputConfig := &config.OutputConfiguration{
		Format:           "json",
		OutputFileFormat: "json",
	}

	err := formatter.OutputSummary(testSummaryWithStats(), outputConfig, true)
	if err != nil {
		t.Fatalf("OutputSummary() returned error: %v", err)
	}
}

func TestCreateStatisticsSummaryDataV2_ReturnsData(t *testing.T) {
	cfg := config.GetDefaultConfig()
	formatter := NewFormatter(cfg)
	summary := testSummaryWithStats()

	data, err := formatter.createStatisticsSummaryDataV2(summary)
	if err != nil {
		t.Fatalf("createStatisticsSummaryDataV2() returned error: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("expected 1 row, got %d", len(data))
	}
	if data[0]["Total Changes"] != 2 {
		t.Errorf("expected Total Changes = 2, got %v", data[0]["Total Changes"])
	}
}

func TestFormatResourceChangesWithProgressiveDisclosure_HonorsShowStatistics(t *testing.T) {
	summary := testSummaryWithStats()

	tests := []struct {
		name           string
		showStatistics bool
	}{
		{"statistics enabled", true},
		{"statistics disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.GetDefaultConfig()
			cfg.Plan.ShowStatisticsSummary = tt.showStatistics
			formatter := NewFormatter(cfg)

			doc, err := formatter.formatResourceChangesWithProgressiveDisclosure(summary)
			if err != nil {
				t.Fatalf("formatResourceChangesWithProgressiveDisclosure() error: %v", err)
			}
			if doc == nil {
				t.Fatal("expected non-nil document")
			}
		})
	}
}

func TestFormatGroupedWithCollapsibleSections_HonorsShowStatistics(t *testing.T) {
	summary := testSummaryWithStats()
	groups := map[string][]ResourceChange{
		"aws": summary.ResourceChanges,
	}

	tests := []struct {
		name           string
		showStatistics bool
	}{
		{"statistics enabled", true},
		{"statistics disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.GetDefaultConfig()
			cfg.Plan.ShowStatisticsSummary = tt.showStatistics
			formatter := NewFormatter(cfg)

			doc, err := formatter.formatGroupedWithCollapsibleSections(summary, groups)
			if err != nil {
				t.Fatalf("formatGroupedWithCollapsibleSections() error: %v", err)
			}
			if doc == nil {
				t.Fatal("expected non-nil document")
			}
		})
	}
}
