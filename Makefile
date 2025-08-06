# Makefile for labours-go
# Provides convenient targets for building, testing, and development tasks

.PHONY: build test test-all test-unit test-integration test-visual test-bench clean help
.DEFAULT_GOAL := help

# Configuration
BINARY_NAME=labours-go
BUILD_DIR=bin
TEST_OUTPUT_DIR=test_output
COVERAGE_FILE=coverage.out

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build flags
BUILD_FLAGS=-ldflags="-w -s"
TEST_FLAGS=-timeout=10m -race

help: ## Show this help message
	@echo 'Usage: make <target>'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build the labours-go binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

build-all: ## Build for multiple platforms
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Multi-platform build complete"

# Test targets
test: ## Run all tests with standard configuration
	@./scripts/run_tests.sh

test-quick: ## Run tests without coverage or benchmarks (faster)
	@COVERAGE=false BENCHMARKS=false ./scripts/run_tests.sh

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	$(GOTEST) $(TEST_FLAGS) -coverprofile=$(COVERAGE_FILE) ./internal/... ./cmd/...
	@echo "Unit tests complete"

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	$(GOTEST) $(TEST_FLAGS) ./test/integration_test.go ./test/testdata_generator.go
	@echo "Integration tests complete"

test-visual: ## Run visual regression tests
	@echo "Running visual regression tests..."
	@VISUAL_REGRESSION=true ./scripts/run_tests.sh

test-bench: ## Run benchmark tests
	@echo "Running benchmark tests..."
	@mkdir -p $(TEST_OUTPUT_DIR)
	$(GOTEST) -bench=. -benchmem -run=^$$ ./test/benchmark_test.go ./test/testdata_generator.go | tee $(TEST_OUTPUT_DIR)/benchmarks.txt
	@echo "Benchmark results saved to $(TEST_OUTPUT_DIR)/benchmarks.txt"

test-all: ## Run comprehensive test suite with all options
	@BENCHMARKS=true VISUAL_REGRESSION=true ./scripts/run_tests.sh

# Coverage targets
coverage: test-unit ## Generate and view coverage report
	@mkdir -p $(TEST_OUTPUT_DIR)
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(TEST_OUTPUT_DIR)/coverage.html
	@echo "Coverage report generated: $(TEST_OUTPUT_DIR)/coverage.html"

coverage-func: test-unit ## Show coverage by function
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)

# Development targets
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies updated"

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) ./...
	@echo "Code formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOCMD) vet ./...
	@echo "Vet complete"

lint: ## Run golangci-lint (if available)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running golangci-lint..."; \
		golangci-lint run; \
		echo "Linting complete"; \
	else \
		echo "golangci-lint not found, please install it:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Data generation targets
testdata: ## Generate test data files
	@echo "Generating test data..."
	$(GOCMD) run test/create_sample_data.go
	@echo "Test data generated"

golden-regen: ## Regenerate golden files for visual tests
	@echo "Regenerating golden files..."
	@REGENERATE_GOLDEN=true VISUAL_REGRESSION=true ./scripts/run_tests.sh
	@echo "Golden files regenerated"

# Quality targets
quality: fmt vet lint ## Run all code quality checks

check: quality test-quick ## Quick development check (quality + fast tests)

ci: ## Continuous integration target
	@echo "Running CI pipeline..."
	@make quality
	@make test-all
	@echo "CI pipeline complete"

# Install targets
install: build ## Install the binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install .
	@echo "Installation complete"

# Clean targets
clean: ## Clean build artifacts and test outputs
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(TEST_OUTPUT_DIR)
	rm -f $(COVERAGE_FILE)
	rm -f $(BINARY_NAME)
	@echo "Clean complete"

clean-testdata: ## Remove generated test data
	@echo "Cleaning test data..."
	rm -rf test/testdata/*.pb
	rm -rf test/golden/*.png
	@echo "Test data cleaned"

# Development workflow targets
dev-setup: deps testdata ## Set up development environment
	@echo "Development environment ready"

dev-test: fmt vet test-quick ## Quick development test cycle

release-check: ## Pre-release checks
	@echo "Running pre-release checks..."
	@make clean
	@make quality
	@make test-all
	@make build-all
	@echo "Release checks complete"

# Docker targets (if needed)
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t labours-go:latest .
	@echo "Docker build complete"

docker-test: ## Run tests in Docker
	@echo "Running tests in Docker..."
	docker run --rm -v $(PWD):/app -w /app golang:1.22 make test
	@echo "Docker tests complete"

# Documentation targets
docs: ## Generate documentation
	@echo "Generating documentation..."
	$(GOCMD) doc -all . > docs/api.md
	@echo "Documentation generated"

# Show current configuration
config: ## Show current configuration
	@echo "Configuration:"
	@echo "  Binary name: $(BINARY_NAME)"
	@echo "  Build directory: $(BUILD_DIR)"
	@echo "  Test output directory: $(TEST_OUTPUT_DIR)"
	@echo "  Go version: $(shell $(GOCMD) version)"
	@echo "  Working directory: $(PWD)"

# Show project status
status: ## Show project status
	@echo "Project Status:"
	@echo "  Git branch: $(shell git branch --show-current 2>/dev/null || echo 'unknown')"
	@echo "  Git status: $(shell git status --porcelain | wc -l) files changed"
	@echo "  Go modules: $(shell $(GOMOD) list -m all | wc -l) dependencies"
	@echo "  Test files: $(shell find . -name '*_test.go' | wc -l) test files"
	@echo "  Source files: $(shell find . -name '*.go' -not -name '*_test.go' | wc -l) source files"