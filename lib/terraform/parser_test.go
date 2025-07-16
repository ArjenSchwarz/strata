package terraform

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOutputParser(t *testing.T) {
	parser := NewOutputParser()
	require.NotNil(t, parser)

	// Verify it implements the interface
	var _ TerraformOutputParser = parser
}

func TestDefaultOutputParser_ParsePlanOutput(t *testing.T) {
	parser := NewOutputParser()

	tests := []struct {
		name     string
		output   string
		expected *PlanOutput
	}{
		{
			name:   "no changes detected",
			output: "No changes. Your infrastructure matches the configuration.",
			expected: &PlanOutput{
				HasChanges:      false,
				ResourceChanges: struct{ Add, Change, Destroy int }{0, 0, 0},
				RawOutput:       "No changes. Your infrastructure matches the configuration.",
				ExitCode:        0,
			},
		},
		{
			name:   "no changes detected - alternative message",
			output: "No changes. Infrastructure is up-to-date.",
			expected: &PlanOutput{
				HasChanges:      false,
				ResourceChanges: struct{ Add, Change, Destroy int }{0, 0, 0},
				RawOutput:       "No changes. Infrastructure is up-to-date.",
				ExitCode:        0,
			},
		},
		{
			name: "plan with changes",
			output: `Terraform will perform the following actions:

  # aws_instance.example will be created
  + resource "aws_instance" "example" {
      + ami           = "ami-12345678"
      + instance_type = "t2.micro"
    }

Plan: 1 to add, 0 to change, 0 to destroy.`,
			expected: &PlanOutput{
				HasChanges:      true,
				ResourceChanges: struct{ Add, Change, Destroy int }{1, 0, 0},
				RawOutput: `Terraform will perform the following actions:

  # aws_instance.example will be created
  + resource "aws_instance" "example" {
      + ami           = "ami-12345678"
      + instance_type = "t2.micro"
    }

Plan: 1 to add, 0 to change, 0 to destroy.`,
				ExitCode: 0,
			},
		},
		{
			name: "plan with multiple changes",
			output: `Terraform will perform the following actions:

Plan: 5 to add, 3 to change, 2 to destroy.`,
			expected: &PlanOutput{
				HasChanges:      true,
				ResourceChanges: struct{ Add, Change, Destroy int }{5, 3, 2},
				RawOutput: `Terraform will perform the following actions:

Plan: 5 to add, 3 to change, 2 to destroy.`,
				ExitCode: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParsePlanOutput(tt.output)

			require.NoError(t, err)
			assert.Equal(t, tt.expected.HasChanges, result.HasChanges)
			assert.Equal(t, tt.expected.ResourceChanges.Add, result.ResourceChanges.Add)
			assert.Equal(t, tt.expected.ResourceChanges.Change, result.ResourceChanges.Change)
			assert.Equal(t, tt.expected.ResourceChanges.Destroy, result.ResourceChanges.Destroy)
			assert.Equal(t, tt.expected.RawOutput, result.RawOutput)
			assert.Equal(t, tt.expected.ExitCode, result.ExitCode)
		})
	}
}

func TestDefaultOutputParser_ParseApplyOutput(t *testing.T) {
	parser := NewOutputParser()

	tests := []struct {
		name     string
		output   string
		expected *ApplyOutput
	}{
		{
			name: "successful apply",
			output: `aws_instance.example: Creating...
aws_instance.example: Creation complete after 30s [id=i-1234567890abcdef0]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.`,
			expected: &ApplyOutput{
				Success:         true,
				ResourceChanges: struct{ Added, Changed, Destroyed int }{1, 0, 0},
				Error:           "",
				RawOutput: `aws_instance.example: Creating...
aws_instance.example: Creation complete after 30s [id=i-1234567890abcdef0]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.`,
				ExitCode: 0,
			},
		},
		{
			name: "failed apply with error",
			output: `aws_instance.example: Creating...

Error: Error launching source instance: InvalidAMI.NotFound`,
			expected: &ApplyOutput{
				Success: false,
				Error:   "Error: Error launching source instance: InvalidAMI.NotFound",
				RawOutput: `aws_instance.example: Creating...

Error: Error launching source instance: InvalidAMI.NotFound`,
				ExitCode: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseApplyOutput(tt.output)

			require.NoError(t, err)
			assert.Equal(t, tt.expected.Success, result.Success)
			assert.Equal(t, tt.expected.ResourceChanges.Added, result.ResourceChanges.Added)
			assert.Equal(t, tt.expected.ResourceChanges.Changed, result.ResourceChanges.Changed)
			assert.Equal(t, tt.expected.ResourceChanges.Destroyed, result.ResourceChanges.Destroyed)
			assert.Equal(t, tt.expected.Error, result.Error)
			assert.Equal(t, tt.expected.RawOutput, result.RawOutput)
			assert.Equal(t, tt.expected.ExitCode, result.ExitCode)
		})
	}
}

func TestDefaultOutputParser_detectChangesFromResourceLines(t *testing.T) {
	parser := &DefaultOutputParser{}

	type expectedResult struct {
		hasChanges bool
		add        int
		change     int
		destroy    int
	}

	tests := []struct {
		name     string
		output   string
		expected expectedResult
	}{
		{
			name: "mixed resource indicators",
			output: `  + resource "aws_instance" "example1" {
  ~ resource "aws_s3_bucket" "example2" {
  - resource "aws_db_instance" "example3" {`,
			expected: expectedResult{
				hasChanges: true,
				add:        1,
				change:     1,
				destroy:    1,
			},
		},
		{
			name: "no resource changes",
			output: `Some other output
No resource indicators here`,
			expected: expectedResult{
				hasChanges: false,
				add:        0,
				change:     0,
				destroy:    0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &PlanOutput{}
			hasChanges := parser.detectChangesFromResourceLines(tt.output, result)

			assert.Equal(t, tt.expected.hasChanges, hasChanges)
			assert.Equal(t, tt.expected.add, result.ResourceChanges.Add)
			assert.Equal(t, tt.expected.change, result.ResourceChanges.Change)
			assert.Equal(t, tt.expected.destroy, result.ResourceChanges.Destroy)
		})
	}
}

func TestDefaultOutputParser_extractExitCode(t *testing.T) {
	parser := &DefaultOutputParser{}

	tests := []struct {
		name     string
		output   string
		expected int
	}{
		{
			name:     "no error",
			output:   "Plan: 1 to add, 0 to change, 0 to destroy.",
			expected: 0,
		},
		{
			name:     "with error",
			output:   "Error: Invalid configuration",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.extractExitCode(tt.output)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkParsePlanOutput(b *testing.B) {
	parser := NewOutputParser()
	output := `Terraform will perform the following actions:

  # aws_instance.example will be created
  + resource "aws_instance" "example" {
      + ami           = "ami-12345678"
      + instance_type = "t2.micro"
    }

Plan: 1 to add, 0 to change, 0 to destroy.`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParsePlanOutput(output)
	}
}
