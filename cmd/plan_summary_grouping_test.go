package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// writeCombinedMultiProviderPlan merges the resource changes of two sample plans
// into a single multi-provider plan with more resources than the default threshold.
func writeCombinedMultiProviderPlan(t *testing.T, dir string) string {
	t.Helper()
	load := func(name string) map[string]json.RawMessage {
		t.Helper()
		data, err := os.ReadFile(filepath.Join("..", "samples", name))
		if err != nil {
			t.Fatalf("reading sample %s: %v", name, err)
		}
		var plan map[string]json.RawMessage
		if err := json.Unmarshal(data, &plan); err != nil {
			t.Fatalf("parsing sample %s: %v", name, err)
		}
		return plan
	}
	base := load("replacements-sample.json")
	extra := load("k8s-sample.json")

	var baseChanges, extraChanges []json.RawMessage
	if err := json.Unmarshal(base["resource_changes"], &baseChanges); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(extra["resource_changes"], &extraChanges); err != nil {
		t.Fatal(err)
	}
	merged, err := json.Marshal(append(baseChanges, extraChanges...))
	if err != nil {
		t.Fatal(err)
	}
	base["resource_changes"] = merged

	out, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "combined.json")
	if err := os.WriteFile(path, out, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestRunPlanSummary_LegacyGroupingConfig verifies T-1552: the legacy
// plan.group-by-provider and plan.grouping-threshold settings are honoured by
// `strata plan summary`, while the new plan.grouping block takes precedence.
func TestRunPlanSummary_LegacyGroupingConfig(t *testing.T) {
	tests := map[string]struct {
		configYAML  string
		wantGrouped bool
	}{
		"legacy group-by-provider false disables grouping": {
			configYAML:  "plan:\n  group-by-provider: false\n  grouping-threshold: 99\n",
			wantGrouped: false,
		},
		"legacy grouping-threshold above change count disables grouping": {
			configYAML:  "plan:\n  group-by-provider: true\n  grouping-threshold: 99\n",
			wantGrouped: false,
		},
		"legacy settings enabling grouping keep grouping": {
			configYAML:  "plan:\n  group-by-provider: true\n  grouping-threshold: 2\n",
			wantGrouped: true,
		},
		"new grouping block overrides legacy fields": {
			configYAML:  "plan:\n  group-by-provider: false\n  grouping-threshold: 99\n  grouping:\n    enabled: true\n    threshold: 2\n",
			wantGrouped: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			planPath := writeCombinedMultiProviderPlan(t, dir)
			configPath := filepath.Join(dir, "strata.yaml")
			if err := os.WriteFile(configPath, []byte(tc.configYAML), 0o600); err != nil {
				t.Fatal(err)
			}
			outFile := filepath.Join(dir, "summary.md")

			viper.Reset()
			t.Cleanup(viper.Reset)
			viper.SetConfigFile(configPath)
			if err := viper.ReadInConfig(); err != nil {
				t.Fatalf("ReadInConfig() error = %v", err)
			}
			viper.Set("output", "markdown")
			viper.Set("output-file", outFile)
			viper.Set("output-file-format", "markdown")
			viper.Set("plan.show-details", true)

			if err := runPlanSummary(planSummaryCmd, []string{planPath}); err != nil {
				t.Fatalf("runPlanSummary() error = %v", err)
			}

			data, err := os.ReadFile(outFile)
			if err != nil {
				t.Fatalf("reading output: %v", err)
			}
			got := string(data)
			gotGrouped := strings.Contains(got, "AWS Provider (") && strings.Contains(got, "KUBERNETES Provider (")
			if gotGrouped != tc.wantGrouped {
				t.Errorf("provider grouped output = %v, want %v\n%s", gotGrouped, tc.wantGrouped, got)
			}
			if !tc.wantGrouped && !strings.Contains(got, "Resource Changes") {
				t.Errorf("expected a single Resource Changes table\n%s", got)
			}
		})
	}
}
