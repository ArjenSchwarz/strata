package plan

import (
	"os"
	"testing"
)

// skipIfIntegrationTestsDisabled skips the test if integration tests are not enabled
// Integration tests should be run with the INTEGRATION environment variable set
func skipIfIntegrationTestsDisabled(t *testing.T) {
	t.Helper()

	if os.Getenv("INTEGRATION") == "" {
		t.Skip("Skipping integration test. Set INTEGRATION=1 to run integration tests.")
	}
}
