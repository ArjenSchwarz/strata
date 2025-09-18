package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/ArjenSchwarz/strata/config"
)

// BenchmarkAnalysis_SmallPlan benchmarks analysis with a small plan (10 resources)
func BenchmarkAnalysis_SmallPlan(b *testing.B) {
	planPath := createBenchmarkPlan("small_benchmark_plan.json", 10)
	b.Cleanup(func() { os.RemoveAll(filepath.Dir(planPath)) })

	cfg := getBenchmarkConfig()
	analyzer := NewAnalyzer(nil, cfg)

	for b.Loop() {
		summary := analyzer.GenerateSummary(planPath)
		if summary == nil {
			b.Fatal("Expected non-nil summary")
		}
	}
}

// BenchmarkAnalysis_MediumPlan benchmarks analysis with a medium plan (100 resources)
func BenchmarkAnalysis_MediumPlan(b *testing.B) {
	planPath := createBenchmarkPlan("medium_benchmark_plan.json", 100)
	b.Cleanup(func() { os.RemoveAll(filepath.Dir(planPath)) })

	cfg := getBenchmarkConfig()
	analyzer := NewAnalyzer(nil, cfg)

	for b.Loop() {
		summary := analyzer.GenerateSummary(planPath)
		if summary == nil {
			b.Fatal("Expected non-nil summary")
		}
	}
}

// BenchmarkAnalysis_LargePlan benchmarks analysis with a large plan (1000 resources)
func BenchmarkAnalysis_LargePlan(b *testing.B) {
	planPath := createBenchmarkPlan("large_benchmark_plan.json", 1000)
	b.Cleanup(func() { os.RemoveAll(filepath.Dir(planPath)) })

	cfg := getBenchmarkConfig()
	analyzer := NewAnalyzer(nil, cfg)

	for b.Loop() {
		summary := analyzer.GenerateSummary(planPath)
		if summary == nil {
			b.Fatal("Expected non-nil summary")
		}
	}
}

// BenchmarkFormatting_ProgressiveDisclosure benchmarks the progressive disclosure formatter
func BenchmarkFormatting_ProgressiveDisclosure(b *testing.B) {
	planPath := createBenchmarkPlan("format_benchmark_plan.json", 100)
	b.Cleanup(func() { os.RemoveAll(filepath.Dir(planPath)) })

	cfg := getBenchmarkConfig()
	cfg.Plan.ExpandableSections.Enabled = true
	analyzer := NewAnalyzer(nil, cfg)
	formatter := NewFormatter(cfg)

	summary := analyzer.GenerateSummary(planPath)
	if summary == nil {
		b.Fatal("Failed to generate summary for benchmark")
	}

	for b.Loop() {
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
	b.Cleanup(func() { os.RemoveAll(filepath.Dir(planPath)) })

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

	for b.Loop() {
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
			b.Cleanup(func() { os.RemoveAll(filepath.Dir(planPath)) })

			cfg := getBenchmarkConfig()
			analyzer := NewAnalyzer(nil, cfg)

			b.ResetTimer()
			for b.Loop() {
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
			t.Cleanup(func() { os.RemoveAll(filepath.Dir(planPath)) })

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
	t.Cleanup(func() { os.RemoveAll(filepath.Dir(planPath)) })

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
	t.Cleanup(func() { os.RemoveAll(filepath.Dir(planPath)) })

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
	t.Cleanup(func() { os.RemoveAll(filepath.Dir(planPath)) })

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
	builder := CreateMultiProviderPlan(resourceCount)
	planPath, err := builder.SaveToTempFile(filename)
	if err != nil {
		panic(fmt.Sprintf("Failed to create benchmark plan: %v", err))
	}

	return planPath
}

// Helper function to create benchmark plans with many properties
func createPropertyBenchmarkPlan(filename string, propertyCount, propertySize int) string {
	builder := CreatePropertyBenchmarkPlan(propertyCount, propertySize)
	planPath, err := builder.SaveToTempFile(filename)
	if err != nil {
		panic(fmt.Sprintf("Failed to create property benchmark plan: %v", err))
	}

	return planPath
}

// BenchmarkSortResourceTableData benchmarks the data pipeline sorting function with various data sizes
// This directly tests the performance improvement from replacing ActionSortTransformer
func BenchmarkSortResourceTableData(b *testing.B) {
	tests := []struct {
		name string
		size int
	}{
		{"Small_10", 10},
		{"Medium_100", 100},
		{"Large_1000", 1000},
		{"ExtraLarge_5000", 5000},
		{"Huge_10000", 10000},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			// Create test data with mixed priorities and danger levels
			tableData := createBenchmarkTableData(tt.size)

			b.ResetTimer()
			b.ReportAllocs() // Track memory allocations

			for b.Loop() {
				// Make a copy for each iteration to avoid sorting already sorted data
				testData := make([]map[string]any, len(tableData))
				copy(testData, tableData)

				sortResourceTableData(testData)
			}
		})
	}
}

// BenchmarkApplyDecorations benchmarks the decoration function with various data sizes
func BenchmarkApplyDecorations(b *testing.B) {
	tests := []struct {
		name string
		size int
	}{
		{"Small_10", 10},
		{"Medium_100", 100},
		{"Large_1000", 1000},
		{"ExtraLarge_5000", 5000},
		{"Huge_10000", 10000},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			// Create test data for decoration
			tableData := createBenchmarkTableData(tt.size)

			b.ResetTimer()
			b.ReportAllocs() // Track memory allocations

			for b.Loop() {
				// Make a copy for each iteration
				testData := make([]map[string]any, len(tableData))
				for i, row := range tableData {
					testData[i] = make(map[string]any)
					for k, v := range row {
						testData[i][k] = v
					}
				}

				applyDecorations(testData)
			}
		})
	}
}

// BenchmarkDataPipelineSortingComplete benchmarks the complete data pipeline sorting process
// This combines sorting and decoration to measure the full data-level transformation performance
func BenchmarkDataPipelineSortingComplete(b *testing.B) {
	tests := []struct {
		name string
		size int
	}{
		{"Complete_Small_10", 10},
		{"Complete_Medium_100", 100},
		{"Complete_Large_1000", 1000},
		{"Complete_ExtraLarge_5000", 5000},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			// Create test data
			originalData := createBenchmarkTableData(tt.size)

			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				// Make a deep copy for each iteration
				testData := make([]map[string]any, len(originalData))
				for i, row := range originalData {
					testData[i] = make(map[string]any)
					for k, v := range row {
						testData[i][k] = v
					}
				}

				// Step 1: Sort the raw data
				sortResourceTableData(testData)

				// Step 2: Apply decorations after sorting
				applyDecorations(testData)
			}
		})
	}
}

// BenchmarkDataPipelineSortingWorstCase benchmarks sorting with worst-case data distribution
// This tests performance when data is reverse sorted and all items have different priorities
func BenchmarkDataPipelineSortingWorstCase(b *testing.B) {
	sizes := []int{100, 1000, 5000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("WorstCase_%d", size), func(b *testing.B) {
			// Create worst-case data: reverse sorted with maximum variation
			tableData := createWorstCaseTableData(size)

			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				// Make a copy for each iteration
				testData := make([]map[string]any, len(tableData))
				copy(testData, tableData)

				sortResourceTableData(testData)
			}
		})
	}
}

// createBenchmarkTableData creates test data with realistic distribution of priorities and danger levels
// Matches typical Terraform plan patterns with mixed resource types and changes
func createBenchmarkTableData(size int) []map[string]any {
	actions := []string{"Remove", "Replace", "Modify", "Add"}

	data := make([]map[string]any, size)

	for i := 0; i < size; i++ {
		isDangerous := (i%5 == 0) // Every 5th resource is dangerous (20%)
		actionIndex := i % len(actions)

		data[i] = map[string]any{
			"ActionType":  actions[actionIndex],
			"IsDangerous": isDangerous,
			"Resource":    fmt.Sprintf("aws_instance.resource_%04d", i),
			"Type":        "aws_instance",
			"ID":          fmt.Sprintf("i-%08d", i),
			"Replacement": "N/A",
			"Module":      "root",
			"Danger":      "",
		}
	}

	return data
}

// createWorstCaseTableData creates data in reverse-sorted order to test worst-case performance
// This simulates data that would require maximum sorting effort
func createWorstCaseTableData(size int) []map[string]any {
	actions := []string{"Add", "Modify", "Replace", "Remove"} // Reverse priority order

	data := make([]map[string]any, size)

	for i := 0; i < size; i++ {
		// Create reverse alphabetical order for resources
		resourceNum := size - i - 1

		// Alternate danger status to create maximum sorting complexity
		isDangerous := (i%2 == 1)

		data[i] = map[string]any{
			"ActionType":  actions[i%len(actions)],
			"IsDangerous": isDangerous,
			"Resource":    fmt.Sprintf("zzz_resource_%04d", resourceNum),
			"Type":        "aws_instance",
			"ID":          fmt.Sprintf("i-%08d", resourceNum),
			"Replacement": "N/A",
			"Module":      "root",
			"Danger":      "",
		}
	}

	return data
}

// Helper function to create benchmark configuration for performance tests
func getBenchmarkConfig() *config.Config {
	return &config.Config{
		ExpandAll: false,
		Plan: config.PlanConfig{
			ExpandableSections: config.ExpandableSectionsConfig{
				Enabled:             true,
				AutoExpandDangerous: true,
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
