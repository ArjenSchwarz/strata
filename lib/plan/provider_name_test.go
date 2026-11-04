package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ArjenSchwarz/strata/config"
	tfjson "github.com/hashicorp/terraform-json"
)

func TestNormalizeProviderName(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"registry address":         {input: "registry.terraform.io/hashicorp/googlebeta", want: "googlebeta"},
		"third party registry":     {input: "registry.terraform.io/integrations/github", want: "github"},
		"custom registry host":     {input: "example.com/acme/widget", want: "widget"},
		"legacy provider prefix":   {input: "provider.aws", want: "aws"},
		"bare provider name":       {input: "aws", want: "aws"},
		"empty provider name":      {input: "", want: ""},
		"whitespace provider name": {input: "  ", want: ""},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := normalizeProviderName(tc.input); got != tc.want {
				t.Errorf("normalizeProviderName(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// googleBetaPlan builds a plan where google_compute_instance resources are managed
// by both the google and googlebeta providers, plus resources without provider_name.
func googleBetaPlan() *tfjson.Plan {
	rc := func(addr, typ, providerName string) *tfjson.ResourceChange {
		return &tfjson.ResourceChange{
			Address:      addr,
			Type:         typ,
			Name:         strings.ReplaceAll(addr, ".", "_"),
			ProviderName: providerName,
			Change: &tfjson.Change{
				Actions: tfjson.Actions{tfjson.ActionCreate},
				After:   map[string]any{"name": addr},
			},
		}
	}
	return &tfjson.Plan{
		FormatVersion:    "1.2",
		TerraformVersion: "1.9.0",
		ResourceChanges: []*tfjson.ResourceChange{
			rc("google_compute_instance.ga", "google_compute_instance", "registry.terraform.io/hashicorp/google"),
			rc("google_compute_instance.beta1", "google_compute_instance", "registry.terraform.io/hashicorp/googlebeta"),
			rc("google_compute_instance.beta2", "google_compute_instance", "registry.terraform.io/hashicorp/googlebeta"),
			rc("aws_s3_bucket.legacy", "aws_s3_bucket", ""),
		},
	}
}

func TestAnalyzer_ProviderFromProviderName(t *testing.T) {
	cfg := config.GetDefaultConfig()
	cfg.Plan.Grouping.Threshold = 2
	analyzer := NewAnalyzer(googleBetaPlan(), cfg)

	summary, err := analyzer.GenerateSummary("")
	if err != nil {
		t.Fatalf("GenerateSummary() error = %v", err)
	}

	want := map[string]string{
		"google_compute_instance.ga":    "google",
		"google_compute_instance.beta1": "googlebeta",
		"google_compute_instance.beta2": "googlebeta",
		"aws_s3_bucket.legacy":          "aws", // falls back to type prefix
	}
	for _, change := range summary.ResourceChanges {
		if got := change.Provider; got != want[change.Address] {
			t.Errorf("%s: Provider = %q, want %q", change.Address, got, want[change.Address])
		}
	}

	groups := analyzer.groupByProvider(summary.ResourceChanges)
	wantCounts := map[string]int{"google": 1, "googlebeta": 2, "aws": 1}
	if len(groups) != len(wantCounts) {
		t.Fatalf("groupByProvider() returned %d groups, want %d: %v", len(groups), len(wantCounts), getMapKeys(groups))
	}
	for provider, count := range wantCounts {
		if got := len(groups[provider]); got != count {
			t.Errorf("group %q has %d resources, want %d", provider, got, count)
		}
	}
}

// TestFormatter_GroupsByProviderName exercises the full analyze + render path and
// verifies googlebeta resources get their own provider section.
func TestFormatter_GroupsByProviderName(t *testing.T) {
	cfg := config.GetDefaultConfig()
	cfg.Plan.Grouping.Threshold = 2
	summary, err := NewAnalyzer(googleBetaPlan(), cfg).GenerateSummary("")
	if err != nil {
		t.Fatalf("GenerateSummary() error = %v", err)
	}

	outFile := filepath.Join(t.TempDir(), "summary.md")
	outputConfig := &config.OutputConfiguration{
		Format:           "markdown",
		OutputFile:       outFile,
		OutputFileFormat: "markdown",
	}
	if err := NewFormatter(cfg).OutputSummary(summary, outputConfig, true); err != nil {
		t.Fatalf("OutputSummary() error = %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	got := string(data)
	for _, want := range []string{
		"GOOGLEBETA Provider (2 changes)",
		"GOOGLE Provider (1 changes)",
		"AWS Provider (1 changes)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q\n%s", want, got)
		}
	}
}
