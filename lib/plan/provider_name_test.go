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

func testProviderChange(addr, typ, providerName string) *tfjson.ResourceChange {
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

// googleBetaPlan builds a plan where google_compute_instance resources are managed
// by both the google and googlebeta providers, plus resources without provider_name.
func googleBetaPlan() *tfjson.Plan {
	rc := testProviderChange
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
	wantCounts := map[string]int{
		"registry.terraform.io/hashicorp/google":     1,
		"registry.terraform.io/hashicorp/googlebeta": 2,
		"aws": 1, // no provider_name: grouped by type prefix
	}
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

	got := renderMarkdown(t, cfg, summary)
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

// renderMarkdown renders a summary through Formatter.OutputSummary to a markdown file.
func renderMarkdown(t *testing.T, cfg *config.Config, summary *PlanSummary) string {
	t.Helper()
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
	return string(data)
}

// TestProviderGrouping_SameTypeDifferentNamespace verifies that providers sharing a
// type name but from different namespaces are grouped separately and labelled
// unambiguously, while unambiguous providers keep their short labels.
func TestProviderGrouping_SameTypeDifferentNamespace(t *testing.T) {
	rc := testProviderChange
	tfPlan := &tfjson.Plan{
		FormatVersion:    "1.2",
		TerraformVersion: "1.9.0",
		ResourceChanges: []*tfjson.ResourceChange{
			rc("widget_thing.official", "widget_thing", "registry.terraform.io/hashicorp/widget"),
			rc("widget_thing.fork1", "widget_thing", "registry.terraform.io/acme/widget"),
			rc("widget_thing.fork2", "widget_thing", "registry.terraform.io/acme/widget"),
			rc("google_compute_instance.ga", "google_compute_instance", "registry.terraform.io/hashicorp/google"),
			rc("google_compute_instance.beta", "google_compute_instance", "registry.terraform.io/hashicorp/googlebeta"),
		},
	}
	cfg := config.GetDefaultConfig()
	cfg.Plan.Grouping.Threshold = 2
	analyzer := NewAnalyzer(tfPlan, cfg)
	summary, err := analyzer.GenerateSummary("")
	if err != nil {
		t.Fatalf("GenerateSummary() error = %v", err)
	}

	groups := analyzer.groupByProvider(summary.ResourceChanges)
	wantCounts := map[string]int{
		"registry.terraform.io/hashicorp/widget":     1,
		"registry.terraform.io/acme/widget":          2,
		"registry.terraform.io/hashicorp/google":     1,
		"registry.terraform.io/hashicorp/googlebeta": 1,
	}
	if len(groups) != len(wantCounts) {
		t.Fatalf("groupByProvider() returned %d groups, want %d: %v", len(groups), len(wantCounts), getMapKeys(groups))
	}
	for key, count := range wantCounts {
		if got := len(groups[key]); got != count {
			t.Errorf("group %q has %d resources, want %d", key, got, count)
		}
	}

	got := renderMarkdown(t, cfg, summary)
	for _, want := range []string{
		"HASHICORP/WIDGET Provider (1 changes)",
		"ACME/WIDGET Provider (2 changes)",
		"GOOGLE Provider (1 changes)",
		"GOOGLEBETA Provider (1 changes)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q\n%s", want, got)
		}
	}
	if strings.Contains(got, "REGISTRY.TERRAFORM.IO") {
		t.Errorf("labels should not include the default registry host\n%s", got)
	}
}

func TestProviderGroupLabels(t *testing.T) {
	tests := map[string]struct {
		keys []string
		want map[string]string
	}{
		"unambiguous providers use short names": {
			keys: []string{"registry.terraform.io/hashicorp/aws", "registry.terraform.io/hashicorp/googlebeta", "kubernetes"},
			want: map[string]string{
				"registry.terraform.io/hashicorp/aws":        "aws",
				"registry.terraform.io/hashicorp/googlebeta": "googlebeta",
				"kubernetes": "kubernetes",
			},
		},
		"colliding namespaces are disambiguated": {
			keys: []string{"registry.terraform.io/hashicorp/widget", "registry.terraform.io/acme/widget"},
			want: map[string]string{
				"registry.terraform.io/hashicorp/widget": "hashicorp/widget",
				"registry.terraform.io/acme/widget":      "acme/widget",
			},
		},
		"colliding hosts keep the custom host": {
			keys: []string{"registry.terraform.io/acme/widget", "example.com/acme/widget"},
			want: map[string]string{
				"registry.terraform.io/acme/widget": "acme/widget",
				"example.com/acme/widget":           "example.com/acme/widget",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := providerGroupLabels(tc.keys)
			if len(got) != len(tc.want) {
				t.Fatalf("providerGroupLabels() = %v, want %v", got, tc.want)
			}
			for key, want := range tc.want {
				if got[key] != want {
					t.Errorf("label for %q = %q, want %q", key, got[key], want)
				}
			}
		})
	}
}
