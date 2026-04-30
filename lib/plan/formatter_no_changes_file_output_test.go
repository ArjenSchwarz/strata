package plan

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ArjenSchwarz/strata/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOutputSummary_NoChanges_CreatesOutputFile verifies that when a plan has no
// resource or output changes, the output file is still created (T-351).
func TestOutputSummary_NoChanges_CreatesOutputFile(t *testing.T) {
	cfg := &config.Config{}
	formatter := NewFormatter(cfg)

	summary := &PlanSummary{
		PlanFile:         "empty.tfplan",
		TerraformVersion: "1.6.0",
		Workspace:        "default",
		Backend:          BackendInfo{Type: "local", Location: "terraform.tfstate"},
		CreatedAt:        time.Now(),
		Statistics:       ChangeStatistics{},
		ResourceChanges:  []ResourceChange{},
		OutputChanges:    []OutputChange{},
	}

	outputFile := filepath.Join(t.TempDir(), "no-changes.json")
	outputConfig := &config.OutputConfiguration{
		Format:           "json",
		OutputFile:       outputFile,
		OutputFileFormat: "json",
	}

	err := formatter.OutputSummary(summary, outputConfig, true)
	require.NoError(t, err)

	_, err = os.Stat(outputFile)
	assert.NoError(t, err, "output file should be created even when plan has no changes")
}
