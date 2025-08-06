# Testing Guide for Labours-Go

This document provides comprehensive information about the testing infrastructure for the labours-go project.

## Overview

The labours-go project includes a complete testing suite with:

- Unit tests for all components
- Integration tests with real data
- Visual regression tests for chart consistency
- Performance benchmarks
- Automated CI/CD pipeline

## Test Structure

```
test/
├── testdata/              # Sample hercules data files
│   ├── simple_burndown.pb    # Small test dataset
│   ├── realistic_burndown.pb # Large test dataset
│   └── README.md            # Test data documentation
├── golden/                # Golden files for visual regression
│   ├── burndown_project_golden.png
│   ├── ownership_golden.png
│   └── devs_golden.png
├── integration_test.go    # End-to-end integration tests
├── visual_regression_test.go # Visual consistency tests
├── benchmark_test.go      # Performance benchmarks
├── testdata_generator.go  # Test data generation utilities
└── create_sample_data.go  # Test data creation tool
```

## Running Tests

### Quick Start

```bash
# Run all tests with default settings
make test

# Run tests quickly (no coverage/benchmarks)
make test-quick

# Run comprehensive test suite
make test-all
```

### Specific Test Types

```bash
# Unit tests only
make test-unit

# Integration tests only
make test-integration

# Visual regression tests
make test-visual

# Performance benchmarks
make test-bench
```

### Using the Test Runner Script

```bash
# Basic test run
./scripts/run_tests.sh

# With verbose output
./scripts/run_tests.sh --verbose

# With benchmarks
./scripts/run_tests.sh --benchmarks

# With visual regression tests
./scripts/run_tests.sh --visual

# Regenerate golden files
./scripts/run_tests.sh --regenerate-golden
```

## Test Categories

### 1. Unit Tests

Unit tests are located alongside source code files with `_test.go` suffix:

- `internal/modes/*_test.go` - Analysis mode tests
- `internal/readers/*_test.go` - Data reader tests
- `internal/graphics/*_test.go` - Visualization tests

**Key Features:**

- Test individual functions and methods
- Mock dependencies for isolation
- Fast execution
- High code coverage

**Example:**

```go
func TestGenerateBurndownPlot(t *testing.T) {
    tmpDir := t.TempDir()
    outputPath := filepath.Join(tmpDir, "test_burndown.png")

    testMatrix := [][]int{
        {100, 90, 80},
        {120, 100, 85},
    }

    err := generateBurndownPlot("test", testMatrix, outputPath, false, &startTime, &endTime, "day")
    if err != nil {
        t.Errorf("generateBurndownPlot() error = %v", err)
    }

    // Verify output file exists
    if _, err := os.Stat(outputPath); os.IsNotExist(err) {
        t.Errorf("Output file was not created: %s", outputPath)
    }
}
```

### 2. Integration Tests

Integration tests verify end-to-end functionality with real data:

**Test Cases:**

- Complete protobuf data processing
- Multi-format data reading
- CLI command execution
- Output file generation

**Example:**

```go
func TestEndToEndBurndownProject(t *testing.T) {
    // Generate test data
    generator := NewTestDataGenerator(12345)
    testData := generator.GenerateSimpleBurndownData()

    // Write and read data
    pbData, err := generator.SerializeToBytes(testData)
    // ... test reading and processing
}
```

### 3. Visual Regression Tests

Visual regression tests ensure chart output consistency:

**Features:**

- Golden file comparison
- Pixel-perfect matching
- Difference highlighting
- Automated golden file generation

**Workflow:**

1. Generate chart with test data
2. Compare with golden file
3. Fail if differences detected
4. Save diff image for inspection

**Example:**

```bash
# Run visual tests
make test-visual

# Regenerate golden files (when chart design changes)
make golden-regen
```

### 4. Performance Benchmarks

Benchmark tests measure and track performance:

**Metrics:**

- Operations per second
- Memory usage
- Allocation patterns
- Execution time

**Example:**

```go
func BenchmarkDataGeneration(b *testing.B) {
    generator := NewTestDataGenerator(time.Now().UnixNano())

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        data := generator.GenerateSimpleBurndownData()
        if data == nil {
            b.Fatal("Generated nil data")
        }
    }
}
```

## Test Data

### Generation

Test data is generated using deterministic algorithms with fixed seeds for reproducibility:

```bash
# Generate test data files
make testdata

# Or run directly
go run test/create_sample_data.go
```

### Types

1. **Simple Burndown Data**

   - 3 people, 2 files, 30 days
   - Small scale for fast tests
   - File: `simple_burndown.pb`

2. **Realistic Burndown Data**

   - 10 people, 50 files, 365 days
   - Large scale for performance tests
   - File: `realistic_burndown.pb`

3. **Specialized Data**
   - Developer statistics
   - Language statistics
   - Time series data
   - File coupling data

## Coverage

### Generating Coverage Reports

```bash
# Run tests with coverage
make coverage

# View coverage by function
make coverage-func

# HTML report location
open test_output/coverage.html
```

### Coverage Goals

- **Target:** 80%+ overall coverage
- **Critical paths:** 90%+ coverage
- **New code:** 100% coverage required

### Exclusions

Coverage excludes:

- Generated protobuf files
- Test utility functions
- Main package entry points

## Continuous Integration

### GitHub Actions

The CI pipeline includes:

- Multi-Go version testing (1.21, 1.22)
- Code quality checks
- Security scanning
- Performance regression detection
- Cross-platform builds

### Workflow Triggers

- Push to main/master/develop
- Pull requests
- Manual dispatch

### Quality Gates

All checks must pass:

- ✅ Unit tests
- ✅ Integration tests
- ✅ Code formatting
- ✅ Linting
- ✅ Security scan
- ✅ Build verification

## Writing Tests

### Best Practices

1. **Test Names**: Use descriptive names explaining what is being tested
2. **Table Tests**: Use for multiple similar test cases
3. **Temporary Directories**: Always use `t.TempDir()` for file operations
4. **Cleanup**: Defer cleanup operations
5. **Assertions**: Use clear, descriptive error messages

### Example Test Structure

```go
func TestFunctionName(t *testing.T) {
    // Arrange - set up test data
    input := setupTestData()
    expected := expectedResult()

    // Act - execute the function
    result, err := functionUnderTest(input)

    // Assert - verify results
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }

    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

### Table-Driven Tests

```go
func TestCalculation(t *testing.T) {
    testCases := []struct {
        name     string
        input    int
        expected int
    }{
        {"positive", 5, 10},
        {"negative", -3, -6},
        {"zero", 0, 0},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := calculate(tc.input)
            if result != tc.expected {
                t.Errorf("Expected %d, got %d", tc.expected, result)
            }
        })
    }
}
```

## Debugging Tests

### Running Individual Tests

```bash
# Run specific test
go test -run TestFunctionName ./internal/modes/

# Run with verbose output
go test -v -run TestFunctionName ./internal/modes/

# Run with race detection
go test -race -run TestFunctionName ./internal/modes/
```

### Test Debugging

1. **Add Print Statements**: Use `t.Logf()` for debug output
2. **Temporary Files**: Check `t.TempDir()` contents
3. **Coverage**: Use `go test -cover` to see what's not tested
4. **Profiling**: Use `go test -cpuprofile` for performance issues

### Common Issues

1. **Flaky Tests**: Often due to timing or random data
2. **Race Conditions**: Use `-race` flag to detect
3. **Resource Leaks**: Check for unclosed files/connections
4. **Platform Differences**: Test on multiple OS when possible

## Performance Testing

### Benchmarking Guidelines

1. **Realistic Data**: Use representative data sizes
2. **Stable Environment**: Run on consistent hardware
3. **Multiple Runs**: Use `-count=5` for statistical significance
4. **Memory Profiling**: Include `-benchmem` for allocation tracking

### Performance Regression

The CI pipeline automatically compares performance:

- Baseline from main branch
- PR branch performance
- Automated reporting of significant changes

### Optimization Workflow

1. Write benchmark
2. Establish baseline
3. Optimize code
4. Verify improvement
5. Check for regressions

## Contributing

### Adding Tests

When adding new functionality:

1. Write tests first (TDD approach)
2. Ensure comprehensive coverage
3. Include edge cases
4. Add integration test if needed
5. Update golden files for visual changes

### Test Review Checklist

- [ ] Tests cover happy path
- [ ] Tests cover error cases
- [ ] Tests cover edge cases
- [ ] Test names are descriptive
- [ ] No hardcoded paths or values
- [ ] Proper cleanup and resource management
- [ ] Updated documentation

## Troubleshooting

### Common Test Failures

1. **File Not Found**: Check test data generation
2. **Permission Denied**: Check file/directory permissions
3. **Race Conditions**: Use `-race` flag to identify
4. **Timeout**: Increase timeout with `-timeout` flag
5. **Memory Issues**: Check for memory leaks with profiling

### Getting Help

1. Check existing test patterns
2. Review test documentation
3. Look at similar tests in codebase
4. Ask in project discussions

## Future Improvements

### Planned Enhancements

- [ ] Property-based testing with quick
- [ ] Mutation testing
- [ ] Performance budgets
- [ ] Visual diff improvements
- [ ] Test parallelization optimization

### Monitoring

- Test execution time tracking
- Coverage trend analysis
- Performance regression alerts
- Flaky test detection

---

This testing infrastructure ensures the labours-go project maintains high quality, performance, and reliability standards.
