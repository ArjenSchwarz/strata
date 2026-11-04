package plan

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ArjenSchwarz/strata/config"
)

// renderPlanToFile runs the full parse → analyze → OutputSummary path for a testdata
// plan and returns the content written to the output file. Stdout is discarded.
func renderPlanToFile(t *testing.T, planFile string, cfg *config.Config, format string, showDetails bool) string {
	t.Helper()

	planPath := filepath.Join("..", "..", "testdata", planFile)
	tfPlan, err := NewParser(planPath).LoadPlan()
	if err != nil {
		t.Fatalf("LoadPlan() error = %v", err)
	}
	summary, err := NewAnalyzer(tfPlan, cfg).GenerateSummary(planPath)
	if err != nil {
		t.Fatalf("GenerateSummary() error = %v", err)
	}

	outputFile := filepath.Join(t.TempDir(), "summary."+format)
	outputConfig := &config.OutputConfiguration{
		Format:           format,
		OutputFile:       outputFile,
		OutputFileFormat: format,
	}

	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("failed to open %s: %v", os.DevNull, err)
	}
	oldStdout := os.Stdout
	os.Stdout = devNull
	t.Cleanup(func() {
		os.Stdout = oldStdout
		_ = devNull.Close()
	})

	err = NewFormatter(cfg).OutputSummary(summary, outputConfig, showDetails)
	os.Stdout = oldStdout
	if err != nil {
		t.Fatalf("OutputSummary() error = %v", err)
	}

	f, err := os.Open(outputFile)
	if err != nil {
		t.Fatalf("failed to open output file: %v", err)
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	return string(content)
}

// T-1567: the absent side of a sensitive output addition/deletion must render as "-",
// not "(sensitive value)".
func TestOutputSummary_SensitiveOutputAddRemoveAbsentSide(t *testing.T) {
	cfg := &config.Config{Plan: config.PlanConfig{ShowDetails: true}}
	got := renderPlanToFile(t, "sensitive_output_add_remove_plan.json", cfg, "csv", true)

	rows := map[string]string{}
	for line := range strings.SplitSeq(got, "\n") {
		if name, _, ok := strings.Cut(line, ","); ok {
			rows[name] = line
		}
	}

	tests := map[string]struct {
		output string
		want   string
	}{
		"sensitive addition has no current value": {
			output: "api_key",
			want:   "api_key,Add,-,(sensitive value),⚠️",
		},
		"sensitive deletion has no planned value": {
			output: "old_secret",
			want:   "old_secret,Remove,(sensitive value),-,⚠️",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if rows[tc.output] != tc.want {
				t.Errorf("row for %s = %q, want %q\nfull output:\n%s", tc.output, rows[tc.output], tc.want, got)
			}
		})
	}
}

// T-1703: --show-no-ops must work independently of --details. With details disabled the
// resource rows (including no-ops) are still rendered, but property details are omitted.
func TestOutputSummary_ShowNoOpsWithoutDetails(t *testing.T) {
	tests := map[string]struct {
		showNoOps           bool
		showDetails         bool
		alwaysShowSensitive bool
		wantContains        []string
		wantNotContains     []string
	}{
		"show-no-ops without details renders resource rows without property changes": {
			showNoOps:           true,
			showDetails:         false,
			alwaysShowSensitive: true,
			wantContains: []string{
				"### Resource Changes",
				"aws\\_s3\\_bucket.assets",
				"aws\\_instance.web\\_server",
				"aws\\_rds\\_instance.database",
			},
			wantNotContains: []string{
				"Sensitive Resource Changes",
				"No sensitive resource changes detected.",
				"Property Changes",
				"instance\\_type",
			},
		},
		"show-no-ops with details renders property changes": {
			showNoOps:    true,
			showDetails:  true,
			wantContains: []string{"### Resource Changes", "aws\\_s3\\_bucket.assets", "Property Changes"},
		},
		"details disabled without show-no-ops keeps sensitive-only view": {
			showNoOps:           false,
			showDetails:         false,
			alwaysShowSensitive: true,
			wantContains:        []string{"Sensitive Resource Changes"},
			wantNotContains:     []string{"### Resource Changes", "aws\\_s3\\_bucket.assets"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := &config.Config{
				Plan: config.PlanConfig{
					ShowDetails:         tc.showDetails,
					ShowNoOps:           tc.showNoOps,
					AlwaysShowSensitive: tc.alwaysShowSensitive,
					HighlightDangers:    true,
				},
				SensitiveProperties: []config.SensitiveProperty{
					{ResourceType: "aws_instance", Property: "user_data"},
				},
			}
			got := renderPlanToFile(t, "output_refinements_plan.json", cfg, "markdown", tc.showDetails)

			for _, want := range tc.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q\nfull output:\n%s", want, got)
				}
			}
			for _, notWant := range tc.wantNotContains {
				if strings.Contains(got, notWant) {
					t.Errorf("output unexpectedly contains %q\nfull output:\n%s", notWant, got)
				}
			}
		})
	}
}
