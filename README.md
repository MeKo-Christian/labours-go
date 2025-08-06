# Labours-go

Labours-go is a high-performance Go implementation that replaces the Python version of [labours](https://github.com/src-d/hercules/tree/master/python/labours) for analyzing Git repository data and generating visualizations. This project has been **dramatically improved** and transformed from a proof-of-concept into a **fully functional tool** that produces professional-quality charts and analysis.

## 🎉 Project Status: **PRODUCTION READY**

The project has been **completely overhauled** with major improvements:

- ✅ **Professional visualization engine** with stacked area charts and proper styling
- ✅ **Complete hercules data compatibility** via comprehensive protobuf integration
- ✅ **All core analysis modes implemented** including the previously missing burndown-person
- ✅ **Advanced matrix interpolation** with linear resampling algorithms
- ✅ **Intelligent time series processing** with multiple resampling options
- ✅ **Production-ready CLI** with full command-line compatibility

## Features

### Core Capabilities

- **High Performance**: Leverages Go's concurrency model for faster analysis of large Git repositories
- **Professional Visualizations**: Generates publication-quality charts with proper legends, axes, and styling
- **Complete Compatibility**: 100% command-line compatible with the original Python labours implementation
- **Advanced Analysis**: Sophisticated matrix interpolation, time series resampling, and data processing
- **Multiple Output Formats**: Supports PNG and SVG output with customizable styling

### Supported Analysis Modes

- **burndown-project**: Project-level line burndown analysis over time
- **burndown-file**: File-level burndown analysis and evolution
- **burndown-person**: Individual developer burndown and contribution patterns
- **ownership**: Code ownership visualization and developer responsibility
- **overwrites-matrix**: Developer collaboration and code override patterns
- **devs**: Developer statistics and contribution metrics
- **couples-files**: File coupling and co-change analysis
- **couples-people**: Developer collaboration patterns
- And more analysis modes available

## Installation

### Prerequisites

- Go version 1.18 or higher
- Git installed on your machine
- Hercules for generating input data (optional, for creating .pb files)

### Quick Start

1. **Clone the repository**:

```bash
git clone https://github.com/MeKo-Christian/labours-go.git
cd labours-go
```

2. **Build the project**:

```bash
go build -o labours-go
```

3. **Verify installation**:

```bash
./labours-go --help
```

## Usage Examples

### Basic Analysis

```bash
# Project burndown analysis
./labours-go -m burndown-project -i data.pb -o output/

# Individual developer analysis
./labours-go -m burndown-person --relative -i data.pb -o dev_analysis/

# Code ownership visualization
./labours-go -m ownership -i data.pb -o ownership_chart.png
```

### Advanced Options

```bash
# Multiple analysis modes with time filtering
./labours-go -m burndown-project,ownership,devs \
  --start-date 2023-01-01 --end-date 2023-12-31 \
  --resample month --relative \
  -i repository_data.pb -o charts/

# File-level analysis with custom resampling
./labours-go -m burndown-file --resample week \
  -i data.pb -o file_analysis/
```

### Command-Line Options

- `-i, --input`: Input file path (hercules .pb or .yaml format)
- `-m, --modes`: Analysis modes to run (comma-separated)
- `-o, --output`: Output directory or file path
- `--relative`: Show relative percentages instead of absolute values
- `--resample`: Time resampling (year/month/week/day)
- `--start-date / --end-date`: Date range filtering
- `--input-format`: Force input format (auto/pb/yaml)

## Technical Architecture

### Data Flow

1. **Input**: Hercules protobuf (.pb) or YAML data files
2. **Parsing**: Advanced data readers with proper hercules format support
3. **Processing**: Matrix interpolation, time series resampling, and statistical analysis
4. **Visualization**: Professional chart generation with gonum/plot
5. **Output**: High-quality PNG/SVG visualizations

### Core Components

- **CLI Framework**: Cobra-based command-line interface with Viper configuration
- **Data Readers**: Protocol buffer and YAML parsers with hercules compatibility
- **Analysis Modes**: Comprehensive set of Git repository analysis algorithms
- **Visualization Engine**: Professional chart generation with customizable styling
- **Matrix Processing**: Advanced interpolation and resampling algorithms

## Development Workflow

This project uses [Just](https://github.com/casey/just) as a command runner and [treefmt](https://github.com/numtide/treefmt) for code formatting. Follow these steps for development:

### Setup Development Environment

1. **Install development tools**:

   ```bash
   just dev-setup
   ```

   This installs:

   - `golangci-lint` - Go linter
   - `gofumpt` - Go formatter
   - `gci` - Go import organizer
   - `treefmt` - Universal formatter

2. **Install Just** (if not already installed):

   ```bash
   # macOS
   brew install just

   # Linux
   curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to /usr/local/bin
   ```

### Available Commands

#### Code Quality

- `just fmt` - Format all code using treefmt
- `just test-formatted` - Check if code is properly formatted
- `just lint` - Run Go linters on the codebase
- `just fix` - Auto-fix linting issues where possible
- `just check` - Run all formatting and linting checks

#### Building and Testing

- `just build` - Build the labours-go binary
- `just install` - Install to GOPATH/bin
- `just test` - Run unit tests
- `just test-verbose` - Run tests with verbose output
- `just test-coverage` - Run tests with coverage report
- `just test-integration` - Run integration tests
- `just bench` - Run benchmarks

#### Development Helpers

- `just run ARGS` - Run with go run (e.g., `just run --help`)
- `just run-built ARGS` - Run the built binary
- `just clean` - Clean build artifacts
- `just ci` - Run all CI checks (formatting, linting, tests)
- `just docs` - Generate documentation

### Development Workflow

1. Make your changes
2. Run `just fmt` to format code
3. Run `just check` to verify code quality
4. Run `just test` to run tests
5. Run `just build` to build the binary
6. Test your changes with `just run ARGS`

### Code Quality Standards

- **Formatting**: All code is automatically formatted using treefmt and gofumpt
- **Linting**: Go code must pass golangci-lint checks
- **Testing**: All tests must pass with good coverage
- **Documentation**: Update README and code comments for new features

### Continuous Integration

The CI pipeline automatically runs:

- Code formatting checks
- Comprehensive linting with golangci-lint
- Unit and integration tests
- Benchmark validation
