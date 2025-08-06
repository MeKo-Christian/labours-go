# Hercules + Labours-Go Usage Examples

This document shows all the different ways to use Hercules with Labours-Go for Git repository analysis.

## Method 1: Direct CLI Integration (New!)

The simplest way to analyze any git repository:

```bash
# Analyze any repository directly with one command
./labours-go --from-repo /path/to/your/repo -m burndown-project -o analysis.png

# Multiple modes
./labours-go --from-repo /path/to/your/repo -m burndown-project,devs -o analysis/

# With themes
./labours-go --from-repo /path/to/your/repo -m burndown-project --theme dark -o analysis.png

# Custom hercules path
./labours-go --hercules /custom/path/hercules --from-repo /path/to/repo -m burndown-project
```

**How it works:**
- Labours-Go automatically finds and runs Hercules 
- Processes the data and creates visualizations
- Cleans up temporary files
- One-command operation

## Method 2: Quick Analysis Script

For standard analysis with multiple charts:

```bash
# Quick analysis with defaults (burndown + developer stats)
./scripts/quick_analysis.sh /path/to/your/repo

# Custom output directory
./scripts/quick_analysis.sh /path/to/your/repo ./my_analysis

# Output shows:
#   📋 Results saved to: my_analysis
#   Generated files:
#   my_analysis/burndown.yaml
#   my_analysis/devs.yaml  
#   my_analysis/burndown_project.png
#   my_analysis/developer_stats.png
```

## Method 3: Comprehensive Analysis Script

For complete analysis with all modes:

```bash
# All analysis modes
./scripts/analyze_with_hercules.sh /path/to/repo -m burndown,devs,couples,ownership

# With custom settings
./scripts/analyze_with_hercules.sh /path/to/repo \
    -m burndown,devs,couples \
    -t dark \
    -o complete_analysis/ \
    --hercules /custom/hercules/path

# Protocol Buffer format (faster for large repos)
./scripts/analyze_with_hercules.sh /path/to/repo --pb -o analysis/
```

## Method 4: Manual Two-Step Process

For maximum control:

```bash
# Step 1: Run Hercules manually
hercules --burndown --burndown-files --burndown-people /path/to/repo > data.yaml
hercules --devs /path/to/repo > devs.yaml

# Step 2: Create visualizations
./labours-go -i data.yaml -m burndown-project -o project_burndown.png
./labours-go -i data.yaml -m burndown-file -o file_breakdown.png
./labours-go -i devs.yaml -m devs -o developer_stats.png
```

## Real-World Examples

### Analyzing Your Own Project

```bash
# Analyze the labours-go project itself
./labours-go --from-repo . -m burndown-project,devs -o labours_go_analysis.png

# Expected output:
# Using hercules: /usr/local/bin/hercules
# Analyzing repository: .
# Running hercules burndown analysis...
# Hercules analysis complete, creating visualizations...
# Chart saved to labours_go_analysis.png
```

### Batch Analysis of Multiple Projects

```bash
#!/bin/bash
for repo in ~/projects/*/; do
    if [[ -d "$repo/.git" ]]; then
        echo "Analyzing $(basename $repo)"
        ./labours-go --from-repo "$repo" \
            -m burndown-project \
            -o "analysis/$(basename $repo)_burndown.png"
    fi
done
```

### CI/CD Integration

```yaml
# .github/workflows/repo-analysis.yml
name: Repository Analysis
on:
  schedule:
    - cron: '0 6 * * 1'  # Weekly
jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0
      - name: Setup Go
        uses: actions/setup-go@v3
        with:
          go-version: '1.19'
      - name: Install Hercules
        run: |
          wget https://github.com/src-d/hercules/releases/latest/download/hercules-linux-amd64
          chmod +x hercules-linux-amd64
          sudo mv hercules-linux-amd64 /usr/local/bin/hercules
      - name: Build Labours-Go
        run: go build -o labours-go
      - name: Run Analysis
        run: |
          ./labours-go --from-repo . \
            -m burndown-project,devs \
            --theme dark \
            -o analysis/weekly-report.png
```

## Analysis Type Examples

### 1. Code Evolution (Burndown)

Shows how code changes over time:

```bash
# Project-level evolution
./labours-go --from-repo /path/to/repo -m burndown-project -o evolution.png

# File-level breakdown (if data available)  
./labours-go --from-repo /path/to/repo -m burndown-file -o file_evolution.png

# Per-developer evolution
./labours-go --from-repo /path/to/repo -m burndown-person -o developer_evolution.png
```

### 2. Developer Activity

Shows developer contribution patterns:

```bash
# Developer statistics
./labours-go --from-repo /path/to/repo -m devs -o developer_activity.png

# Expected chart: Bar chart showing commits/lines added/removed per developer
```

### 3. File Coupling

Shows which files are frequently modified together:

```bash
# File coupling heatmap
./labours-go --from-repo /path/to/repo -m couples-files -o file_coupling.png

# Expected chart: Heatmap showing co-modification patterns
```

## Advanced Usage

### Custom Hercules Flags

```bash
# Reduce granularity for faster processing
./labours-go --from-repo /large/repo \
    --hercules-flags "--granularity 10 --sampling 10" \
    -m burndown-project

# Use hibernation for memory efficiency
./labours-go --from-repo /large/repo \
    --hercules-flags "--hibernation-distance 1000" \
    -m burndown-project
```

### Theme Customization

```bash
# List available themes
./labours-go --list-themes

# Export theme for customization
./labours-go --export-theme dark

# Use custom theme
./labours-go --from-repo /repo \
    --load-theme my-custom-theme.yaml \
    -m burndown-project
```

### Output Formats

```bash
# PNG output (default)
./labours-go --from-repo /repo -m burndown-project -o chart.png

# SVG output
./labours-go --from-repo /repo -m burndown-project -o chart.svg

# Directory output (multiple files)
./labours-go --from-repo /repo -m burndown-project,devs -o analysis/
```

## Comparison: Before vs After Integration

### Before Integration (Manual)
```bash
# 1. Find hercules binary
which hercules

# 2. Run hercules manually
hercules --burndown /repo > data.yaml

# 3. Run labours-go
./labours-go -i data.yaml -m burndown-project -o chart.png

# 4. Clean up
rm data.yaml
```

### After Integration (Automated)
```bash
# One command does everything
./labours-go --from-repo /repo -m burndown-project -o chart.png
```

## Performance Tips

### For Large Repositories

```bash
# Use Protocol Buffer format (faster)
./scripts/analyze_with_hercules.sh /large/repo --pb

# Reduce time granularity
./labours-go --from-repo /large/repo \
    --hercules-flags "--granularity 5 --sampling 5" \
    -m burndown-project

# Use hibernation to reduce memory
./labours-go --from-repo /large/repo \
    --hercules-flags "--hibernation-distance 1000" \
    -m burndown-project
```

### For Multiple Repositories

```bash
# Parallel processing
for repo in repos/*/; do
    (
        ./labours-go --from-repo "$repo" \
            -m burndown-project \
            -o "results/$(basename $repo).png"
    ) &
done
wait
```

## Troubleshooting Examples

### Hercules Not Found

```bash
# Error: hercules binary not found
./labours-go --from-repo /repo -m burndown-project

# Solution: Specify path
./labours-go --hercules /custom/path/hercules --from-repo /repo -m burndown-project

# Or install to standard location
sudo cp hercules /usr/local/bin/hercules
```

### Large Repository Issues

```bash
# Memory issues with large repos
./labours-go --from-repo /huge/repo \
    --hercules-flags "--hibernation-distance 500 --granularity 20" \
    -m burndown-project
```

### No Data Available

```bash
# Some analyses need specific git history
./labours-go --from-repo /new/repo -m couples-files
# May not work on new repositories with few commits

# Solution: Use simpler analysis
./labours-go --from-repo /new/repo -m burndown-project
```

## Integration Success Summary

✅ **Direct Integration**: `--from-repo` flag for one-command analysis  
✅ **Auto-Detection**: Automatically finds hercules binary  
✅ **Multiple Scripts**: Quick analysis and comprehensive analysis scripts  
✅ **Theme Support**: All themes work with hercules integration  
✅ **Error Handling**: Graceful fallbacks and clear error messages  
✅ **Performance**: Protocol Buffer support for large repositories  
✅ **Documentation**: Complete integration guide and examples  

The hercules + labours-go integration is now **production-ready** and provides a seamless Git analytics pipeline!