# Benchmarking Guide

This document explains how to run and analyze benchmarks for the Strata project.

## Running Benchmarks

### Basic Benchmark Execution

Run all benchmarks with memory profiling:

```bash
go test -bench=. -benchmem ./lib/plan/
```

Run specific benchmark patterns:

```bash
# Run only analysis benchmarks
go test -bench=BenchmarkAnalysis -benchmem ./lib/plan/

# Run only formatting benchmarks  
go test -bench=BenchmarkFormatting -benchmem ./lib/plan/

# Run property analysis benchmarks
go test -bench=BenchmarkPropertyAnalysis -benchmem ./lib/plan/
```

### Performance Testing

Run performance target tests:

```bash
# Run performance validation tests
go test -run=TestPerformanceTargets ./lib/plan/

# Run memory usage tests
go test -run=TestMemoryUsage ./lib/plan/
```

## Statistical Analysis with benchstat

### Installation

Install the benchstat tool:

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

### Basic Usage

1. Run benchmarks multiple times and save results:

```bash
# Run benchmarks 10 times and save to file
go test -bench=. -benchmem -count=10 ./lib/plan/ > bench_results.txt
```

2. Analyze results with benchstat:

```bash
# Show statistical summary
benchstat bench_results.txt
```

### Comparing Performance

To compare performance between different versions or changes:

1. Create baseline measurements:

```bash
# Before changes
go test -bench=. -benchmem -count=10 ./lib/plan/ > bench_before.txt
```

2. Make your changes, then create new measurements:

```bash
# After changes
go test -bench=. -benchmem -count=10 ./lib/plan/ > bench_after.txt
```

3. Compare with benchstat:

```bash
# Compare before and after
benchstat bench_before.txt bench_after.txt
```

### Interpreting Results

benchstat provides:
- **Mean execution time** with confidence intervals
- **Memory allocation statistics** (bytes allocated, allocations per operation)
- **Performance change analysis** when comparing results
- **Statistical significance** indicators

Example output:
```
name                           old time/op    new time/op    delta
BenchmarkAnalysis_SmallPlan-8    1.23ms ± 2%    1.15ms ± 3%   -6.50%  (p=0.000 n=10+10)

name                           old alloc/op   new alloc/op   delta
BenchmarkAnalysis_SmallPlan-8     245kB ± 0%     240kB ± 0%   -2.04%  (p=0.000 n=10+10)
```

### Best Practices

1. **Run multiple iterations**: Use `-count=10` or higher for reliable statistics
2. **Stable environment**: Run benchmarks on a quiet system for consistent results  
3. **Baseline comparisons**: Always compare against a known baseline
4. **Statistical significance**: Pay attention to p-values and confidence intervals
5. **Memory profiling**: Always include `-benchmem` for complete performance picture

## Available Benchmarks

### Analysis Benchmarks
- `BenchmarkAnalysis_SmallPlan`: Tests with 10 resources
- `BenchmarkAnalysis_MediumPlan`: Tests with 100 resources  
- `BenchmarkAnalysis_LargePlan`: Tests with 1000 resources

### Formatting Benchmarks
- `BenchmarkFormatting_ProgressiveDisclosure`: Tests collapsible section formatting
- `BenchmarkFormatting_GroupedSections`: Tests provider grouping performance

### Property Analysis Benchmarks
- `BenchmarkPropertyAnalysis`: Tests various property data sizes and counts

### Performance Tests
- `TestPerformanceTargets`: Validates execution time requirements
- `TestMemoryUsage`: Validates memory consumption limits
- `TestCollapsibleFormatterPerformance`: Compares formatting performance