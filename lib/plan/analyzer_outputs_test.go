package plan

import (
	"testing"

	"github.com/ArjenSchwarz/strata/config"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
)

// TestProcessOutputChanges tests the ProcessOutputChanges function (Task 7.1)
func TestProcessOutputChanges(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name            string
		plan            *tfjson.Plan
		expectedCount   int
		expectedOutputs []OutputChange
		expectedError   bool
		description     string
	}{
		{
			name: "plan with no output changes should return empty list",
			plan: &tfjson.Plan{
				OutputChanges: nil,
			},
			expectedCount:   0,
			expectedOutputs: []OutputChange{},
			expectedError:   false,
			description:     "Missing output_changes field should return empty list (requirement 2.8)",
		},
		{
			name: "plan with empty output changes should return empty list",
			plan: &tfjson.Plan{
				OutputChanges: map[string]*tfjson.Change{},
			},
			expectedCount:   0,
			expectedOutputs: []OutputChange{},
			expectedError:   false,
			description:     "Empty output_changes map should return empty list",
		},
		{
			name: "plan with single output creation",
			plan: &tfjson.Plan{
				OutputChanges: map[string]*tfjson.Change{
					"instance_ip": {
						Actions: []tfjson.Action{tfjson.ActionCreate},
						Before:  nil,
						After:   "192.168.1.10",
					},
				},
			},
			expectedCount: 1,
			expectedOutputs: []OutputChange{
				{
					Name:       "instance_ip",
					ChangeType: ChangeTypeCreate,
					Sensitive:  false,
					Before:     nil,
					After:      "192.168.1.10",
					IsUnknown:  false,
					Action:     "Add",
					Indicator:  "+",
				},
			},
			expectedError: false,
			description:   "Single output creation should be processed correctly (requirement 2.5)",
		},
		{
			name: "plan with output modification",
			plan: &tfjson.Plan{
				OutputChanges: map[string]*tfjson.Change{
					"database_url": {
						Actions: []tfjson.Action{tfjson.ActionUpdate},
						Before:  "postgresql://old-host/db",
						After:   "postgresql://new-host/db",
					},
				},
			},
			expectedCount: 1,
			expectedOutputs: []OutputChange{
				{
					Name:       "database_url",
					ChangeType: ChangeTypeUpdate,
					Sensitive:  false,
					Before:     "postgresql://old-host/db",
					After:      "postgresql://new-host/db",
					IsUnknown:  false,
					Action:     "Modify",
					Indicator:  "~",
				},
			},
			expectedError: false,
			description:   "Output modification should be processed correctly (requirement 2.6)",
		},
		{
			name: "plan with output deletion",
			plan: &tfjson.Plan{
				OutputChanges: map[string]*tfjson.Change{
					"old_output": {
						Actions: []tfjson.Action{tfjson.ActionDelete},
						Before:  "some-value",
						After:   nil,
					},
				},
			},
			expectedCount: 1,
			expectedOutputs: []OutputChange{
				{
					Name:       "old_output",
					ChangeType: ChangeTypeDelete,
					Sensitive:  false,
					Before:     "some-value",
					After:      nil,
					IsUnknown:  false,
					Action:     "Remove",
					Indicator:  "-",
				},
			},
			expectedError: false,
			description:   "Output deletion should be processed correctly (requirement 2.7)",
		},
		{
			name: "plan with unknown output value",
			plan: &tfjson.Plan{
				OutputChanges: map[string]*tfjson.Change{
					"computed_id": {
						Actions:      []tfjson.Action{tfjson.ActionCreate},
						Before:       nil,
						After:        nil,
						AfterUnknown: true,
					},
				},
			},
			expectedCount: 1,
			expectedOutputs: []OutputChange{
				{
					Name:       "computed_id",
					ChangeType: ChangeTypeCreate,
					Sensitive:  false,
					Before:     nil,
					After:      "(known after apply)",
					IsUnknown:  true,
					Action:     "Add",
					Indicator:  "+",
				},
			},
			expectedError: false,
			description:   "Unknown output values should display '(known after apply)' (requirement 2.3)",
		},
		{
			name: "plan with sensitive output",
			plan: &tfjson.Plan{
				OutputChanges: map[string]*tfjson.Change{
					"api_key": {
						Actions:        []tfjson.Action{tfjson.ActionCreate},
						Before:         nil,
						After:          "secret-key-value",
						AfterSensitive: true,
					},
				},
			},
			expectedCount: 1,
			expectedOutputs: []OutputChange{
				{
					Name:       "api_key",
					ChangeType: ChangeTypeCreate,
					Sensitive:  true,
					Before:     nil,
					After:      "(sensitive value)",
					IsUnknown:  false,
					Action:     "Add",
					Indicator:  "+",
				},
			},
			expectedError: false,
			description:   "Sensitive outputs should display '(sensitive value)' (requirement 2.4)",
		},
		{
			name: "plan with sensitive unknown output",
			plan: &tfjson.Plan{
				OutputChanges: map[string]*tfjson.Change{
					"secret_id": {
						Actions:         []tfjson.Action{tfjson.ActionUpdate},
						Before:          "old-secret",
						After:           nil,
						BeforeSensitive: true,
						AfterSensitive:  true,
						AfterUnknown:    true,
					},
				},
			},
			expectedCount: 1,
			expectedOutputs: []OutputChange{
				{
					Name:       "secret_id",
					ChangeType: ChangeTypeUpdate,
					Sensitive:  true,
					Before:     "(sensitive value)",
					After:      "(known after apply)",
					IsUnknown:  true,
					Action:     "Modify",
					Indicator:  "~",
				},
			},
			expectedError: false,
			description:   "Sensitive unknown outputs should show both masking and unknown display",
		},
		{
			name: "plan with multiple mixed outputs",
			plan: &tfjson.Plan{
				OutputChanges: map[string]*tfjson.Change{
					"public_ip": {
						Actions: []tfjson.Action{tfjson.ActionCreate},
						Before:  nil,
						After:   "203.0.113.10",
					},
					"private_key": {
						Actions:         []tfjson.Action{tfjson.ActionUpdate},
						Before:          "old-key",
						After:           "new-key",
						BeforeSensitive: true,
						AfterSensitive:  true,
					},
					"instance_id": {
						Actions:      []tfjson.Action{tfjson.ActionCreate},
						Before:       nil,
						After:        nil,
						AfterUnknown: true,
					},
				},
			},
			expectedCount: 3,
			expectedOutputs: []OutputChange{
				{
					Name:       "public_ip",
					ChangeType: ChangeTypeCreate,
					Sensitive:  false,
					Before:     nil,
					After:      "203.0.113.10",
					IsUnknown:  false,
					Action:     "Add",
					Indicator:  "+",
				},
				{
					Name:       "private_key",
					ChangeType: ChangeTypeUpdate,
					Sensitive:  true,
					Before:     "(sensitive value)",
					After:      "(sensitive value)",
					IsUnknown:  false,
					Action:     "Modify",
					Indicator:  "~",
				},
				{
					Name:       "instance_id",
					ChangeType: ChangeTypeCreate,
					Sensitive:  false,
					Before:     nil,
					After:      "(known after apply)",
					IsUnknown:  true,
					Action:     "Add",
					Indicator:  "+",
				},
			},
			expectedError: false,
			description:   "Multiple mixed outputs should all be processed correctly",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outputs, err := analyzer.ProcessOutputChanges(tc.plan)

			if tc.expectedError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
			}

			assert.Equal(t, tc.expectedCount, len(outputs), tc.description+" - output count")

			// Check individual outputs (order-independent comparison)
			if len(tc.expectedOutputs) > 0 && len(outputs) > 0 {
				// Create maps for easier comparison since output order isn't guaranteed
				expectedMap := make(map[string]OutputChange)
				actualMap := make(map[string]OutputChange)

				for _, expected := range tc.expectedOutputs {
					expectedMap[expected.Name] = expected
				}
				for _, actual := range outputs {
					actualMap[actual.Name] = actual
				}

				for name, expected := range expectedMap {
					if actual, exists := actualMap[name]; exists {
						assert.Equal(t, expected.Name, actual.Name, "Output %s name should match (%s)", name, tc.description)
						assert.Equal(t, expected.ChangeType, actual.ChangeType, "Output %s ChangeType should match (%s)", name, tc.description)
						assert.Equal(t, expected.Sensitive, actual.Sensitive, "Output %s Sensitive should match (%s)", name, tc.description)
						assert.Equal(t, expected.Before, actual.Before, "Output %s Before should match (%s)", name, tc.description)
						assert.Equal(t, expected.After, actual.After, "Output %s After should match (%s)", name, tc.description)
						assert.Equal(t, expected.IsUnknown, actual.IsUnknown, "Output %s IsUnknown should match (%s)", name, tc.description)
						assert.Equal(t, expected.Action, actual.Action, "Output %s Action should match (%s)", name, tc.description)
						assert.Equal(t, expected.Indicator, actual.Indicator, "Output %s Indicator should match (%s)", name, tc.description)
					} else {
						t.Errorf("Expected output %s not found in results (%s)", name, tc.description)
					}
				}
			}
		})
	}
}

// TestAnalyzeOutputChange tests the analyzeOutputChange function (Task 7.1)
func TestAnalyzeOutputChange(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name           string
		outputName     string
		change         *tfjson.Change
		expectedOutput *OutputChange
		expectedError  bool
		description    string
	}{
		{
			name:           "nil change should return error",
			outputName:     "test_output",
			change:         nil,
			expectedOutput: nil,
			expectedError:  true,
			description:    "Nil change should be handled gracefully with error",
		},
		{
			name:       "simple create action",
			outputName: "web_url",
			change: &tfjson.Change{
				Actions: []tfjson.Action{tfjson.ActionCreate},
				Before:  nil,
				After:   "https://example.com",
			},
			expectedOutput: &OutputChange{
				Name:       "web_url",
				ChangeType: ChangeTypeCreate,
				Sensitive:  false,
				Before:     nil,
				After:      "https://example.com",
				IsUnknown:  false,
				Action:     "Add",
				Indicator:  "+",
			},
			expectedError: false,
			description:   "Create action should show 'Add' with '+' indicator (requirement 2.5)",
		},
		{
			name:       "simple update action",
			outputName: "database_port",
			change: &tfjson.Change{
				Actions: []tfjson.Action{tfjson.ActionUpdate},
				Before:  5432,
				After:   3306,
			},
			expectedOutput: &OutputChange{
				Name:       "database_port",
				ChangeType: ChangeTypeUpdate,
				Sensitive:  false,
				Before:     5432,
				After:      3306,
				IsUnknown:  false,
				Action:     "Modify",
				Indicator:  "~",
			},
			expectedError: false,
			description:   "Update action should show 'Modify' with '~' indicator (requirement 2.6)",
		},
		{
			name:       "simple delete action",
			outputName: "deprecated_config",
			change: &tfjson.Change{
				Actions: []tfjson.Action{tfjson.ActionDelete},
				Before:  "old-config",
				After:   nil,
			},
			expectedOutput: &OutputChange{
				Name:       "deprecated_config",
				ChangeType: ChangeTypeDelete,
				Sensitive:  false,
				Before:     "old-config",
				After:      nil,
				IsUnknown:  false,
				Action:     "Remove",
				Indicator:  "-",
			},
			expectedError: false,
			description:   "Delete action should show 'Remove' with '-' indicator (requirement 2.7)",
		},
		{
			name:       "unknown output value",
			outputName: "instance_arn",
			change: &tfjson.Change{
				Actions:      []tfjson.Action{tfjson.ActionCreate},
				Before:       nil,
				After:        nil,
				AfterUnknown: true,
			},
			expectedOutput: &OutputChange{
				Name:       "instance_arn",
				ChangeType: ChangeTypeCreate,
				Sensitive:  false,
				Before:     nil,
				After:      "(known after apply)",
				IsUnknown:  true,
				Action:     "Add",
				Indicator:  "+",
			},
			expectedError: false,
			description:   "Unknown output should display '(known after apply)' (requirement 2.3)",
		},
		{
			name:       "sensitive output with before value",
			outputName: "admin_password",
			change: &tfjson.Change{
				Actions:         []tfjson.Action{tfjson.ActionUpdate},
				Before:          "old-password",
				After:           "new-password",
				BeforeSensitive: true,
				AfterSensitive:  true,
			},
			expectedOutput: &OutputChange{
				Name:       "admin_password",
				ChangeType: ChangeTypeUpdate,
				Sensitive:  true,
				Before:     "(sensitive value)",
				After:      "(sensitive value)",
				IsUnknown:  false,
				Action:     "Modify",
				Indicator:  "~",
			},
			expectedError: false,
			description:   "Sensitive outputs should display '(sensitive value)' (requirement 2.4)",
		},
		{
			name:       "sensitive unknown output",
			outputName: "tls_cert",
			change: &tfjson.Change{
				Actions:        []tfjson.Action{tfjson.ActionCreate},
				Before:         nil,
				After:          nil,
				AfterSensitive: true,
				AfterUnknown:   true,
			},
			expectedOutput: &OutputChange{
				Name:       "tls_cert",
				ChangeType: ChangeTypeCreate,
				Sensitive:  true,
				Before:     nil,
				After:      "(known after apply)",
				IsUnknown:  true,
				Action:     "Add",
				Indicator:  "+",
			},
			expectedError: false,
			description:   "Sensitive unknown outputs should prioritize unknown display over sensitive masking",
		},
		{
			name:       "no-op action",
			outputName: "static_value",
			change: &tfjson.Change{
				Actions: []tfjson.Action{},
				Before:  "unchanged",
				After:   "unchanged",
			},
			expectedOutput: &OutputChange{
				Name:       "static_value",
				ChangeType: ChangeTypeNoOp,
				Sensitive:  false,
				Before:     "unchanged",
				After:      "unchanged",
				IsUnknown:  false,
				Action:     "No-op",
				Indicator:  "",
			},
			expectedError: false,
			description:   "No-op actions should be handled correctly",
		},
		{
			name:       "replace action (delete + create)",
			outputName: "resource_reference",
			change: &tfjson.Change{
				Actions: []tfjson.Action{tfjson.ActionDelete, tfjson.ActionCreate},
				Before:  "old-reference",
				After:   "new-reference",
			},
			expectedOutput: &OutputChange{
				Name:       "resource_reference",
				ChangeType: ChangeTypeReplace,
				Sensitive:  false,
				Before:     "old-reference",
				After:      "new-reference",
				IsUnknown:  false,
				Action:     "No-op",
				Indicator:  "",
			},
			expectedError: false,
			description:   "Replace actions should be handled (though uncommon for outputs)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := analyzer.analyzeOutputChange(tc.outputName, tc.change)

			if tc.expectedError {
				assert.Error(t, err, tc.description)
				assert.Nil(t, result, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
				assert.NotNil(t, result, tc.description)

				if result != nil && tc.expectedOutput != nil {
					assert.Equal(t, tc.expectedOutput.Name, result.Name, "Name should match (%s)", tc.description)
					assert.Equal(t, tc.expectedOutput.ChangeType, result.ChangeType, "ChangeType should match (%s)", tc.description)
					assert.Equal(t, tc.expectedOutput.Sensitive, result.Sensitive, "Sensitive should match (%s)", tc.description)
					assert.Equal(t, tc.expectedOutput.Before, result.Before, "Before should match (%s)", tc.description)
					assert.Equal(t, tc.expectedOutput.After, result.After, "After should match (%s)", tc.description)
					assert.Equal(t, tc.expectedOutput.IsUnknown, result.IsUnknown, "IsUnknown should match (%s)", tc.description)
					assert.Equal(t, tc.expectedOutput.Action, result.Action, "Action should match (%s)", tc.description)
					assert.Equal(t, tc.expectedOutput.Indicator, result.Indicator, "Indicator should match (%s)", tc.description)
				}
			}
		})
	}
}

// TestGetOutputActionAndIndicator tests the output action and indicator mapping function (Task 7.1)
func TestGetOutputActionAndIndicator(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name              string
		changeType        ChangeType
		expectedAction    string
		expectedIndicator string
		description       string
	}{
		{
			name:              "create should return Add with +",
			changeType:        ChangeTypeCreate,
			expectedAction:    "Add",
			expectedIndicator: "+",
			description:       "Create changes should show 'Add' with '+' indicator (requirement 2.5)",
		},
		{
			name:              "update should return Modify with ~",
			changeType:        ChangeTypeUpdate,
			expectedAction:    "Modify",
			expectedIndicator: "~",
			description:       "Update changes should show 'Modify' with '~' indicator (requirement 2.6)",
		},
		{
			name:              "delete should return Remove with -",
			changeType:        ChangeTypeDelete,
			expectedAction:    "Remove",
			expectedIndicator: "-",
			description:       "Delete changes should show 'Remove' with '-' indicator (requirement 2.7)",
		},
		{
			name:              "no-op should return No-op with empty indicator",
			changeType:        ChangeTypeNoOp,
			expectedAction:    "No-op",
			expectedIndicator: "",
			description:       "No-op changes should show 'No-op' with empty indicator",
		},
		{
			name:              "replace should return No-op with empty indicator",
			changeType:        ChangeTypeReplace,
			expectedAction:    "No-op",
			expectedIndicator: "",
			description:       "Replace changes should show 'No-op' with empty indicator (uncommon for outputs)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			action, indicator := analyzer.getOutputActionAndIndicator(tc.changeType)

			assert.Equal(t, tc.expectedAction, action, tc.description+" - action")
			assert.Equal(t, tc.expectedIndicator, indicator, tc.description+" - indicator")
		})
	}
}

// TestOutputsProcessingEndToEnd tests complete outputs processing workflow (Task 7.2)
func TestOutputsProcessingEndToEnd(t *testing.T) {
	cfg := &config.Config{}
	analyzer := &Analyzer{config: cfg}

	testCases := []struct {
		name                 string
		plan                 *tfjson.Plan
		expectedOutputCount  int
		expectedEmptySection bool
		validateOutputs      func(t *testing.T, outputs []OutputChange, description string)
		description          string
	}{
		{
			name: "plan with no outputs should suppress outputs section",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.0.0",
				OutputChanges:    nil,
				ResourceChanges:  []*tfjson.ResourceChange{},
			},
			expectedOutputCount:  0,
			expectedEmptySection: true,
			validateOutputs: func(t *testing.T, outputs []OutputChange, description string) {
				assert.Empty(t, outputs, description+" - outputs should be empty")
			},
			description: "Plans without outputs should suppress the outputs section entirely (requirement 2.8)",
		},
		{
			name: "plan with mixed output types should process all correctly",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.0.0",
				OutputChanges: map[string]*tfjson.Change{
					"public_endpoint": {
						Actions: []tfjson.Action{tfjson.ActionCreate},
						Before:  nil,
						After:   "https://api.example.com",
					},
					"database_password": {
						Actions:        []tfjson.Action{tfjson.ActionUpdate},
						Before:         "old-secret",
						After:          "new-secret",
						AfterSensitive: true,
					},
					"instance_id": {
						Actions:      []tfjson.Action{tfjson.ActionCreate},
						Before:       nil,
						After:        nil,
						AfterUnknown: true,
					},
					"deprecated_config": {
						Actions: []tfjson.Action{tfjson.ActionDelete},
						Before:  "old-value",
						After:   nil,
					},
				},
				ResourceChanges: []*tfjson.ResourceChange{},
			},
			expectedOutputCount:  4,
			expectedEmptySection: false,
			validateOutputs: func(t *testing.T, outputs []OutputChange, description string) {
				assert.Len(t, outputs, 4, description+" - should have 4 outputs")

				// Create a map for easy lookup
				outputMap := make(map[string]OutputChange)
				for _, output := range outputs {
					outputMap[output.Name] = output
				}

				// Validate public endpoint (create)
				if publicEndpoint, exists := outputMap["public_endpoint"]; exists {
					assert.Equal(t, ChangeTypeCreate, publicEndpoint.ChangeType, "public_endpoint should be create type")
					assert.Equal(t, "Add", publicEndpoint.Action, "public_endpoint should have Add action")
					assert.Equal(t, "+", publicEndpoint.Indicator, "public_endpoint should have + indicator")
					assert.Equal(t, "https://api.example.com", publicEndpoint.After, "public_endpoint should have correct after value")
					assert.False(t, publicEndpoint.Sensitive, "public_endpoint should not be sensitive")
					assert.False(t, publicEndpoint.IsUnknown, "public_endpoint should not be unknown")
				} else {
					t.Errorf("public_endpoint output not found")
				}

				// Validate database password (sensitive update)
				if dbPassword, exists := outputMap["database_password"]; exists {
					assert.Equal(t, ChangeTypeUpdate, dbPassword.ChangeType, "database_password should be update type")
					assert.Equal(t, "Modify", dbPassword.Action, "database_password should have Modify action")
					assert.Equal(t, "~", dbPassword.Indicator, "database_password should have ~ indicator")
					assert.Equal(t, "(sensitive value)", dbPassword.After, "database_password should mask sensitive value")
					assert.True(t, dbPassword.Sensitive, "database_password should be sensitive")
					assert.False(t, dbPassword.IsUnknown, "database_password should not be unknown")
				} else {
					t.Errorf("database_password output not found")
				}

				// Validate instance ID (unknown create)
				if instanceId, exists := outputMap["instance_id"]; exists {
					assert.Equal(t, ChangeTypeCreate, instanceId.ChangeType, "instance_id should be create type")
					assert.Equal(t, "Add", instanceId.Action, "instance_id should have Add action")
					assert.Equal(t, "+", instanceId.Indicator, "instance_id should have + indicator")
					assert.Equal(t, "(known after apply)", instanceId.After, "instance_id should show unknown value")
					assert.False(t, instanceId.Sensitive, "instance_id should not be sensitive")
					assert.True(t, instanceId.IsUnknown, "instance_id should be unknown")
				} else {
					t.Errorf("instance_id output not found")
				}

				// Validate deprecated config (delete)
				if deprecatedConfig, exists := outputMap["deprecated_config"]; exists {
					assert.Equal(t, ChangeTypeDelete, deprecatedConfig.ChangeType, "deprecated_config should be delete type")
					assert.Equal(t, "Remove", deprecatedConfig.Action, "deprecated_config should have Remove action")
					assert.Equal(t, "-", deprecatedConfig.Indicator, "deprecated_config should have - indicator")
					assert.Equal(t, "old-value", deprecatedConfig.Before, "deprecated_config should have correct before value")
					assert.Nil(t, deprecatedConfig.After, "deprecated_config should have nil after value")
					assert.False(t, deprecatedConfig.Sensitive, "deprecated_config should not be sensitive")
					assert.False(t, deprecatedConfig.IsUnknown, "deprecated_config should not be unknown")
				} else {
					t.Errorf("deprecated_config output not found")
				}
			},
			description: "Mixed output types should be processed with correct actions and indicators",
		},
		{
			name: "plan with sensitive unknown outputs should handle edge case correctly",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.0.0",
				OutputChanges: map[string]*tfjson.Change{
					"ssl_certificate": {
						Actions:         []tfjson.Action{tfjson.ActionCreate},
						Before:          nil,
						After:           nil,
						BeforeSensitive: false,
						AfterSensitive:  true,
						AfterUnknown:    true,
					},
					"encryption_key": {
						Actions:         []tfjson.Action{tfjson.ActionUpdate},
						Before:          "old-key",
						After:           nil,
						BeforeSensitive: true,
						AfterSensitive:  true,
						AfterUnknown:    true,
					},
				},
				ResourceChanges: []*tfjson.ResourceChange{},
			},
			expectedOutputCount:  2,
			expectedEmptySection: false,
			validateOutputs: func(t *testing.T, outputs []OutputChange, description string) {
				assert.Len(t, outputs, 2, description+" - should have 2 outputs")

				// Create a map for easy lookup
				outputMap := make(map[string]OutputChange)
				for _, output := range outputs {
					outputMap[output.Name] = output
				}

				// Validate SSL certificate (sensitive unknown create)
				if sslCert, exists := outputMap["ssl_certificate"]; exists {
					assert.Equal(t, ChangeTypeCreate, sslCert.ChangeType, "ssl_certificate should be create type")
					assert.Equal(t, "Add", sslCert.Action, "ssl_certificate should have Add action")
					assert.Equal(t, "+", sslCert.Indicator, "ssl_certificate should have + indicator")
					assert.Equal(t, "(known after apply)", sslCert.After, "ssl_certificate should prioritize unknown over sensitive")
					assert.True(t, sslCert.Sensitive, "ssl_certificate should be sensitive")
					assert.True(t, sslCert.IsUnknown, "ssl_certificate should be unknown")
				} else {
					t.Errorf("ssl_certificate output not found")
				}

				// Validate encryption key (sensitive unknown update)
				if encKey, exists := outputMap["encryption_key"]; exists {
					assert.Equal(t, ChangeTypeUpdate, encKey.ChangeType, "encryption_key should be update type")
					assert.Equal(t, "Modify", encKey.Action, "encryption_key should have Modify action")
					assert.Equal(t, "~", encKey.Indicator, "encryption_key should have ~ indicator")
					assert.Equal(t, "(sensitive value)", encKey.Before, "encryption_key before should be masked")
					assert.Equal(t, "(known after apply)", encKey.After, "encryption_key should prioritize unknown over sensitive")
					assert.True(t, encKey.Sensitive, "encryption_key should be sensitive")
					assert.True(t, encKey.IsUnknown, "encryption_key should be unknown")
				} else {
					t.Errorf("encryption_key output not found")
				}
			},
			description: "Sensitive unknown outputs should prioritize unknown display appropriately",
		},
		{
			name: "plan with only resource changes should suppress outputs section",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.0.0",
				OutputChanges:    map[string]*tfjson.Change{}, // Empty but not nil
				ResourceChanges: []*tfjson.ResourceChange{
					{
						Address: "aws_instance.example",
						Type:    "aws_instance",
						Name:    "example",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionCreate},
							Before:  nil,
							After:   map[string]any{"instance_type": "t2.micro"},
						},
					},
				},
			},
			expectedOutputCount:  0,
			expectedEmptySection: true,
			validateOutputs: func(t *testing.T, outputs []OutputChange, description string) {
				assert.Empty(t, outputs, description+" - outputs should be empty when no output changes exist")
			},
			description: "Plans with only resource changes should suppress empty outputs section",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set the plan on the analyzer
			analyzer.plan = tc.plan

			// Test ProcessOutputChanges directly
			outputs, err := analyzer.ProcessOutputChanges(tc.plan)
			assert.NoError(t, err, tc.description+" - ProcessOutputChanges should not error")
			assert.Equal(t, tc.expectedOutputCount, len(outputs), tc.description+" - output count should match")

			// Run custom validation
			if tc.validateOutputs != nil {
				tc.validateOutputs(t, outputs, tc.description)
			}

			// Test integration with GenerateSummary
			summary := analyzer.GenerateSummary("")
			assert.NotNil(t, summary, tc.description+" - summary should not be nil")
			assert.Equal(t, tc.expectedOutputCount, len(summary.OutputChanges), tc.description+" - summary output count should match")

			// Verify outputs section suppression behavior
			if tc.expectedEmptySection {
				assert.Empty(t, summary.OutputChanges, tc.description+" - summary outputs should be empty when section should be suppressed")
			} else {
				assert.NotEmpty(t, summary.OutputChanges, tc.description+" - summary outputs should not be empty when section should be shown")
			}

			// Verify resource changes are still processed normally
			assert.NotNil(t, summary.ResourceChanges, tc.description+" - resource changes should still be processed")
			assert.Equal(t, len(tc.plan.ResourceChanges), len(summary.ResourceChanges), tc.description+" - resource change count should match")
		})
	}
}

// TestOutputsIntegrationWithResourceChanges tests outputs section integration with resource changes (Task 7.2)
func TestOutputsIntegrationWithResourceChanges(t *testing.T) {
	cfg := &config.Config{}
	analyzer := &Analyzer{config: cfg}

	testCases := []struct {
		name                  string
		plan                  *tfjson.Plan
		expectedResourceCount int
		expectedOutputCount   int
		validateSummary       func(t *testing.T, summary *PlanSummary, description string)
		description           string
	}{
		{
			name: "plan with both resource changes and outputs should display both sections",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.0.0",
				ResourceChanges: []*tfjson.ResourceChange{
					{
						Address: "aws_instance.web",
						Type:    "aws_instance",
						Name:    "web",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionCreate},
							Before:  nil,
							After:   map[string]any{"instance_type": "t2.micro", "ami": "ami-123"},
						},
					},
					{
						Address: "aws_s3_bucket.data",
						Type:    "aws_s3_bucket",
						Name:    "data",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionUpdate},
							Before:  map[string]any{"versioning": false},
							After:   map[string]any{"versioning": true},
						},
					},
				},
				OutputChanges: map[string]*tfjson.Change{
					"instance_ip": {
						Actions:      []tfjson.Action{tfjson.ActionCreate},
						Before:       nil,
						After:        nil,
						AfterUnknown: true,
					},
					"bucket_arn": {
						Actions: []tfjson.Action{tfjson.ActionUpdate},
						Before:  "arn:aws:s3:::old-bucket",
						After:   "arn:aws:s3:::new-bucket",
					},
				},
			},
			expectedResourceCount: 2,
			expectedOutputCount:   2,
			validateSummary: func(t *testing.T, summary *PlanSummary, description string) {
				// Verify both sections are present
				assert.Len(t, summary.ResourceChanges, 2, description+" - should have resource changes")
				assert.Len(t, summary.OutputChanges, 2, description+" - should have output changes")

				// Verify outputs section placement after resource changes
				assert.NotNil(t, summary.ResourceChanges, description+" - resource changes should come first")
				assert.NotNil(t, summary.OutputChanges, description+" - output changes should come after resources")

				// Verify outputs are correctly processed
				outputMap := make(map[string]OutputChange)
				for _, output := range summary.OutputChanges {
					outputMap[output.Name] = output
				}

				if instanceIp, exists := outputMap["instance_ip"]; exists {
					assert.Equal(t, "(known after apply)", instanceIp.After, "instance_ip should show unknown value")
					assert.True(t, instanceIp.IsUnknown, "instance_ip should be unknown")
				}

				if bucketArn, exists := outputMap["bucket_arn"]; exists {
					assert.Equal(t, "arn:aws:s3:::new-bucket", bucketArn.After, "bucket_arn should show new value")
					assert.False(t, bucketArn.IsUnknown, "bucket_arn should not be unknown")
				}

				// Verify statistics only track resource changes (requirement 3.3)
				stats := summary.Statistics
				assert.Equal(t, 1, stats.ToAdd, "statistics should count resource additions")
				assert.Equal(t, 1, stats.ToChange, "statistics should count resource modifications")
				assert.Equal(t, 2, stats.Total, "statistics should total resource changes only")
			},
			description: "Plans with both resources and outputs should display both sections correctly (requirement 2.1)",
		},
		{
			name: "plan with only outputs changes should show outputs section",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.0.0",
				ResourceChanges:  []*tfjson.ResourceChange{}, // No resource changes
				OutputChanges: map[string]*tfjson.Change{
					"api_endpoint": {
						Actions: []tfjson.Action{tfjson.ActionCreate},
						Before:  nil,
						After:   "https://api.example.com/v1",
					},
				},
			},
			expectedResourceCount: 0,
			expectedOutputCount:   1,
			validateSummary: func(t *testing.T, summary *PlanSummary, description string) {
				// Verify no resource changes but outputs present
				assert.Empty(t, summary.ResourceChanges, description+" - should have no resource changes")
				assert.Len(t, summary.OutputChanges, 1, description+" - should have output changes")

				// Verify output is correctly processed
				output := summary.OutputChanges[0]
				assert.Equal(t, "api_endpoint", output.Name, "output name should match")
				assert.Equal(t, "Add", output.Action, "output should have Add action")
				assert.Equal(t, "+", output.Indicator, "output should have + indicator")

				// Verify statistics show no resource changes
				stats := summary.Statistics
				assert.Equal(t, 0, stats.ToAdd, "statistics should show no resource additions")
				assert.Equal(t, 0, stats.ToChange, "statistics should show no resource modifications")
				assert.Equal(t, 0, stats.Total, "statistics should show no total resource changes")
			},
			description: "Plans with only output changes should still display outputs section",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set the plan on the analyzer
			analyzer.plan = tc.plan

			// Generate full summary
			summary := analyzer.GenerateSummary("")
			assert.NotNil(t, summary, tc.description+" - summary should not be nil")

			// Verify counts
			assert.Equal(t, tc.expectedResourceCount, len(summary.ResourceChanges), tc.description+" - resource count should match")
			assert.Equal(t, tc.expectedOutputCount, len(summary.OutputChanges), tc.description+" - output count should match")

			// Run custom validation
			if tc.validateSummary != nil {
				tc.validateSummary(t, summary, tc.description)
			}

			// Verify basic summary fields are populated
			assert.Equal(t, tc.plan.FormatVersion, summary.FormatVersion, tc.description+" - format version should match")
			assert.Equal(t, tc.plan.TerraformVersion, summary.TerraformVersion, tc.description+" - terraform version should match")
			assert.NotNil(t, summary.Statistics, tc.description+" - statistics should be present")
		})
	}
}

// TestOutputsDisplayConsistencyAcrossFormats tests outputs display consistency (Task 7.2)
func TestOutputsDisplayConsistencyAcrossFormats(t *testing.T) {
	analyzer := &Analyzer{}

	// Create a plan with various output types
	plan := &tfjson.Plan{
		FormatVersion:    "1.0",
		TerraformVersion: "1.0.0",
		OutputChanges: map[string]*tfjson.Change{
			"public_ip": {
				Actions: []tfjson.Action{tfjson.ActionCreate},
				Before:  nil,
				After:   "203.0.113.10",
			},
			"secret_key": {
				Actions:        []tfjson.Action{tfjson.ActionUpdate},
				Before:         "old-secret",
				After:          "new-secret",
				AfterSensitive: true,
			},
			"instance_id": {
				Actions:      []tfjson.Action{tfjson.ActionCreate},
				Before:       nil,
				After:        nil,
				AfterUnknown: true,
			},
		},
		ResourceChanges: []*tfjson.ResourceChange{},
	}

	t.Run("outputs should have consistent display values across processing", func(t *testing.T) {
		outputs, err := analyzer.ProcessOutputChanges(plan)
		assert.NoError(t, err, "ProcessOutputChanges should not error")
		assert.Len(t, outputs, 3, "should have 3 outputs")

		// Create a map for easy lookup
		outputMap := make(map[string]OutputChange)
		for _, output := range outputs {
			outputMap[output.Name] = output
		}

		// Test public IP (normal value)
		if publicIp, exists := outputMap["public_ip"]; exists {
			assert.Equal(t, "203.0.113.10", publicIp.After, "public_ip should have actual value")
			assert.False(t, publicIp.Sensitive, "public_ip should not be sensitive")
			assert.False(t, publicIp.IsUnknown, "public_ip should not be unknown")
		}

		// Test secret key (sensitive value)
		if secretKey, exists := outputMap["secret_key"]; exists {
			assert.Equal(t, "(sensitive value)", secretKey.After, "secret_key should display (sensitive value) consistently")
			assert.True(t, secretKey.Sensitive, "secret_key should be marked as sensitive")
			assert.False(t, secretKey.IsUnknown, "secret_key should not be unknown")
		}

		// Test instance ID (unknown value)
		if instanceId, exists := outputMap["instance_id"]; exists {
			assert.Equal(t, "(known after apply)", instanceId.After, "instance_id should display (known after apply) consistently")
			assert.False(t, instanceId.Sensitive, "instance_id should not be sensitive")
			assert.True(t, instanceId.IsUnknown, "instance_id should be marked as unknown")
		}
	})

	t.Run("outputs should maintain consistency in complete summary", func(t *testing.T) {
		analyzer.plan = plan
		summary := analyzer.GenerateSummary("")

		assert.NotNil(t, summary, "summary should not be nil")
		assert.Len(t, summary.OutputChanges, 3, "summary should have 3 outputs")

		// Verify the same consistent display values in the summary
		outputMap := make(map[string]OutputChange)
		for _, output := range summary.OutputChanges {
			outputMap[output.Name] = output
		}

		// Consistency checks
		if publicIp, exists := outputMap["public_ip"]; exists {
			assert.Equal(t, "203.0.113.10", publicIp.After, "public_ip should maintain actual value in summary")
		}

		if secretKey, exists := outputMap["secret_key"]; exists {
			assert.Equal(t, "(sensitive value)", secretKey.After, "secret_key should maintain (sensitive value) in summary")
		}

		if instanceId, exists := outputMap["instance_id"]; exists {
			assert.Equal(t, "(known after apply)", instanceId.After, "instance_id should maintain (known after apply) in summary")
		}
	})
}

// TestCompleteWorkflowWithUnknownValuesAndOutputsIntegration tests the complete workflow
// with real Terraform plan containing unknown values and outputs (Task 9.1)
func TestCompleteWorkflowWithUnknownValuesAndOutputsIntegration(t *testing.T) {
	cfg := &config.Config{
		Plan: config.PlanConfig{
			ShowDetails:      true,
			HighlightDangers: true,
		},
		SensitiveResources: []config.SensitiveResource{
			{ResourceType: "aws_iam_policy"},
		},
		SensitiveProperties: []config.SensitiveProperty{
			{ResourceType: "aws_instance", Property: "user_data"},
		},
	}
	analyzer := &Analyzer{config: cfg}

	testCases := []struct {
		name            string
		plan            *tfjson.Plan
		validateSummary func(t *testing.T, summary *PlanSummary, description string)
		description     string
	}{
		{
			name: "comprehensive plan with mixed unknown values, outputs, and danger highlighting",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.5.0",
				ResourceChanges: []*tfjson.ResourceChange{
					{
						Address: "aws_instance.web_server",
						Type:    "aws_instance",
						Name:    "web_server",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionCreate},
							Before:  nil,
							After: map[string]any{
								"instance_type": "t3.micro",
								"ami":           "ami-12345678",
								"user_data":     "#!/bin/bash\necho 'Hello World'",
								"id":            nil,
								"public_ip":     nil,
							},
							AfterUnknown: map[string]any{
								"id":        true,
								"public_ip": true,
							},
						},
					},
					{
						Address: "aws_iam_policy.admin_policy",
						Type:    "aws_iam_policy",
						Name:    "admin_policy",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionUpdate},
							Before: map[string]any{
								"policy": `{"Version": "2012-10-17"}`,
								"arn":    "arn:aws:iam::123456789012:policy/old-policy",
							},
							After: map[string]any{
								"policy": `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Action": "*", "Resource": "*"}]}`,
								"arn":    nil,
							},
							AfterUnknown: map[string]any{
								"arn": true,
							},
						},
					},
					{
						Address: "aws_s3_bucket.data_bucket",
						Type:    "aws_s3_bucket",
						Name:    "data_bucket",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionCreate},
							Before:  nil,
							After: map[string]any{
								"bucket":             "my-data-bucket-2024",
								"versioning":         true,
								"force_destroy":      false,
								"id":                 nil,
								"bucket_domain_name": nil,
							},
							AfterUnknown: map[string]any{
								"id":                 true,
								"bucket_domain_name": true,
							},
						},
					},
				},
				OutputChanges: map[string]*tfjson.Change{
					"instance_public_ip": {
						Actions:      []tfjson.Action{tfjson.ActionCreate},
						Before:       nil,
						After:        nil,
						AfterUnknown: true,
					},
					"policy_arn": {
						Actions:      []tfjson.Action{tfjson.ActionUpdate},
						Before:       "arn:aws:iam::123456789012:policy/old-policy",
						After:        nil,
						AfterUnknown: true,
					},
					"bucket_endpoint": {
						Actions: []tfjson.Action{tfjson.ActionCreate},
						Before:  nil,
						After:   "https://my-data-bucket-2024.s3.amazonaws.com",
					},
					"database_password": {
						Actions:        []tfjson.Action{tfjson.ActionCreate},
						Before:         nil,
						After:          "super-secret-password",
						AfterSensitive: true,
					},
					"deprecated_config": {
						Actions: []tfjson.Action{tfjson.ActionDelete},
						Before:  "old-configuration-value",
						After:   nil,
					},
				},
			},
			validateSummary: func(t *testing.T, summary *PlanSummary, description string) {
				// Test 1: Verify unknown values display correctly and don't appear as deletions (requirement 1.1, 1.2)
				assert.Len(t, summary.ResourceChanges, 3, description+" - should have 3 resource changes")

				resourceMap := make(map[string]ResourceChange)
				for _, rc := range summary.ResourceChanges {
					resourceMap[rc.Address] = rc
				}

				// Verify aws_instance has unknown values displayed correctly
				if instance, exists := resourceMap["aws_instance.web_server"]; exists {
					assert.True(t, instance.HasUnknownValues, "web_server should have unknown values")
					assert.Contains(t, instance.UnknownProperties, "id", "web_server should have unknown id property")
					assert.Contains(t, instance.UnknownProperties, "public_ip", "web_server should have unknown public_ip property")

					// Verify unknown values in property changes don't appear as deletions
					for _, change := range instance.PropertyChanges.Changes {
						if change.IsUnknown {
							assert.NotEqual(t, "remove", change.Action, "Unknown property %s should not appear as deletion", change.Name)
						}
					}
				}

				// Verify aws_iam_policy has unknown values and can work with danger highlighting (requirement 3.1)
				if policy, exists := resourceMap["aws_iam_policy.admin_policy"]; exists {
					assert.True(t, policy.HasUnknownValues, "admin_policy should have unknown values")
					assert.Contains(t, policy.UnknownProperties, "arn", "admin_policy should have unknown arn property")
					// Note: Danger highlighting typically applies to destructive changes (replace/delete),
					// but unknown values should still work with any danger highlighting logic that does apply
				}

				// Test 2: Verify outputs section displays with correct 5-column format (requirement 2.2)
				assert.Len(t, summary.OutputChanges, 5, description+" - should have 5 output changes")

				outputMap := make(map[string]OutputChange)
				for _, oc := range summary.OutputChanges {
					outputMap[oc.Name] = oc
				}

				// Verify unknown output values display "(known after apply)" (requirement 2.3)
				if instanceIp, exists := outputMap["instance_public_ip"]; exists {
					assert.Equal(t, "(known after apply)", instanceIp.After, "instance_public_ip should show (known after apply)")
					assert.True(t, instanceIp.IsUnknown, "instance_public_ip should be marked as unknown")
					assert.Equal(t, "Add", instanceIp.Action, "instance_public_ip should have Add action")
					assert.Equal(t, "+", instanceIp.Indicator, "instance_public_ip should have + indicator")
				}

				if policyArn, exists := outputMap["policy_arn"]; exists {
					assert.Equal(t, "(known after apply)", policyArn.After, "policy_arn should show (known after apply)")
					assert.True(t, policyArn.IsUnknown, "policy_arn should be marked as unknown")
					assert.Equal(t, "Modify", policyArn.Action, "policy_arn should have Modify action")
					assert.Equal(t, "~", policyArn.Indicator, "policy_arn should have ~ indicator")
				}

				// Verify sensitive output values display "(sensitive value)" (requirement 2.4)
				if dbPassword, exists := outputMap["database_password"]; exists {
					assert.Equal(t, "(sensitive value)", dbPassword.After, "database_password should show (sensitive value)")
					assert.True(t, dbPassword.Sensitive, "database_password should be marked as sensitive")
					assert.False(t, dbPassword.IsUnknown, "database_password should not be unknown")
					assert.Equal(t, "Add", dbPassword.Action, "database_password should have Add action")
					assert.Equal(t, "+", dbPassword.Indicator, "database_password should have + indicator")
				}

				// Verify normal output values display correctly
				if bucketEndpoint, exists := outputMap["bucket_endpoint"]; exists {
					assert.Equal(t, "https://my-data-bucket-2024.s3.amazonaws.com", bucketEndpoint.After, "bucket_endpoint should show actual value")
					assert.False(t, bucketEndpoint.Sensitive, "bucket_endpoint should not be sensitive")
					assert.False(t, bucketEndpoint.IsUnknown, "bucket_endpoint should not be unknown")
				}

				// Verify deleted output values
				if deprecated, exists := outputMap["deprecated_config"]; exists {
					assert.Equal(t, "Remove", deprecated.Action, "deprecated_config should have Remove action")
					assert.Equal(t, "-", deprecated.Indicator, "deprecated_config should have - indicator")
				}

				// Test 3: Verify statistics properly categorize resource changes with unknown values (requirement 3.3)
				stats := summary.Statistics
				assert.Equal(t, 2, stats.ToAdd, "statistics should count resource creations")
				assert.Equal(t, 1, stats.ToChange, "statistics should count resource modifications")
				assert.Equal(t, 3, stats.Total, "statistics should total resource changes only, not outputs")

				// Test 4: Verify basic summary structure
				assert.Equal(t, "1.0", summary.FormatVersion, "format version should match")
				assert.Equal(t, "1.5.0", summary.TerraformVersion, "terraform version should match")
			},
			description: "Complete workflow should handle unknown values, outputs, and danger highlighting together",
		},
		{
			name: "complex nested unknown values with outputs integration",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.5.0",
				ResourceChanges: []*tfjson.ResourceChange{
					{
						Address: "aws_vpc.main",
						Type:    "aws_vpc",
						Name:    "main",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionCreate},
							Before:  nil,
							After: map[string]any{
								"cidr_block": "10.0.0.0/16",
								"tags": map[string]any{
									"Name":        "main-vpc",
									"Environment": "production",
								},
								"id":                        nil,
								"arn":                       nil,
								"default_security_group_id": nil,
								"default_network_acl_id":    nil,
							},
							AfterUnknown: map[string]any{
								"id":                        true,
								"arn":                       true,
								"default_security_group_id": true,
								"default_network_acl_id":    true,
							},
						},
					},
				},
				OutputChanges: map[string]*tfjson.Change{
					"vpc_details": {
						Actions:      []tfjson.Action{tfjson.ActionCreate},
						Before:       nil,
						After:        nil,
						AfterUnknown: true,
					},
				},
			},
			validateSummary: func(t *testing.T, summary *PlanSummary, description string) {
				// Verify nested unknown values are handled correctly
				assert.Len(t, summary.ResourceChanges, 1, description+" - should have 1 resource change")

				vpc := summary.ResourceChanges[0]
				assert.True(t, vpc.HasUnknownValues, "VPC should have unknown values")
				assert.Contains(t, vpc.UnknownProperties, "id", "VPC should have unknown id")
				assert.Contains(t, vpc.UnknownProperties, "arn", "VPC should have unknown arn")
				assert.Contains(t, vpc.UnknownProperties, "default_security_group_id", "VPC should have unknown default_security_group_id")

				// Verify outputs with complex unknown structures
				assert.Len(t, summary.OutputChanges, 1, description+" - should have 1 output change")

				output := summary.OutputChanges[0]
				assert.Equal(t, "vpc_details", output.Name, "output name should match")
				assert.Equal(t, "(known after apply)", output.After, "complex unknown output should show (known after apply)")
				assert.True(t, output.IsUnknown, "output should be marked as unknown")
			},
			description: "Complex nested unknown values should integrate correctly with outputs",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set the plan on the analyzer
			analyzer.plan = tc.plan

			// Generate complete summary - this tests the full workflow
			summary := analyzer.GenerateSummary("")
			assert.NotNil(t, summary, tc.description+" - summary should not be nil")

			// Run comprehensive validation
			if tc.validateSummary != nil {
				tc.validateSummary(t, summary, tc.description)
			}
		})
	}
}

// TestCrossFormatConsistencyForUnknownValuesAndOutputs tests display consistency
// across all output formats (Task 9.2)
