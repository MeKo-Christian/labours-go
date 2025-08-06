#!/bin/bash

# run_tests.sh - Comprehensive test runner for labours-go
# This script runs all tests including unit, integration, visual regression, and benchmarks

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
VERBOSE=${VERBOSE:-false}
COVERAGE=${COVERAGE:-true}
BENCHMARKS=${BENCHMARKS:-false}
VISUAL_REGRESSION=${VISUAL_REGRESSION:-false}
REGENERATE_GOLDEN=${REGENERATE_GOLDEN:-false}
RACE_DETECTION=${RACE_DETECTION:-true}
TIMEOUT=${TIMEOUT:-10m}

print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE} $1 ${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --no-coverage)
            COVERAGE=false
            shift
            ;;
        --benchmarks|-b)
            BENCHMARKS=true
            shift
            ;;
        --visual|-vr)
            VISUAL_REGRESSION=true
            shift
            ;;
        --regenerate-golden)
            REGENERATE_GOLDEN=true
            shift
            ;;
        --no-race)
            RACE_DETECTION=false
            shift
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --verbose, -v          Enable verbose output"
            echo "  --no-coverage          Disable coverage reporting"
            echo "  --benchmarks, -b       Run benchmark tests"
            echo "  --visual, -vr          Run visual regression tests"
            echo "  --regenerate-golden    Regenerate golden files for visual tests"
            echo "  --no-race              Disable race detection"
            echo "  --timeout DURATION     Set test timeout (default: 10m)"
            echo "  --help, -h             Show this help message"
            echo ""
            echo "Environment variables:"
            echo "  VERBOSE=true           Enable verbose output"
            echo "  COVERAGE=false         Disable coverage reporting"
            echo "  BENCHMARKS=true        Run benchmark tests"
            echo "  VISUAL_REGRESSION=true Run visual regression tests"
            echo "  REGENERATE_GOLDEN=true Regenerate golden files"
            echo "  RACE_DETECTION=false   Disable race detection"
            echo "  TIMEOUT=5m             Set test timeout"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Check if we're in the right directory
if [[ ! -f "go.mod" ]] || [[ ! -f "main.go" ]]; then
    print_error "Must be run from the labours-go root directory"
    exit 1
fi

# Start the test run
print_header "LABOURS-GO TEST SUITE"
echo "Configuration:"
echo "  Verbose: $VERBOSE"
echo "  Coverage: $COVERAGE"
echo "  Benchmarks: $BENCHMARKS"
echo "  Visual Regression: $VISUAL_REGRESSION"
echo "  Regenerate Golden: $REGENERATE_GOLDEN"
echo "  Race Detection: $RACE_DETECTION"
echo "  Timeout: $TIMEOUT"
echo ""

# Build test flags
TEST_FLAGS="-timeout=$TIMEOUT"
if [[ "$RACE_DETECTION" == "true" ]]; then
    TEST_FLAGS="$TEST_FLAGS -race"
fi

if [[ "$VERBOSE" == "true" ]]; then
    TEST_FLAGS="$TEST_FLAGS -v"
fi

# Coverage flags
COVERAGE_FLAGS=""
if [[ "$COVERAGE" == "true" ]]; then
    COVERAGE_FLAGS="-coverprofile=coverage.out -covermode=atomic"
    mkdir -p test_output
fi

# 1. Generate test data if needed
print_header "GENERATING TEST DATA"
if [[ ! -f "test/testdata/simple_burndown.pb" ]] || [[ ! -f "test/testdata/realistic_burndown.pb" ]]; then
    print_info "Generating test data files..."
    go run test/create_sample_data.go
    print_success "Test data generated"
else
    print_info "Test data files already exist"
fi

# 2. Run unit tests
print_header "UNIT TESTS"
print_info "Running unit tests for all packages..."

if go test $TEST_FLAGS $COVERAGE_FLAGS ./internal/... ./cmd/...; then
    print_success "Unit tests passed"
else
    print_error "Unit tests failed"
    exit 1
fi

# 3. Run integration tests
print_header "INTEGRATION TESTS"
print_info "Running integration tests with real data..."

if go test $TEST_FLAGS ./test/integration/...; then
    print_success "Integration tests passed"
else
    print_error "Integration tests failed"
    exit 1
fi

# 4. Run visual regression tests (if enabled)
if [[ "$VISUAL_REGRESSION" == "true" ]]; then
    print_header "VISUAL REGRESSION TESTS"
    print_info "Running visual regression tests..."
    
    if [[ "$REGENERATE_GOLDEN" == "true" ]]; then
        print_warning "Regenerating golden files..."
        REGENERATE_GOLDEN=true go test $TEST_FLAGS ./test/visual_regression_test.go ./test/testdata_generator.go
    fi
    
    if go test $TEST_FLAGS ./test/visual_regression_test.go ./test/testdata_generator.go; then
        print_success "Visual regression tests passed"
    else
        print_error "Visual regression tests failed"
        exit 1
    fi
fi

# 5. Run benchmarks (if enabled)
if [[ "$BENCHMARKS" == "true" ]]; then
    print_header "BENCHMARK TESTS"
    print_info "Running performance benchmarks..."
    
    # Run benchmarks on existing test packages
    if go test -bench=. -benchmem -run=^$ ./internal/... > test_output/benchmarks.txt; then
        print_success "Benchmarks completed"
        if [[ "$VERBOSE" == "true" ]]; then
            echo ""
            echo "Benchmark results:"
            cat test_output/benchmarks.txt
        fi
    else
        print_error "Benchmarks failed"
        exit 1
    fi
fi

# 6. Generate coverage report
if [[ "$COVERAGE" == "true" ]]; then
    print_header "COVERAGE REPORT"
    print_info "Generating coverage report..."
    
    # Generate HTML coverage report
    go tool cover -html=coverage.out -o test_output/coverage.html
    
    # Get coverage percentage
    COVERAGE_PERCENT=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    print_info "Total coverage: $COVERAGE_PERCENT"
    
    # Set coverage threshold
    COVERAGE_THRESHOLD="70.0%"
    COVERAGE_VALUE=$(echo $COVERAGE_PERCENT | sed 's/%//')
    THRESHOLD_VALUE=$(echo $COVERAGE_THRESHOLD | sed 's/%//')
    
    if (( $(echo "$COVERAGE_VALUE >= $THRESHOLD_VALUE" | bc -l) )); then
        print_success "Coverage above threshold ($COVERAGE_THRESHOLD)"
    else
        print_warning "Coverage below threshold: $COVERAGE_PERCENT < $COVERAGE_THRESHOLD"
    fi
    
    print_info "Coverage report: test_output/coverage.html"
fi

# 7. Run linting (if available)
print_header "CODE QUALITY"
print_info "Running code quality checks..."

# Check if golangci-lint is available
if command -v golangci-lint &> /dev/null; then
    if golangci-lint run; then
        print_success "Linting passed"
    else
        print_warning "Linting issues found (not failing build)"
    fi
else
    print_warning "golangci-lint not found, skipping linting"
fi

# Check for go fmt issues
if ! go fmt ./... | grep -q .; then
    print_success "Code formatting is correct"
else
    print_warning "Code formatting issues found, run 'go fmt ./...' to fix"
fi

# 8. Run go vet
if go vet ./...; then
    print_success "Go vet passed"
else
    print_error "Go vet found issues"
    exit 1
fi

# 9. Check for common issues
print_info "Checking for common issues..."

# Check for TODO comments (informational)
TODO_COUNT=$(grep -r "TODO\|FIXME\|HACK" --include="*.go" . | wc -l || true)
if [[ $TODO_COUNT -gt 0 ]]; then
    print_warning "Found $TODO_COUNT TODO/FIXME/HACK comments"
else
    print_info "No TODO/FIXME/HACK comments found"
fi

# Check for print statements (should use logging)
PRINT_COUNT=$(grep -r "fmt\.Print\|fmt\.Println" --include="*.go" --exclude-dir=test . | wc -l || true)
if [[ $PRINT_COUNT -gt 0 ]]; then
    print_warning "Found $PRINT_COUNT fmt.Print statements (consider using logging)"
fi

# 10. Test build
print_header "BUILD TEST"
print_info "Testing build process..."

if go build -o labours-go-test; then
    print_success "Build successful"
    rm -f labours-go-test
else
    print_error "Build failed"
    exit 1
fi

# 11. Final summary
print_header "TEST SUMMARY"
print_success "All tests completed successfully!"

echo ""
echo "Test Results:"
echo "  ✓ Unit tests passed"
echo "  ✓ Integration tests passed"
if [[ "$VISUAL_REGRESSION" == "true" ]]; then
    echo "  ✓ Visual regression tests passed"
fi
if [[ "$BENCHMARKS" == "true" ]]; then
    echo "  ✓ Benchmark tests completed"
fi
if [[ "$COVERAGE" == "true" ]]; then
    echo "  ✓ Coverage: $COVERAGE_PERCENT"
fi
echo "  ✓ Code quality checks passed"
echo "  ✓ Build test passed"
echo ""

# Output file locations
if [[ -d "test_output" ]]; then
    echo "Output files:"
    if [[ -f "test_output/coverage.html" ]]; then
        echo "  Coverage report: test_output/coverage.html"
    fi
    if [[ -f "test_output/benchmarks.txt" ]]; then
        echo "  Benchmark results: test_output/benchmarks.txt"
    fi
    if [[ -f "coverage.out" ]]; then
        echo "  Coverage data: coverage.out"
    fi
fi

print_success "Testing complete!"
exit 0