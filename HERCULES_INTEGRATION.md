# Hercules Integration Guide

This guide explains how to use Hercules with Labours-Go to create a complete Git repository analytics pipeline.

## Overview

**Hercules** → **Labours-Go** represents a powerful two-stage pipeline:

1. **Hercules** analyzes Git repositories and extracts structured data
2. **Labours-Go** creates professional visualizations from that data

```
Git Repository → [Hercules Analysis] → Data Files (.yaml/.pb) → [Labours-Go Visualization] → Charts & Images
```

## Prerequisites

### Required Software

- **Hercules Binary**: Download or build from [hercules repository](https://github.com/src-d/hercules)
- **Labours-Go**: This tool (build with `go build -o labours-go`)
- **Git Repository**: Any git repository you want to analyze

### Verify Your Setup

```bash
# Check hercules is available
hercules --help

# Check labours-go is built
./labours-go --help

# Verify git repository
ls /path/to/your/repo/.git
```

## Quick Start

### 1. Use the Quick Analysis Script

```bash
# Analyze any repository with one command
./scripts/quick_analysis.sh /path/to/your/repository

# Specify output directory
./scripts/quick_analysis.sh /path/to/your/repository ./my_analysis_results
```

This will:
- Run Hercules burndown and developer analysis
- Generate project burndown chart, file burndown charts, and developer statistics
- Save all results to the specified directory

### 2. Manual Pipeline

```bash
# Step 1: Generate data with Hercules
hercules --burndown --burndown-files --burndown-people /path/to/repo > burndown.yaml
hercules --devs /path/to/repo > devs.yaml

# Step 2: Create visualizations with Labours-Go
./labours-go -i burndown.yaml -m burndown-project -o project_burndown.png
./labours-go -i devs.yaml -m devs -o developer_stats.png
```

## Complete Analysis Types

### Available Hercules Analysis Modes

| Hercules Flag | Description | Labours-Go Modes |
|---------------|-------------|-------------------|
| `--burndown` | Code evolution over time | `burndown-project`, `burndown-file`, `burndown-person` |
| `--devs` | Developer statistics | `devs` |
| `--couples` | File/developer coupling | `couples-files`, `couples-people` |
| `--file-history` | File ownership analysis | `ownership` |

### Comprehensive Analysis Script

For complete analysis of all supported modes:

```bash
./scripts/analyze_with_hercules.sh /path/to/repo -m burndown,devs,couples,ownership -o complete_analysis/
```

Advanced options:
```bash
# Use dark theme
./scripts/analyze_with_hercules.sh /path/to/repo -t dark -o analysis/

# Use Protocol Buffer format (faster for large repos)
./scripts/analyze_with_hercules.sh /path/to/repo --pb -o analysis/

# Custom hercules binary location
./scripts/analyze_with_hercules.sh /path/to/repo --hercules /custom/path/hercules
```

## Data Formats

### YAML Format (Default)

- **Human-readable**
- **Easier to debug**
- **Slightly larger file size**
- **Universal compatibility**

```bash
hercules --burndown /path/to/repo > analysis.yaml
./labours-go -i analysis.yaml -m burndown-project -o chart.png
```

### Protocol Buffer Format

- **Faster processing**
- **Smaller file size**
- **Binary format**
- **Optimal for large repositories**

```bash
hercules --burndown --pb /path/to/repo > analysis.pb
./labours-go -i analysis.pb -m burndown-project -o chart.png
```

## Analysis Examples

### 1. Project Evolution (Burndown)

Shows how code has evolved over time, highlighting new vs. modified vs. old code.

```bash
# Generate burndown data
hercules --burndown --burndown-files /path/to/repo > burndown.yaml

# Create project-level burndown chart
./labours-go -i burndown.yaml -m burndown-project -o project_evolution.png

# Create file-level breakdown (if data available)
./labours-go -i burndown.yaml -m burndown-file -o file_breakdown.png
```

**Result**: Stacked area charts showing code age distribution over time.

### 2. Developer Activity

Analyzes individual developer contributions and patterns.

```bash
# Generate developer data
hercules --devs /path/to/repo > devs.yaml

# Create developer statistics chart
./labours-go -i devs.yaml -m devs -o developer_activity.png
```

**Result**: Bar charts and line graphs showing commits, lines added/removed per developer.

### 3. File Coupling Analysis

Identifies files that are frequently modified together.

```bash
# Generate coupling data
hercules --couples /path/to/repo > couples.yaml

# Create coupling visualization
./labours-go -i couples.yaml -m couples-files -o file_coupling.png
```

**Result**: Heatmap showing file co-modification patterns.

### 4. Code Ownership

Shows which developers have the most influence over different parts of the codebase.

```bash
# Generate file history data
hercules --file-history /path/to/repo > ownership.yaml

# Create ownership visualization
./labours-go -i ownership.yaml -m ownership -o code_ownership.png
```

**Result**: Charts showing code ownership distribution.

## Theming and Customization

### Built-in Themes

Labours-Go supports several built-in themes:

```bash
# List available themes
./labours-go --list-themes

# Use dark theme
./labours-go -i data.yaml -m burndown-project --theme dark -o chart.png

# Use minimal theme
./labours-go -i data.yaml -m burndown-project --theme minimal -o chart.png
```

Available themes:
- **default**: Classic blue/orange with white background
- **dark**: Dark theme with bright colors
- **minimal**: Grayscale minimalist appearance  
- **vibrant**: High-contrast bright colors

### Custom Themes

```bash
# Export existing theme for customization
./labours-go --export-theme dark

# Edit the generated YAML file and load it
./labours-go --load-theme my-custom-theme.yaml -i data.yaml -m burndown-project -o chart.png
```

## Performance Tips

### For Large Repositories

1. **Use Protocol Buffer format**:
   ```bash
   hercules --burndown --pb /large/repo > analysis.pb
   ```

2. **Use hibernation options** (reduces memory usage):
   ```bash
   hercules --burndown --hibernation-distance 1000 /large/repo > analysis.yaml
   ```

3. **Reduce granularity** for faster processing:
   ```bash
   hercules --burndown --granularity 10 --sampling 10 /large/repo > analysis.yaml
   ```

### Batch Processing

For analyzing multiple repositories:

```bash
#!/bin/bash
for repo in /path/to/repos/*; do
    if [[ -d "$repo/.git" ]]; then
        echo "Analyzing $repo"
        ./scripts/quick_analysis.sh "$repo" "analysis/$(basename $repo)"
    fi
done
```

## Troubleshooting

### Common Issues

#### 1. "proto: cannot parse invalid wire-format data"

**Problem**: Protocol buffer format mismatch
**Solution**: Use YAML format instead:
```bash
hercules --burndown /path/to/repo > analysis.yaml  # Remove --pb flag
```

#### 2. "hercules: command not found"

**Problem**: Hercules not in PATH
**Solutions**:
```bash
# Option 1: Use full path
/path/to/hercules/hercules --burndown /repo > analysis.yaml

# Option 2: Set environment variable
export HERCULES_BINARY="/path/to/hercules/hercules"
./scripts/quick_analysis.sh /repo
```

#### 3. Empty or Missing Data

**Problem**: Analysis produces no usable data
**Causes & Solutions**:
- **Small repository**: Use `--sampling 1 --granularity 1`
- **Recent repository**: Reduce time intervals
- **No merge history**: Some analyses require merge commits

#### 4. "No configuration file found"

**Problem**: Missing config (not actually an error)
**Solution**: This is just a warning, analysis continues normally

### Debug Mode

Enable verbose output for troubleshooting:

```bash
# Hercules verbose mode
hercules --burndown --print-actions /repo > analysis.yaml

# Labours-Go debug mode (if implemented)
./labours-go -i analysis.yaml -m burndown-project -o chart.png --verbose
```

## Advanced Integration

### Custom Pipeline Scripts

Create your own analysis pipeline:

```bash
#!/bin/bash
set -e

REPO="$1"
OUTPUT="$2"

echo "Starting comprehensive analysis of $REPO"

# Create output structure
mkdir -p "$OUTPUT"/{data,charts,reports}

# Phase 1: Data extraction
echo "Extracting data with Hercules..."
hercules --burndown --pb "$REPO" > "$OUTPUT/data/burndown.pb"
hercules --devs "$REPO" > "$OUTPUT/data/devs.yaml"
hercules --couples "$REPO" > "$OUTPUT/data/couples.yaml"

# Phase 2: Visualization
echo "Creating visualizations..."
./labours-go -i "$OUTPUT/data/burndown.pb" -m burndown-project -o "$OUTPUT/charts/evolution.png"
./labours-go -i "$OUTPUT/data/devs.yaml" -m devs -o "$OUTPUT/charts/developers.png"
./labours-go -i "$OUTPUT/data/couples.yaml" -m couples-files -o "$OUTPUT/charts/coupling.png"

# Phase 3: Report generation (custom)
echo "Repository Analysis Report" > "$OUTPUT/reports/summary.txt"
echo "Generated: $(date)" >> "$OUTPUT/reports/summary.txt"
echo "Repository: $REPO" >> "$OUTPUT/reports/summary.txt"

echo "Analysis complete! Results in $OUTPUT"
```

### Integration with CI/CD

Example GitHub Actions workflow:

```yaml
name: Repository Analytics
on:
  schedule:
    - cron: '0 6 * * 1'  # Weekly on Monday
  workflow_dispatch:

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0  # Full history for analysis
      
      - name: Setup Go
        uses: actions/setup-go@v3
        with:
          go-version: '1.19'
      
      - name: Install Hercules
        run: |
          curl -L https://github.com/src-d/hercules/releases/latest/download/hercules-linux-amd64 -o hercules
          chmod +x hercules
      
      - name: Build Labours-Go
        run: go build -o labours-go
      
      - name: Run Analysis
        run: ./scripts/quick_analysis.sh . analytics-$(date +%Y%m%d)
      
      - name: Upload Results
        uses: actions/upload-artifact@v3
        with:
          name: repository-analytics
          path: analytics-*
```

## Analysis Results

All generated analysis results are centralized in the **`analysis_results/`** directory:
- **Main Charts**: Project burndown and developer statistics  
- **Sample Data**: Example hercules output files
- **Comprehensive Analysis**: Complete analysis with multiple chart types
- **README**: Detailed explanation of all generated files

See `analysis_results/README.md` for a complete index of available visualizations.

## Further Reading

- **Hercules Documentation**: https://github.com/src-d/hercules
- **Labours-Go Modes**: See `CLAUDE.md` for complete mode descriptions
- **Visualization Examples**: Check `TESTING.md` for more examples
- **Python Labours**: Original implementation reference at https://github.com/src-d/hercules/tree/master/python/labours