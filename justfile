# Justfile for labours-go development

# Default recipe - show available commands
default:
    @just --list

# === ESSENTIAL COMMANDS ===

# Build the project
build:
    @echo "🔨 Building labours-go"
    go build -o labours-go

# Run tests
test:
    @echo "🧪 Running tests"
    go test ./...

# Check code quality (format + lint)
check:
    @echo "✅ Running code quality checks"
    @if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run ./...; else echo "golangci-lint not installed, skipping lint"; fi
    @if command -v treefmt >/dev/null 2>&1; then treefmt --fail-on-change --allow-missing-formatter; else echo "treefmt not installed, skipping format check"; fi

# Clean build artifacts
clean:
    @echo "🧹 Cleaning build artifacts"
    rm -f labours-go coverage.out coverage.html

# === DEVELOPMENT HELPERS ===

# Run with arguments (e.g., just run -i data.yaml -m burndown-project)
run *ARGS:
    @echo "🚀 Running labours-go {{ARGS}}"
    go run main.go {{ARGS}}

# Run the built binary
run-built *ARGS:
    just build
    ./labours-go {{ARGS}}

# === TESTING ===

# Run tests with coverage report
test-coverage:
    @echo "📊 Running tests with coverage"
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report: coverage.html"

# Run integration tests
test-integration:
    @echo "🔗 Running integration tests"
    ./scripts/run_tests.sh

# === CHART GENERATION ===

# Generate example burndown chart
demo-burndown:
    @echo "📈 Generating demo burndown chart"
    just build
    ./labours-go -i example_data/hercules_burndown.yaml -m burndown-project -o demo_burndown.png
    @echo "Chart saved as demo_burndown.png"

# Compare with Python reference
test-chart:
    @echo "📊 Testing chart generation vs Python reference"
    just build
    ./labours-go -i example_data/hercules_burndown.yaml -m burndown-project -o analysis_results/test_chart.png
    @echo "Chart saved as analysis_results/test_chart.png"