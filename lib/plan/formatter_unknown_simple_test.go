package plan

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/ArjenSchwarz/strata/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnknownValueFormatting tests that unknown values display correctly in output
func TestUnknownValueFormatting(t *testing.T) {
	tests := []struct {
		name             string
		format           string
		resourceChanges  []ResourceChange
		expectedPatterns []string
	}{
		{
			name:   "table format with unknown values",
			format: "table",
			resourceChanges: []ResourceChange{
				{
					Address:    "aws_instance.example",
					Type:       "aws_instance",
					ChangeType: ChangeTypeCreate,
					PhysicalID: "-",
					PropertyChanges: PropertyChangeAnalysis{
						Count: 2,
						Changes: []PropertyChange{
							{
								Name:      "id",
								After:     "(known after apply)",
								Action:    "add",
								IsUnknown: true,
							},
							{
								Name:   "instance_type",
								After:  "t3.micro",
								Action: "add",
							},
						},
					},
					HasUnknownValues:  true,
					UnknownProperties: []string{"id"},
				},
			},
			expectedPatterns: []string{
				"aws_instance.example",
				"2 properties changed",
				"+ id = (known after apply)",     // Should NOT have quotes
				"+ instance_type = \"t3.micro\"", // Should have quotes
			},
		},
		{
			name:   "json format with unknown values",
			format: "json",
			resourceChanges: []ResourceChange{
				{
					Address:    "aws_instance.example",
					Type:       "aws_instance",
					ChangeType: ChangeTypeUpdate,
					PhysicalID: "i-123",
					PropertyChanges: PropertyChangeAnalysis{
						Count: 1,
						Changes: []PropertyChange{
							{
								Name:      "availability_zone",
								Before:    "us-east-1a",
								After:     "(known after apply)",
								Action:    "update",
								IsUnknown: true,
							},
						},
					},
					HasUnknownValues:  true,
					UnknownProperties: []string{"availability_zone"},
				},
			},
			expectedPatterns: []string{
				`"after":"(known after apply)"`, // JSON should have the value as a string
				`"is_unknown":true`,
				`"has_unknown_values":true`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create config
			cfg := &config.Config{
				Plan: config.PlanConfig{
					ShowDetails: true,
					ExpandableSections: config.ExpandableSectionsConfig{
						Enabled: true,
					},
				},
			}
			formatter := NewFormatter(cfg)

			// Create summary
			summary := &PlanSummary{
				PlanFile:         "test.tfplan",
				TerraformVersion: "1.6.0",
				Workspace:        "default",
				Backend: BackendInfo{
					Type:     "local",
					Location: "terraform.tfstate",
				},
				CreatedAt: time.Now(),
				Statistics: ChangeStatistics{
					Total:    len(tc.resourceChanges),
					ToAdd:    1,
					ToChange: 0,
				},
				ResourceChanges: tc.resourceChanges,
			}

			// Create output config
			outputConfig := &config.OutputConfiguration{
				Format:    tc.format,
				UseColors: false,
				UseEmoji:  false,
			}

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run formatter
			err := formatter.OutputSummary(summary, outputConfig, true)
			require.NoError(t, err)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Check patterns
			for _, pattern := range tc.expectedPatterns {
				assert.Contains(t, output, pattern,
					"Output should contain pattern: %s", pattern)
			}
		})
	}
}

// TestPropertyChangeFormattingWithUnknownValues specifically tests the formatPropertyChange function
func TestPropertyChangeFormattingWithUnknownValues(t *testing.T) {
	formatter := NewFormatter(&config.Config{})

	tests := []struct {
		name     string
		change   PropertyChange
		expected string
	}{
		{
			name: "add with unknown value",
			change: PropertyChange{
				Name:      "id",
				After:     "(known after apply)",
				Action:    "add",
				IsUnknown: true,
			},
			expected: "  + id = (known after apply)",
		},
		{
			name: "update to unknown",
			change: PropertyChange{
				Name:      "endpoint",
				Before:    "old.example.com",
				After:     "(known after apply)",
				Action:    "update",
				IsUnknown: true,
			},
			expected: `  ~ endpoint = "old.example.com" -> (known after apply)`,
		},
		{
			name: "complex structure becoming unknown",
			change: PropertyChange{
				Name:      "security_groups",
				Before:    []any{"sg-123", "sg-456"},
				After:     "(known after apply)",
				Action:    "update",
				IsUnknown: true,
			},
			expected: `  ~ security_groups = [ "sg-123", "sg-456" ] -> (known after apply)`,
		},
		{
			name: "simple unknown value without quotes",
			change: PropertyChange{
				Name:      "simple_value",
				After:     "(known after apply)",
				Action:    "add",
				IsUnknown: true,
			},
			expected: `  + simple_value = (known after apply)`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatter.formatPropertyChange(tc.change)
			assert.Equal(t, tc.expected, result)
		})
	}
}
