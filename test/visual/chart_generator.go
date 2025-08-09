package visual

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"labours-go/internal/modes"
	"labours-go/internal/readers"
)

// ChartGenerator handles chart generation for visual testing
type ChartGenerator struct {
	OutputDir string
}

// NewChartGenerator creates a new chart generator instance
func NewChartGenerator(outputDir string) *ChartGenerator {
	return &ChartGenerator{
		OutputDir: outputDir,
	}
}

// GenerateChart creates a chart using the specified mode and input data
func (cg *ChartGenerator) GenerateChart(t *testing.T, mode, inputFile string) (string, error) {
	t.Helper()
	
	// Ensure output directory exists
	if err := os.MkdirAll(cg.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// Create output file path
	outputPath := filepath.Join(cg.OutputDir, fmt.Sprintf("test_%s.png", mode))
	
	// Read input data - auto-detect format
	var reader readers.Reader
	if filepath.Ext(inputFile) == ".yaml" || filepath.Ext(inputFile) == ".yml" {
		reader = &readers.YamlReader{}
	} else {
		reader = &readers.ProtobufReader{}
	}
	
	file, err := os.Open(inputFile)
	if err != nil {
		return "", fmt.Errorf("failed to open input file %s: %w", inputFile, err)
	}
	defer file.Close()
	
	err = reader.Read(file)
	if err != nil {
		return "", fmt.Errorf("failed to read input data: %w", err)
	}
	
	// Generate chart based on mode
	switch mode {
	case "burndown-project":
		err = cg.generateBurndownProject(reader, outputPath, false)
	case "burndown-project-relative":
		err = cg.generateBurndownProject(reader, outputPath, true)
	case "burndown-file":
		err = cg.generateBurndownFile(reader, outputPath)
	case "burndown-person":
		err = cg.generateBurndownPerson(reader, outputPath)
	case "ownership":
		err = cg.generateOwnership(reader, outputPath)
	case "devs":
		err = cg.generateDevs(reader, outputPath)
	case "couples-people":
		err = cg.generateCouplesPeople(reader, outputPath)
	case "couples-files":
		err = cg.generateCouplesFiles(reader, outputPath)
	default:
		return "", fmt.Errorf("unsupported chart mode: %s", mode)
	}
	
	if err != nil {
		return "", fmt.Errorf("failed to generate %s chart: %w", mode, err)
	}
	
	// Verify output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("chart file was not created: %s", outputPath)
	}
	
	return outputPath, nil
}

// generateBurndownProject creates a project burndown chart
func (cg *ChartGenerator) generateBurndownProject(reader readers.Reader, outputPath string, relative bool) error {
	// Set viper config for relative mode
	viper.Set("relative", relative)
	viper.Set("resample", "year") // Default resampling for consistency
	
	// Call the actual burndown project generation using Python-compatible version
	return modes.GenerateBurndownProjectPython(reader, outputPath, relative, "year")
}

// generateBurndownFile creates file-level burndown charts
func (cg *ChartGenerator) generateBurndownFile(reader readers.Reader, outputPath string) error {
	// Use Python-compatible file burndown generation
	viper.Set("relative", false) // Default to absolute
	viper.Set("resample", "year")
	return modes.GenerateBurndownFilePython(reader, outputPath, false, "year")
}

// generateBurndownPerson creates person-level burndown charts
func (cg *ChartGenerator) generateBurndownPerson(reader readers.Reader, outputPath string) error {
	// Use regular burndown person function with nil time parameters for defaults
	return modes.BurndownPerson(reader, outputPath, false, nil, nil, "year")
}

// generateOwnership creates code ownership visualization
func (cg *ChartGenerator) generateOwnership(reader readers.Reader, outputPath string) error {
	// Call the ownership mode
	return modes.OwnershipBurndown(reader, outputPath)
}

// generateDevs creates developer statistics visualization
func (cg *ChartGenerator) generateDevs(reader readers.Reader, outputPath string) error {
	// Call the devs mode with default max people (20)
	return modes.Devs(reader, outputPath, 20)
}

// generateCouplesPeople creates people coupling visualization
func (cg *ChartGenerator) generateCouplesPeople(reader readers.Reader, outputPath string) error {
	// Call the couples-people mode
	return modes.CouplesPeople(reader, outputPath)
}

// generateCouplesFiles creates file coupling visualization
func (cg *ChartGenerator) generateCouplesFiles(reader readers.Reader, outputPath string) error {
	// Call the couples-files mode
	return modes.CouplesFiles(reader, outputPath)
}

// GenerateReferenceSet creates a complete set of reference images for golden file testing
func (cg *ChartGenerator) GenerateReferenceSet(t *testing.T, inputFile string) map[string]string {
	t.Helper()
	
	generatedFiles := make(map[string]string)
	
	// List of modes to generate reference images for
	modes := []string{
		"burndown-project",
		"burndown-project-relative", 
		"ownership",
		"devs",
	}
	
	for _, mode := range modes {
		outputPath, err := cg.GenerateChart(t, mode, inputFile)
		if err != nil {
			t.Logf("Warning: Failed to generate reference for %s: %v", mode, err)
			continue
		}
		
		generatedFiles[mode] = outputPath
		t.Logf("✅ Generated reference image for %s: %s", mode, outputPath)
	}
	
	return generatedFiles
}

// ValidateChartStructure performs structural validation on a generated chart
func (cg *ChartGenerator) ValidateChartStructure(t *testing.T, chartPath string) error {
	t.Helper()
	
	// Load the chart image
	img, err := loadImage(chartPath)
	if err != nil {
		return fmt.Errorf("failed to load chart image: %w", err)
	}
	
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	
	// Basic structural validations
	if width < 400 || height < 200 {
		return fmt.Errorf("chart dimensions too small: %dx%d", width, height)
	}
	
	if width > 4000 || height > 4000 {
		return fmt.Errorf("chart dimensions too large: %dx%d", width, height)
	}
	
	// Check for reasonable aspect ratio (should be wider than tall for most charts)
	aspectRatio := float64(width) / float64(height)
	if aspectRatio < 0.5 || aspectRatio > 5.0 {
		t.Logf("Warning: Unusual aspect ratio: %.2f (width/height)", aspectRatio)
	}
	
	// Validate color usage
	histogram := buildColorHistogram(img)
	
	// Check for sufficient color diversity (should have multiple distinct colors)
	if len(histogram) < 5 {
		return fmt.Errorf("chart has too few colors (%d), may be incorrectly rendered", len(histogram))
	}
	
	// Look for pure white/black dominance (may indicate rendering issues)
	whitePixels := histogram["248,248,248"] + histogram["255,255,255"]
	blackPixels := histogram["0,0,0"] + histogram["8,8,8"]
	
	if whitePixels > 0.9 {
		return fmt.Errorf("chart is mostly white (%.1f%%), may be empty", whitePixels*100)
	}
	
	if blackPixels > 0.9 {
		return fmt.Errorf("chart is mostly black (%.1f%%), may have rendering issues", blackPixels*100)
	}
	
	t.Logf("Chart structure validation passed: %dx%d, %d colors, %.1f%% white", 
		width, height, len(histogram), whitePixels*100)
	
	return nil
}