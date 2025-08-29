package plan

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var updateGolden = flag.Bool("update-golden", false, "Update golden files")

// GoldenFileHelper provides utilities for golden file testing
type GoldenFileHelper struct {
	testdataDir string
}

// NewGoldenFileHelper creates a new golden file helper
func NewGoldenFileHelper(testdataDir string) *GoldenFileHelper {
	return &GoldenFileHelper{
		testdataDir: testdataDir,
	}
}

// CompareOrUpdateGolden compares output with golden file or updates it if -update-golden flag is set
func (g *GoldenFileHelper) CompareOrUpdateGolden(t *testing.T, testName string, got []byte) {
	t.Helper()

	goldenFile := filepath.Join(g.testdataDir, "golden", testName+".golden")

	if *updateGolden {
		// Ensure the golden directory exists
		if err := os.MkdirAll(filepath.Dir(goldenFile), 0755); err != nil {
			t.Fatalf("Failed to create golden directory: %v", err)
		}

		// Update the golden file
		if err := os.WriteFile(goldenFile, got, 0644); err != nil {
			t.Fatalf("Failed to update golden file %s: %v", goldenFile, err)
		}
		t.Logf("Updated golden file: %s", goldenFile)
		return
	}

	// Read expected content from golden file
	want, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("Failed to read golden file %s: %v", goldenFile, err)
	}

	// Compare content
	if string(got) != string(want) {
		t.Errorf("Output doesn't match golden file %s\nGot:\n%s\n\nWant:\n%s\n\nRun with -update-golden to update",
			goldenFile, string(got), string(want))
	}
}

// LoadGoldenFile loads content from a golden file
func (g *GoldenFileHelper) LoadGoldenFile(t *testing.T, testName string) []byte {
	t.Helper()

	goldenFile := filepath.Join(g.testdataDir, "golden", testName+".golden")
	content, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("Failed to read golden file %s: %v", goldenFile, err)
	}

	return content
}

// SaveGoldenFile saves content to a golden file (useful for initial creation)
func (g *GoldenFileHelper) SaveGoldenFile(t *testing.T, testName string, content []byte) {
	t.Helper()

	goldenFile := filepath.Join(g.testdataDir, "golden", testName+".golden")

	// Ensure the golden directory exists
	if err := os.MkdirAll(filepath.Dir(goldenFile), 0755); err != nil {
		t.Fatalf("Failed to create golden directory: %v", err)
	}

	if err := os.WriteFile(goldenFile, content, 0644); err != nil {
		t.Fatalf("Failed to save golden file %s: %v", goldenFile, err)
	}

	t.Logf("Saved golden file: %s", goldenFile)
}
