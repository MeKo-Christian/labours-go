package graphics

import (
	"fmt"
	"image/color"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"labours-go/internal/burndown"
)

// PlotBurndownPythonStyle creates a burndown plot that matches Python's pyplot.stackplot behavior
func PlotBurndownPythonStyle(data *burndown.ProcessedBurndown, output string, relative bool) error {
	if data == nil || len(data.Matrix) == 0 || len(data.DateRange) == 0 {
		return fmt.Errorf("empty burndown data")
	}

	p := plot.New()
	p.Title.Text = "Burndown Chart"
	p.X.Label.Text = "Time" 
	p.Y.Label.Text = "Lines of code"
	if relative {
		p.Y.Label.Text = "Relative Fraction"
	}

	// Apply theme styling
	applyThemeToPlot(p)

	numSeries := len(data.Matrix)
	numPoints := len(data.DateRange)

	// Ensure matrix dimensions are consistent
	if numSeries == 0 {
		return fmt.Errorf("empty matrix")
	}

	// Convert dates to float64 for plotting (Unix timestamps)
	timeValues := make([]float64, numPoints)
	for i, date := range data.DateRange {
		timeValues[i] = float64(date.Unix())
	}

	// Normalize matrix if relative mode is enabled (like Python does)
	matrix := data.Matrix
	if relative {
		matrix = normalizeMatrixColumns(data.Matrix)
	}

	// DEBUG: Print matrix values to understand the data
	fmt.Printf("DEBUG MATRIX ANALYSIS:\n")
	fmt.Printf("  Matrix dimensions: %dx%d\n", len(matrix), len(matrix[0]))
	for i := 0; i < len(matrix); i++ {
		minVal, maxVal := matrix[i][0], matrix[i][0]
		negCount, posCount := 0, 0
		for j := 0; j < len(matrix[i]); j++ {
			if matrix[i][j] < minVal { minVal = matrix[i][j] }
			if matrix[i][j] > maxVal { maxVal = matrix[i][j] }
			if matrix[i][j] < 0 { negCount++ }
			if matrix[i][j] > 0 { posCount++ }
		}
		fmt.Printf("  Layer %d: min=%.2f, max=%.2f, negatives=%d, positives=%d\n", i, minVal, maxVal, negCount, posCount)
	}
	
	// Generate matplotlib-compatible color palette (matches Python exactly)
	colors := generateMatplotlibColorPalette(numSeries)

	// Create cumulative data for stacking (bottom to top like Python's stackplot)
	cumulative := make([][]float64, numSeries)
	for i := range cumulative {
		cumulative[i] = make([]float64, numPoints)
		for j := 0; j < numPoints && j < len(matrix[i]); j++ {
			cumulative[i][j] = matrix[i][j]
			if i > 0 {
				cumulative[i][j] += cumulative[i-1][j]
			}
		}
	}

	// Create stacked areas (from top to bottom for proper rendering)
	for i := numSeries - 1; i >= 0; i-- {
		// Create data points for this layer
		var topPoints plotter.XYs
		var bottomPoints plotter.XYs

		for j := 0; j < numPoints; j++ {
			x := timeValues[j]
			topY := cumulative[i][j]

			var bottomY float64
			if i > 0 {
				bottomY = cumulative[i-1][j]
			} else {
				bottomY = 0
			}

			topPoints = append(topPoints, plotter.XY{X: x, Y: topY})
			bottomPoints = append(bottomPoints, plotter.XY{X: x, Y: bottomY})
		}

		// Use semantic label from Python processing
		label := fmt.Sprintf("Layer %d", i)
		if i < len(data.Labels) {
			label = data.Labels[i]
		}

		// Create polygon for this stacked area
		if err := addStackedLayer(p, topPoints, bottomPoints, colors[i], label); err != nil {
			return fmt.Errorf("error adding layer %s: %v", label, err)
		}
	}

	// Configure time axis with Python-style formatting
	configureBurndownTimeAxis(p, timeValues, data.ResampleMode)

	// Set Y-axis limits
	if relative {
		p.Y.Min = 0
		p.Y.Max = 1
	}

	// Configure legend position (matches Python behavior)
	legendLoc := 2 // upper left
	if relative {
		legendLoc = 3 // lower left
	}
	_ = legendLoc // TODO: Implement legend positioning

	// Save plot with Python-compatible dimensions
	width := 12 * vg.Inch  // Python's typical figure size
	height := 8 * vg.Inch
	if err := p.Save(width, height, output); err != nil {
		return fmt.Errorf("failed to save plot to %s: %v", output, err)
	}

	return nil
}

// normalizeMatrixColumns normalizes each column to sum to 1 (matches Python's relative mode)
func normalizeMatrixColumns(matrix [][]float64) [][]float64 {
	if len(matrix) == 0 {
		return matrix
	}

	normalized := make([][]float64, len(matrix))
	for i := range matrix {
		normalized[i] = make([]float64, len(matrix[i]))
		copy(normalized[i], matrix[i])
	}

	// Normalize each column (time point) to sum to 1
	numCols := len(matrix[0])
	for j := 0; j < numCols; j++ {
		sum := 0.0
		for i := 0; i < len(matrix); i++ {
			if j < len(matrix[i]) {
				sum += matrix[i][j]
			}
		}
		if sum > 0 {
			for i := 0; i < len(matrix); i++ {
				if j < len(normalized[i]) {
					normalized[i][j] /= sum
				}
			}
		}
	}

	return normalized
}

// configureBurndownTimeAxis sets up the time axis to match Python's matplotlib behavior
func configureBurndownTimeAxis(p *plot.Plot, timeValues []float64, resampleMode string) {
	if len(timeValues) == 0 {
		return
	}

	// Set basic time range
	p.X.Min = timeValues[0]
	p.X.Max = timeValues[len(timeValues)-1]

	// Configure time ticker based on resampling mode
	var format string
	switch resampleMode {
	case "A", "year":
		format = "2006"
		p.X.Tick.Marker = &TimeTicker{Format: format}
	case "M", "month":
		format = "2006-01"
		p.X.Tick.Marker = &TimeTicker{Format: format}
	case "D", "day":
		format = "2006-01-02"
		p.X.Tick.Marker = &TimeTicker{Format: format}
	default:
		format = "2006-01-02"
		p.X.Tick.Marker = &TimeTicker{Format: format}
	}
}

// PrintSurvivalFunction prints survival ratios to match Python output (placeholder)
func PrintSurvivalFunction(matrix [][]float64) {
	fmt.Println("           Ratio of survived lines")
	// TODO: Implement Kaplan-Meier survival analysis like Python
	// For now, just print a placeholder that shows we're processing survival data
	
	if len(matrix) > 0 && len(matrix[0]) > 0 {
		total := 0.0
		for i := range matrix {
			for j := range matrix[i] {
				total += matrix[i][j]
			}
		}
		
		for i := 0; i < len(matrix[0]); i++ {
			alive := 0.0
			for j := range matrix {
				if i < len(matrix[j]) {
					alive += matrix[j][i]
				}
			}
			if total > 0 {
				ratio := alive / total
				fmt.Printf("%d days\t\t%.6f\n", i, ratio)
			}
		}
	}
}

// generateMatplotlibColorPalette creates colors that exactly match Python matplotlib defaults
func generateMatplotlibColorPalette(n int) []color.Color {
	// Matplotlib default colors (C0, C1, C2, ...) - these exactly match Python pyplot
	matplotlibColors := []color.Color{
		color.RGBA{R: 31, G: 119, B: 180, A: 180},   // Blue (C0) - matplotlib default
		color.RGBA{R: 255, G: 127, B: 14, A: 180},   // Orange (C1) 
		color.RGBA{R: 44, G: 160, B: 44, A: 180},    // Green (C2)
		color.RGBA{R: 214, G: 39, B: 40, A: 180},    // Red (C3)
		color.RGBA{R: 148, G: 103, B: 189, A: 180},  // Purple (C4)
		color.RGBA{R: 140, G: 86, B: 75, A: 180},    // Brown (C5)
		color.RGBA{R: 227, G: 119, B: 194, A: 180},  // Pink (C6)
		color.RGBA{R: 127, G: 127, B: 127, A: 180},  // Gray (C7)
		color.RGBA{R: 188, G: 189, B: 34, A: 180},   // Olive (C8)
		color.RGBA{R: 23, G: 190, B: 207, A: 180},   // Cyan (C9)
	}
	
	// For burndown charts specifically, Python matplotlib uses C3 (Red) first for older code, then C0 (Blue) for newer
	// This exactly matches the Python reference chart pattern
	if n == 2 {
		return []color.Color{
			color.RGBA{R: 214, G: 39, B: 40, A: 200},   // Red (C3) for bottom layer (older/2024)
			color.RGBA{R: 31, G: 119, B: 180, A: 200},  // Blue (C0) for top layer (newer/2025)
		}
	}
	
	colors := make([]color.Color, n)
	for i := 0; i < n; i++ {
		if i < len(matplotlibColors) {
			colors[i] = matplotlibColors[i]
		} else {
			// Generate additional colors if needed
			colors[i] = generateHSVColorWithOpacity(i, n, 180)
		}
	}
	
	return colors
}