package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ArjenSchwarz/strata/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// readForgetPlanJSON is a minimal plan with a data source read and a managed
// resource forget, the two Terraform actions that used to be hidden as no-ops (T-1415).
const readForgetPlanJSON = `{
  "format_version": "1.2",
  "terraform_version": "1.9.0",
  "resource_changes": [
    {
      "address": "data.aws_ami.ubuntu",
      "mode": "data",
      "type": "aws_ami",
      "name": "ubuntu",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {"actions": ["read"], "before": null, "after": {"id": "ami-123"}, "after_unknown": {"id": true}}
    },
    {
      "address": "aws_s3_bucket.legacy",
      "mode": "managed",
      "type": "aws_s3_bucket",
      "name": "legacy",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {"actions": ["forget"], "before": {"id": "legacy-bucket"}, "after": null}
    }
  ]
}`

// summarizeReadForgetPlan runs the plan through the same parser and analyzer path the CLI uses.
func summarizeReadForgetPlan(t *testing.T) (*PlanSummary, *config.Config) {
	t.Helper()
	planFile := filepath.Join(t.TempDir(), "read-forget.json")
	require.NoError(t, os.WriteFile(planFile, []byte(readForgetPlanJSON), 0o600))

	tfPlan, err := NewParser(planFile).LoadPlan()
	require.NoError(t, err)

	cfg := config.GetDefaultConfig()
	summary, err := NewAnalyzer(tfPlan, cfg).GenerateSummary(planFile)
	require.NoError(t, err)
	return summary, cfg
}

func TestReadAndForgetActions_AreNotNoOps(t *testing.T) {
	summary, _ := summarizeReadForgetPlan(t)

	got := map[string]ChangeType{}
	for _, rc := range summary.ResourceChanges {
		got[rc.Address] = rc.ChangeType
		assert.False(t, rc.IsNoOp, "%s should not be marked as no-op", rc.Address)
		assert.False(t, rc.IsDestructive, "%s should not be destructive", rc.Address)
	}
	assert.Equal(t, ChangeTypeRead, got["data.aws_ami.ubuntu"])
	assert.Equal(t, ChangeTypeForget, got["aws_s3_bucket.legacy"])

	stats := summary.Statistics
	assert.Equal(t, 1, stats.Read)
	assert.Equal(t, 1, stats.Forgotten)
	assert.Equal(t, 0, stats.Unmodified)
	assert.Equal(t, 0, stats.ToDestroy)
	assert.Equal(t, 2, stats.Total)
}

func TestReadAndForgetActions_AreRenderedInOutput(t *testing.T) {
	summary, cfg := summarizeReadForgetPlan(t)

	outputFile := filepath.Join(t.TempDir(), "summary.json")
	outputConfig := &config.OutputConfiguration{
		Format:           "json",
		OutputFile:       outputFile,
		OutputFileFormat: "json",
	}
	require.NoError(t, NewFormatter(cfg).OutputSummary(summary, outputConfig, true))

	raw, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	out := string(raw)

	assert.NotContains(t, out, "No changes detected")
	assert.Contains(t, out, "data.aws_ami.ubuntu")
	assert.Contains(t, out, "aws_s3_bucket.legacy")
	assert.Contains(t, out, `"Action": "Forget"`)
	assert.Contains(t, out, `"Action": "Read"`)
	assert.True(t, strings.Contains(out, `"Forgotten": 1`), "statistics should report the forgotten resource")
	assert.True(t, strings.Contains(out, `"Read": 1`), "statistics should report the data source read")
}
