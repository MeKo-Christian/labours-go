package graphics

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

// PlotStackedBurndown generates a proper stacked area chart for burndown analysis
func PlotStackedBurndown(matrix [][]float64, dateRange []time.Time, output string, relative bool) error {
	if len(matrix) == 0 || len(dateRange) == 0 {
		return fmt.Errorf("empty matrix or date range")
	}

	// Create plot
	p := plot.New()
	p.Title.Text = "Burndown Chart"
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Lines of Code"
	if relative {
		p.Y.Label.Text = "Relative Fraction"
	}

	// Ensure matrix dimensions are consistent
	numSeries := len(matrix)
	if numSeries == 0 {
		return fmt.Errorf("empty matrix")
	}

	numPoints := len(matrix[0])
	if numPoints != len(dateRange) {
		// Adjust date range or matrix to match
		minLen := int(math.Min(float64(numPoints), float64(len(dateRange))))
		numPoints = minLen
		dateRange = dateRange[:minLen]
	}

	// Convert dates to float64 for plotting (Unix timestamps)
	timeValues := make([]float64, numPoints)
	for i, date := range dateRange {
		timeValues[i] = float64(date.Unix())
	}

	// Generate cumulative data for stacking
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

	// Color palette for different series
	colors := generateColorPalette(numSeries)

	// Create stacked areas (bottom to top)
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

		// Create polygon for this stacked area
		if err := addStackedLayer(p, topPoints, bottomPoints, colors[i], fmt.Sprintf("Layer %d", i)); err != nil {
			return fmt.Errorf("error adding layer %d: %v", i, err)
		}
	}

	// Configure time axis
	p.X.Tick.Marker = &TimeTicker{Format: "2006-01-02"}

	// Set reasonable axis ranges
	if len(timeValues) > 0 {
		p.X.Min = timeValues[0]
		p.X.Max = timeValues[len(timeValues)-1]
	}

	// Save the plot
	width := 12 * vg.Inch
	height := 8 * vg.Inch
	if err := p.Save(width, height, output); err != nil {
		return fmt.Errorf("failed to save plot to %s: %v", output, err)
	}

	return nil
}

// addStackedLayer adds a filled area between top and bottom curves
func addStackedLayer(p *plot.Plot, top, bottom plotter.XYs, fillColor color.Color, label string) error {
	if len(top) != len(bottom) {
		return fmt.Errorf("top and bottom point arrays must have equal length")
	}

	// Create polygon points: top curve + reversed bottom curve
	points := make(plotter.XYs, len(top)+len(bottom))

	// Add top curve points
	copy(points[:len(top)], top)

	// Add bottom curve points in reverse order
	for i := range bottom {
		points[len(top)+i] = plotter.XY{X: bottom[len(bottom)-1-i].X, Y: bottom[len(bottom)-1-i].Y}
	}

	// Create polygon plotter
	polygon, err := plotter.NewPolygon(points)
	if err != nil {
		return fmt.Errorf("failed to create polygon: %v", err)
	}

	// Set fill color with some transparency
	polygon.Color = fillColor

	// Add to plot
	p.Add(polygon)

	// Add legend entry (just the top line for clarity)
	line, err := plotter.NewLine(top)
	if err == nil {
		line.Color = fillColor
		line.Width = vg.Points(2)
		p.Legend.Add(label, line)
	}

	return nil
}

// generateColorPalette creates a set of distinct colors for the chart
func generateColorPalette(n int) []color.Color {
	if n <= 0 {
		return []color.Color{}
	}

	// Use predefined colors for better visibility
	baseColors := []color.Color{
		color.RGBA{R: 31, G: 119, B: 180, A: 150},  // Blue
		color.RGBA{R: 255, G: 127, B: 14, A: 150},  // Orange
		color.RGBA{R: 44, G: 160, B: 44, A: 150},   // Green
		color.RGBA{R: 214, G: 39, B: 40, A: 150},   // Red
		color.RGBA{R: 148, G: 103, B: 189, A: 150}, // Purple
		color.RGBA{R: 140, G: 86, B: 75, A: 150},   // Brown
		color.RGBA{R: 227, G: 119, B: 194, A: 150}, // Pink
		color.RGBA{R: 127, G: 127, B: 127, A: 150}, // Gray
		color.RGBA{R: 188, G: 189, B: 34, A: 150},  // Olive
		color.RGBA{R: 23, G: 190, B: 207, A: 150},  // Cyan
	}

	colors := make([]color.Color, n)
	for i := 0; i < n; i++ {
		if i < len(baseColors) {
			colors[i] = baseColors[i]
		} else {
			// Generate additional colors using HSV
			colors[i] = generateHSVColor(i, n)
		}
	}

	return colors
}

// generateHSVColor generates colors using HSV color space for better distribution
func generateHSVColor(index, total int) color.Color {
	// Use golden angle for better color distribution
	goldenAngle := 137.508 // degrees
	hue := math.Mod(float64(index)*goldenAngle, 360)

	// Convert HSV to RGB
	saturation := 0.7
	value := 0.9

	c := value * saturation
	x := c * (1 - math.Abs(math.Mod(hue/60, 2)-1))
	m := value - c

	var r, g, b float64
	switch {
	case hue < 60:
		r, g, b = c, x, 0
	case hue < 120:
		r, g, b = x, c, 0
	case hue < 180:
		r, g, b = 0, c, x
	case hue < 240:
		r, g, b = 0, x, c
	case hue < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return color.RGBA{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
		A: 180, // Semi-transparent
	}
}

// TimeTicker implements plot.Ticker for time-based axes
type TimeTicker struct {
	Format string
}

// Ticks generates tick marks for time axis
func (ticker *TimeTicker) Ticks(min, max float64) []plot.Tick {
	if ticker.Format == "" {
		ticker.Format = "2006-01-02"
	}

	start := time.Unix(int64(min), 0)
	end := time.Unix(int64(max), 0)
	duration := end.Sub(start)

	var interval time.Duration
	var majorTicks []plot.Tick

	// Determine appropriate tick interval based on time range
	switch {
	case duration <= 24*time.Hour:
		interval = time.Hour
	case duration <= 7*24*time.Hour:
		interval = 24 * time.Hour
	case duration <= 30*24*time.Hour:
		interval = 7 * 24 * time.Hour
	case duration <= 365*24*time.Hour:
		interval = 30 * 24 * time.Hour
	default:
		interval = 365 * 24 * time.Hour
	}

	// Generate major ticks
	for t := start.Truncate(interval); t.Before(end) || t.Equal(end); t = t.Add(interval) {
		if t.Unix() >= int64(min) && t.Unix() <= int64(max) {
			majorTicks = append(majorTicks, plot.Tick{
				Value: float64(t.Unix()),
				Label: t.Format(ticker.Format),
			})
		}
	}

	return majorTicks
}

// PlotHeatmap generates a heatmap visualization (placeholder for future ownership/overwrites charts)
func PlotHeatmap(matrix [][]float64, rowLabels, colLabels []string, output string, title string) error {
	p := plot.New()
	p.Title.Text = title

	// This would be implemented with a proper heatmap plotter
	// For now, return a placeholder implementation
	return fmt.Errorf("heatmap plotting not yet implemented")
}

// PlotBarChart generates a bar chart (for developer statistics, language stats, etc.)
func PlotBarChart(values []float64, labels []string, output string, title string) error {
	if len(values) != len(labels) {
		return fmt.Errorf("values and labels must have the same length")
	}

	p := plot.New()
	p.Title.Text = title
	p.Y.Label.Text = "Value"

	// Create bar chart data
	bars := make(plotter.Values, len(values))
	for i, v := range values {
		bars[i] = v
	}

	// Create bar chart
	barChart, err := plotter.NewBarChart(bars, vg.Points(20))
	if err != nil {
		return fmt.Errorf("error creating bar chart: %v", err)
	}

	barChart.Color = plotutil.Color(0)
	p.Add(barChart)

	// Set custom x-axis labels
	p.NominalX(labels...)

	// Save plot
	if err := p.Save(10*vg.Inch, 6*vg.Inch, output); err != nil {
		return fmt.Errorf("failed to save bar chart: %v", err)
	}

	return nil
}
