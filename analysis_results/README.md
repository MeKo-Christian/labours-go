# Analysis Results Directory

All Git analytics outputs are organized in this centralized directory.

## 📊 **Main Analysis Results**

### Project Burndown Charts
- **`burndown_project.png`** - Project-level code evolution over time
- **`burndown_burndown-project.png`** - Alternative burndown visualization  
- **`hercules_integration_test.png`** - Integration test result (burndown chart)
- **`hercules_burndown_chart.png`** - Chart generated from hercules data

### Developer Analysis  
- **`devs_devs.png`** - Developer contribution statistics

## 📁 **Comprehensive Analysis**

The `labours_go_comprehensive/` directory contains:
- **`burndown_project.png`** - Main project burndown chart
- **`burndown.yaml`** - Raw hercules burndown data
- **`devs.yaml`** - Raw hercules developer data
- Individual file-level burndown charts (in subdirectories)

## 🚀 **How These Were Generated**

### Direct CLI Integration
```bash
./labours-go --from-repo . -m burndown-project -o analysis_results/
```

### Quick Analysis Script  
```bash
./scripts/quick_analysis.sh . analysis_results/labours_go_comprehensive
```

### Manual Process
```bash
hercules --burndown . > data.yaml
./labours-go -i data.yaml -m burndown-project -o analysis_results/chart.png
```

## 🎯 **Key Files to View**

**Start with these main visualizations:**
1. **`burndown_project.png`** - Shows code evolution over ~7 years
2. **`labours_go_comprehensive/burndown_project.png`** - Comprehensive analysis
3. **`devs_devs.png`** - Developer contribution patterns

## 📈 **Chart Types Explained**

### Burndown Charts
- **X-axis**: Time (years from 2017-2024)
- **Y-axis**: Lines of code
- **Colors**: Different code age bands
  - Blue: New code (recently added)
  - Orange: Modified code 
  - Gradients: Code of different ages

### Developer Charts
- Bar charts showing commits, lines added/removed per developer
- Activity patterns over time

## 🛠 **Generate New Analysis**

To create fresh analysis results:

```bash
# Quick analysis (recommended)
./scripts/quick_analysis.sh /path/to/repo analysis_results/new_analysis

# Direct integration
./labours-go --from-repo /path/to/repo -m burndown-project,devs -o analysis_results/

# Custom themes
./labours-go --from-repo . --theme dark -m burndown-project -o analysis_results/dark_theme.png
```

## 📂 **Directory Structure**

```
analysis_results/
├── README.md (this file)
├── burndown_project.png                    # Main project evolution chart
├── hercules_integration_test.png           # Integration test result  
├── hercules_burndown_chart.png            # Sample chart from hercules data
├── devs_devs.png                          # Developer statistics
└── labours_go_comprehensive/              # Comprehensive analysis folder
    ├── burndown.yaml                      # Raw hercules data
    ├── devs.yaml                          # Raw developer data  
    └── burndown_project.png               # Main burndown chart
```

All analysis results from the hercules + labours-go integration are now centralized here! 🎉