# Justfile for labours-go development

# Default recipe - show available commands
default:
    @just --list

# Development workflow commands
dev-setup:
    @echo "+ Installing development tools"
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install mvdan.cc/gofumpt@latest
    go install github.com/daixiang0/gci@latest
    go install github.com/numtide/treefmt/cmd/treefmt@latest

# Code quality
fmt:
    @echo "+ Formatting code"
    treefmt --allow-missing-formatter

test-formatted:
    @echo "+ Testing code formatting"
    treefmt --fail-on-change --allow-missing-formatter

lint:
    @echo "+ Running linters"
    golangci-lint run ./...

fix:
    @echo "+ Auto-fixing linting issues"
    golangci-lint run --fix ./...

check: test-formatted lint
    @echo "+ Code quality checks passed"

# Build and test commands
build:
    @echo "+ Building labours-go"
    go build -o labours-go

test:
    @echo "+ Running tests"
    go test ./...

test-verbose:
    @echo "+ Running tests (verbose)"
    go test -v ./...

test-coverage:
    @echo "+ Running tests with coverage"
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report: coverage.html"

# Integration tests (if they exist)
test-integration:
    @echo "+ Running integration tests"
    ./scripts/run_tests.sh

bench:
    @echo "+ Running benchmarks"
    go test -bench=. -benchmem ./...

# Clean up
clean:
    @echo "+ Cleaning build artifacts"
    rm -f labours-go
    rm -f coverage.out coverage.html
    rm -f *.png *.svg

# Install to GOPATH/bin
install:
    @echo "+ Installing labours-go"
    go install

# CI/CD simulation
ci: check test
    @echo "+ All CI checks passed"

# Development helpers
run *ARGS:
    @echo "+ Running labours-go {{ARGS}}"
    go run main.go {{ARGS}}

run-built *ARGS:
    @echo "+ Running built labours-go {{ARGS}}"
    ./labours-go {{ARGS}}

# Generate hercules data for this repository
generate-data OUTPUT_DIR="./data":
    @echo "+ Generating hercules data for labours-go repository"
    @mkdir -p {{OUTPUT_DIR}}
    @echo "+ Running hercules analysis (burndown, devs, couples)..."
    /home/christian/Code/hercules/hercules --burndown --burndown-files --burndown-people . > {{OUTPUT_DIR}}/labours-go_burndown.yaml
    /home/christian/Code/hercules/hercules --devs . > {{OUTPUT_DIR}}/labours-go_devs.yaml
    /home/christian/Code/hercules/hercules --couples . > {{OUTPUT_DIR}}/labours-go_couples.yaml
    @echo "+ Data generated in {{OUTPUT_DIR}}/"
    @ls -la {{OUTPUT_DIR}}/

# Generate hercules data in Protocol Buffer format
generate-data-pb OUTPUT_DIR="./data":
    @echo "+ Generating hercules data for labours-go repository (Protocol Buffer format)"
    @mkdir -p {{OUTPUT_DIR}}
    @echo "+ Running hercules analysis (burndown, devs, couples)..."
    /home/christian/Code/hercules/hercules --pb --burndown --burndown-files --burndown-people . > {{OUTPUT_DIR}}/labours-go_burndown.pb
    /home/christian/Code/hercules/hercules --pb --devs . > {{OUTPUT_DIR}}/labours-go_devs.pb
    /home/christian/Code/hercules/hercules --pb --couples . > {{OUTPUT_DIR}}/labours-go_couples.pb
    @echo "+ Data generated in {{OUTPUT_DIR}}/"
    @ls -la {{OUTPUT_DIR}}/

# Full analysis pipeline - generate data and create charts
analyze OUTPUT_DIR="./analysis" THEME="default":
    @echo "+ Running full analysis pipeline for labours-go repository"
    @mkdir -p {{OUTPUT_DIR}}
    just build
    ./scripts/analyze_with_hercules.sh . -o {{OUTPUT_DIR}} -t {{THEME}} -m burndown,devs,couples
    @echo "+ Analysis complete! Results in {{OUTPUT_DIR}}/"

# Documentation
docs:
    @echo "+ Generating documentation"
    go doc -all