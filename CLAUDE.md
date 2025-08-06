# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Labours-go is a Go implementation replacing the Python version of [labours](https://github.com/src-d/hercules/tree/master/python/labours) for analyzing Git repository data and generating visualizations. The project is still under development and considered a proof-of-concept.

## Build and Development Commands

```bash
# Build the project
go build -o labours-go

# Run the built binary
./labours-go

# Build and install to GOPATH/bin
go install

# Run with go run
go run main.go [flags]

# Clean built binaries
rm labours-go
```

## Architecture

### Core Components

- **main.go**: Entry point that calls cmd.Execute()
- **cmd/**: CLI command structure using Cobra framework
  - `root.go`: Root command definition with all flags and configuration
  - `modes.go`: Mode handlers mapping mode names to execution functions
  - `helpers.go`: Helper functions for date parsing, input detection
- **internal/modes/**: Analysis mode implementations
  - `burndown.go`: Core burndown chart generation logic
  - `burndownProject.go`, `burndownFile.go`, `burndownPerson.go`: Specific burndown implementations
  - `ownership.go`, `overwrites.go`, `devs.go`: Other analysis modes
- **internal/readers/**: Data input handling
  - `reader.go`: Reader interface defining data access methods
  - `pb_reader.go`: Protocol buffer format reader
  - `yaml_reader.go`: YAML format reader
- **internal/graphics/**: Visualization components
  - `stacked-plot.go`: Stacked area chart generation using gonum/plot
  - `heatmap.go`: Heatmap visualization
  - `colors.go`: Color scheme definitions

### Key Libraries

- `github.com/spf13/cobra`: CLI framework
- `github.com/spf13/viper`: Configuration management
- `gonum.org/v1/plot`: Plotting and visualization
- `github.com/schollz/progressbar/v3`: Progress indication
- `github.com/gogo/protobuf`: Protocol buffer support
- `gopkg.in/yaml.v3`: YAML parsing

### Data Flow

1. CLI flags parsed by Cobra/Viper in `cmd/root.go`
2. Input format detected and appropriate reader selected in `cmd/helpers.go`
3. Modes resolved and executed via handlers in `cmd/modes.go`
4. Each mode calls corresponding functions in `internal/modes/`
5. Graphics generation handled by `internal/graphics/` packages
6. Output files saved as SVG/PNG images

### Available Analysis Modes

- `burndown-project`: Project-level burndown analysis
- `burndown-file`: File-level burndown analysis
- `ownership`: Code ownership visualization
- `overwrites-matrix`: Developer overwrite patterns
- `devs`: Developer statistics (placeholder)
- `couples-files`, `couples-people`, `couples-shotness`: Coupling analysis (placeholders)

### Input Formats

- Protocol Buffer (`.pb`) files via `pb_reader.go`
- YAML files via `yaml_reader.go`
- Auto-detection based on file extension or content

### Configuration

- Uses Viper for configuration management
- Looks for `config.yaml` in current directory or `$HOME/.labours-go/`
- All CLI flags can be set via configuration file