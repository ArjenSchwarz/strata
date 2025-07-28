package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ArjenSchwarz/strata/config"
)

// BenchmarkAnalysis_SmallPlan benchmarks analysis with a small plan (10 resources)
func BenchmarkAnalysis_SmallPlan(b *testing.B) {
	planPath := createBenchmarkPlan("small_benchmark_plan.json", 10)
	defer os.RemoveAll(filepath.Dir(planPath))

	cfg := getBenchmarkConfig()
	analyzer := NewAnalyzer(nil, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		summary := analyzer.GenerateSummary(planPath)
		if summary == nil {
			b.Fatal("Expected non-nil summary")
		}
	}
}

// BenchmarkAnalysis_MediumPlan benchmarks analysis with a medium plan (100 resources)
func BenchmarkAnalysis_MediumPlan(b *testing.B) {
	planPath := createBenchmarkPlan("medium_benchmark_plan.json", 100)
	defer os.RemoveAll(filepath.Dir(planPath))

	cfg := getBenchmarkConfig()
	analyzer := NewAnalyzer(nil, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		summary := analyzer.GenerateSummary(planPath)
		if summary == nil {
			b.Fatal("Expected non-nil summary")
		}
	}
}

// BenchmarkAnalysis_LargePlan benchmarks analysis with a large plan (1000 resources)
func BenchmarkAnalysis_LargePlan(b *testing.B) {
	planPath := createBenchmarkPlan("large_benchmark_plan.json", 1000)
	defer os.RemoveAll(filepath.Dir(planPath))

	cfg := getBenchmarkConfig()
	analyzer := NewAnalyzer(nil, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		summary := analyzer.GenerateSummary(planPath)
		if summary == nil {
			b.Fatal("Expected non-nil summary")
		}
	}
}

// BenchmarkFormatting_ProgressiveDisclosure benchmarks the progressive disclosure formatter
func BenchmarkFormatting_ProgressiveDisclosure(b *testing.B) {
	planPath := createBenchmarkPlan("format_benchmark_plan.json", 100)
	defer os.RemoveAll(filepath.Dir(planPath))

	cfg := getBenchmarkConfig()
	cfg.Plan.ExpandableSections.Enabled = true
	analyzer := NewAnalyzer(nil, cfg)
	formatter := NewFormatter(cfg)

	summary := analyzer.GenerateSummary(planPath)
	if summary == nil {
		b.Fatal("Failed to generate summary for benchmark")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc, err := formatter.formatResourceChangesWithProgressiveDisclosure(summary)
		if err != nil {
			b.Fatalf("Formatting failed: %v", err)
		}
		if doc == nil {
			b.Fatal("Expected non-nil document")
		}
	}
}

// BenchmarkFormatting_GroupedSections benchmarks the grouped sections formatter
func BenchmarkFormatting_GroupedSections(b *testing.B) {
	planPath := createBenchmarkPlan("grouped_benchmark_plan.json", 200)
	defer os.RemoveAll(filepath.Dir(planPath))

	cfg := getBenchmarkConfig()
	cfg.Plan.ExpandableSections.Enabled = true
	cfg.Plan.Grouping.Enabled = true
	cfg.Plan.Grouping.Threshold = 50
	analyzer := NewAnalyzer(nil, cfg)
	formatter := NewFormatter(cfg)

	summary := analyzer.GenerateSummary(planPath)
	if summary == nil {
		b.Fatal("Failed to generate summary for benchmark")
	}

	groups := analyzer.groupByProvider(summary.ResourceChanges)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc, err := formatter.formatGroupedWithCollapsibleSections(summary, groups)
		if err != nil {
			b.Fatalf("Grouped formatting failed: %v", err)
		}
		if doc == nil {
			b.Fatal("Expected non-nil document")
		}
	}
}

// BenchmarkPropertyAnalysis benchmarks property change analysis with various data sizes
func BenchmarkPropertyAnalysis(b *testing.B) {
	tests := []struct {
		name       string
		properties int
		size       int // Size of each property value in characters
	}{
		{"SmallProperties", 10, 100},
		{"MediumProperties", 50, 500},
		{"LargeProperties", 100, 1000},
		{"ManySmallProperties", 200, 50},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			planPath := createPropertyBenchmarkPlan(fmt.Sprintf("prop_%s.json", tt.name), tt.properties, tt.size)
			defer os.RemoveAll(filepath.Dir(planPath))

			cfg := getBenchmarkConfig()
			analyzer := NewAnalyzer(nil, cfg)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				summary := analyzer.GenerateSummary(planPath)
				if summary == nil {
					b.Fatal("Expected non-nil summary")
				}
			}
		})
	}
}

// TestPerformanceTargets tests that performance targets are met
func TestPerformanceTargets(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	tests := []struct {
		name        string
		resources   int
		maxDuration time.Duration
		description string
	}{
		{
			name:        "10 resources under 500ms",
			resources:   10,
			maxDuration: 500 * time.Millisecond,
			description: "Small plans should process quickly",
		},
		{
			name:        "100 resources under 1s",
			resources:   100,
			maxDuration: 1 * time.Second,
			description: "Medium plans should process within reasonable time",
		},
		{
			name:        "1000 resources under 10s",
			resources:   1000,
			maxDuration: 10 * time.Second,
			description: "Large plans should still be manageable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			planPath := createBenchmarkPlan(fmt.Sprintf("perf_%d.json", tt.resources), tt.resources)
			defer os.RemoveAll(filepath.Dir(planPath))

			cfg := getBenchmarkConfig()
			analyzer := NewAnalyzer(nil, cfg)

			start := time.Now()
			summary := analyzer.GenerateSummary(planPath)
			duration := time.Since(start)

			if summary == nil {
				t.Fatal("Expected non-nil summary")
			}

			if duration > tt.maxDuration {
				t.Errorf("%s: took %v, expected under %v", tt.description, duration, tt.maxDuration)
			}

			t.Logf("Processed %d resources in %v (target: <%v)", tt.resources, duration, tt.maxDuration)
		})
	}
}

// TestMemoryUsage tests that memory usage stays within reasonable bounds
func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory tests in short mode")
	}

	// Get baseline memory
	runtime.GC()
	var baselineMemStats runtime.MemStats
	runtime.ReadMemStats(&baselineMemStats)

	// Process a large plan
	planPath := createBenchmarkPlan("memory_test_plan.json", 1000)
	defer os.RemoveAll(filepath.Dir(planPath))

	cfg := getBenchmarkConfig()
	analyzer := NewAnalyzer(nil, cfg)

	summary := analyzer.GenerateSummary(planPath)
	if summary == nil {
		t.Fatal("Expected non-nil summary")
	}

	// Format the summary
	formatter := NewFormatter(cfg)
	doc, err := formatter.formatResourceChangesWithProgressiveDisclosure(summary)
	if err != nil {
		t.Fatalf("Formatting failed: %v", err)
	}
	if doc == nil {
		t.Fatal("Expected non-nil document")
	}

	// Check memory usage
	runtime.GC()
	var finalMemStats runtime.MemStats
	runtime.ReadMemStats(&finalMemStats)

	// Use TotalAlloc to measure total memory allocated during the test
	memoryUsed := finalMemStats.TotalAlloc - baselineMemStats.TotalAlloc
	maxMemoryAllowed := uint64(500 * 1024 * 1024) // 500MB

	if memoryUsed > maxMemoryAllowed {
		t.Errorf("Memory usage %d bytes exceeds limit %d bytes", memoryUsed, maxMemoryAllowed)
	}

	t.Logf("Memory used: %d bytes (limit: %d bytes)", memoryUsed, maxMemoryAllowed)
}

// TestPerformanceLimitsEnforcement tests that performance limits are actually enforced
func TestPerformanceLimitsEnforcement(t *testing.T) {
	// Create a plan with many large properties that should trigger limits
	planPath := createPropertyBenchmarkPlan("limits_test.json", 200, 2000)
	defer os.RemoveAll(filepath.Dir(planPath))

	// Set very restrictive limits
	cfg := getBenchmarkConfig()
	cfg.Plan.PerformanceLimits.MaxPropertiesPerResource = 10
	cfg.Plan.PerformanceLimits.MaxPropertySize = 100
	cfg.Plan.PerformanceLimits.MaxTotalMemory = 50 * 1024 // 50KB

	analyzer := NewAnalyzer(nil, cfg)

	start := time.Now()
	summary := analyzer.GenerateSummary(planPath)
	duration := time.Since(start)

	if summary == nil {
		t.Fatal("Expected non-nil summary")
	}

	// Should complete quickly due to limits
	if duration > 2*time.Second {
		t.Errorf("Analysis took too long (%v) despite performance limits", duration)
	}

	// Should have processed resources but possibly truncated properties
	if len(summary.ResourceChanges) == 0 {
		t.Error("Expected at least some resources to be processed")
	}

	t.Logf("Processed with limits in %v", duration)
}

// TestCollapsibleFormatterPerformance compares performance with and without collapsible formatters
func TestCollapsibleFormatterPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance comparison tests in short mode")
	}

	planPath := createBenchmarkPlan("formatter_perf.json", 100)
	defer os.RemoveAll(filepath.Dir(planPath))

	// Test with collapsible sections disabled
	cfg1 := getBenchmarkConfig()
	cfg1.Plan.ExpandableSections.Enabled = false
	analyzer1 := NewAnalyzer(nil, cfg1)
	formatter1 := NewFormatter(cfg1)

	summary1 := analyzer1.GenerateSummary(planPath)
	if summary1 == nil {
		t.Fatal("Expected non-nil summary for first formatter test")
	}

	start1 := time.Now()
	_, err1 := formatter1.formatResourceChangesWithProgressiveDisclosure(summary1)
	duration1 := time.Since(start1)
	if err1 != nil {
		t.Fatalf("Simple formatting failed: %v", err1)
	}

	// Test with collapsible sections enabled
	cfg2 := getBenchmarkConfig()
	cfg2.Plan.ExpandableSections.Enabled = true
	analyzer2 := NewAnalyzer(nil, cfg2)
	formatter2 := NewFormatter(cfg2)

	summary2 := analyzer2.GenerateSummary(planPath)
	if summary2 == nil {
		t.Fatal("Failed to generate summary")
	}

	start2 := time.Now()
	_, err2 := formatter2.formatResourceChangesWithProgressiveDisclosure(summary2)
	duration2 := time.Since(start2)
	if err2 != nil {
		t.Fatalf("Collapsible formatting failed: %v", err2)
	}

	// Collapsible formatting should not be significantly slower
	overhead := float64(duration2) / float64(duration1)
	maxOverhead := 6.0 // Allow up to 6x slower (increased due to multi-table rendering complexity)

	if overhead > maxOverhead {
		t.Errorf("Collapsible formatting is %.1fx slower than simple formatting (max allowed: %.1fx)", overhead, maxOverhead)
	}

	t.Logf("Simple: %v, Collapsible: %v (overhead: %.1fx)", duration1, duration2, overhead)
}

// Helper function to create benchmark plans with specified number of resources
func createBenchmarkPlan(filename string, resourceCount int) string {
	// Create temp directory for test files
	tempDir, err := os.MkdirTemp("", "strata-test-*")
	if err != nil {
		panic(fmt.Sprintf("Failed to create temp directory: %v", err))
	}
	planPath := filepath.Join(tempDir, filename)

	// Create plan JSON with multiple providers for realistic testing
	var resources []string
	providers := []string{"aws", "azurerm", "google", "kubernetes"}

	for i := 0; i < resourceCount; i++ {
		provider := providers[i%len(providers)]
		resourceType := fmt.Sprintf("%s_instance", provider)
		if provider == "kubernetes" {
			resourceType = "kubernetes_deployment"
		}

		resource := fmt.Sprintf(`{
			"address": "%s.resource_%d",
			"mode": "managed",
			"type": "%s",
			"name": "resource_%d",
			"provider_name": "registry.terraform.io/hashicorp/%s",
			"change": {
				"actions": ["%s"],
				"before": %s,
				"after": {
					"name": "resource_%d",
					"type": "benchmark",
					"tags": {"Environment": "test", "Resource": "%d"}
				},
				"after_unknown": {"id": true},
				"before_sensitive": false,
				"after_sensitive": {"tags": {}}
			}
		}`,
			resourceType, i, resourceType, i, provider,
			[]string{"create", "update", "delete"}[i%3],
			func() string {
				if i%3 == 0 {
					return "null"
				}
				return fmt.Sprintf(`{"name": "old_resource_%d", "type": "benchmark"}`, i)
			}(),
			i, i)

		resources = append(resources, resource)
	}

	planJSON := fmt.Sprintf(`{
		"format_version": "1.2",
		"terraform_version": "1.8.5",
		"variables": {},
		"planned_values": {
			"root_module": {
				"resources": []
			}
		},
		"resource_changes": [%s],
		"output_changes": {},
		"prior_state": {
			"format_version": "1.0",
			"terraform_version": "1.8.5",
			"values": {
				"root_module": {}
			}
		},
		"configuration": {
			"provider_config": {},
			"root_module": {}
		}
	}`, strings.Join(resources, ","))

	err = os.WriteFile(planPath, []byte(planJSON), 0644)
	if err != nil {
		panic(fmt.Sprintf("Failed to create benchmark plan: %v", err))
	}

	return planPath
}

// Helper function to create benchmark plans with many properties
func createPropertyBenchmarkPlan(filename string, propertyCount, propertySize int) string {
	// Create temp directory for test files
	tempDir, err := os.MkdirTemp("", "strata-prop-test-*")
	if err != nil {
		panic(fmt.Sprintf("Failed to create temp directory: %v", err))
	}
	planPath := filepath.Join(tempDir, filename)

	// Create before and after objects with many properties
	var beforeProps, afterProps []string
	for i := 0; i < propertyCount; i++ {
		value := strings.Repeat("x", propertySize)
		beforeProps = append(beforeProps, fmt.Sprintf(`"property_%d": "%s"`, i, value))
		afterProps = append(afterProps, fmt.Sprintf(`"property_%d": "%s_updated"`, i, value))
	}

	planJSON := fmt.Sprintf(`{
		"format_version": "1.2",
		"terraform_version": "1.8.5",
		"variables": {},
		"planned_values": {
			"root_module": {
				"resources": []
			}
		},
		"resource_changes": [
			{
				"address": "aws_instance.property_heavy",
				"mode": "managed",
				"type": "aws_instance",
				"name": "property_heavy",
				"provider_name": "registry.terraform.io/hashicorp/aws",
				"change": {
					"actions": ["update"],
					"before": {%s},
					"after": {%s},
					"after_unknown": {},
					"before_sensitive": {},
					"after_sensitive": {}
				}
			}
		],
		"output_changes": {},
		"prior_state": {
			"format_version": "1.0",
			"terraform_version": "1.8.5",
			"values": {
				"root_module": {}
			}
		},
		"configuration": {
			"provider_config": {},
			"root_module": {}
		}
	}`, strings.Join(beforeProps, ","), strings.Join(afterProps, ","))

	err = os.WriteFile(planPath, []byte(planJSON), 0644)
	if err != nil {
		panic(fmt.Sprintf("Failed to create property benchmark plan: %v", err))
	}

	return planPath
}

// Helper function to create benchmark configuration for performance tests
func getBenchmarkConfig() *config.Config {
	return &config.Config{
		ExpandAll: false,
		Plan: config.PlanConfig{
			ExpandableSections: config.ExpandableSectionsConfig{
				Enabled:             true,
				AutoExpandDangerous: true,
				ShowDependencies:    true,
			},
			Grouping: config.GroupingConfig{
				Enabled:   true,
				Threshold: 10,
			},
			PerformanceLimits: config.PerformanceLimitsConfig{
				MaxPropertiesPerResource: 100,
				MaxPropertySize:          1024 * 1024,       // 1MB
				MaxTotalMemory:           100 * 1024 * 1024, // 100MB
				MaxDependencyDepth:       10,
			},
		},
		SensitiveResources: []config.SensitiveResource{
			{ResourceType: "aws_db_instance"},
			{ResourceType: "aws_rds_db_instance"},
		},
		SensitiveProperties: []config.SensitiveProperty{
			{ResourceType: "aws_instance", Property: "user_data"},
		},
	}
}
