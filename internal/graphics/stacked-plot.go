package graphics

import (
	"fmt"
	"image/color"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// PlotStackedBurndown generates the stacked area chart using filled areas.
func PlotStackedBurndown(matrix [][]float64, dateRange []time.Time, output string, relative bool) error {
	p := plot.New()
	p.Title.Text = "Burndown Chart"
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Lines of Code"
	if relative {
		p.Y.Label.Text = "Relative Fraction"
	}

	// Generate cumulative sum for stacked plotting
	cumulative := make([][]float64, len(matrix))
	for i := range matrix {
		cumulative[i] = make([]float64, len(matrix[i]))
		copy(cumulative[i], matrix[i])
		if i > 0 {
			for j := range cumulative[i] {
				cumulative[i][j] += cumulative[i-1][j]
			}
		}
	}

	// Create filled areas for each band
	for i := range cumulative {
		// Define the base and top for the stacked layer
		base := make(plotter.XYs, len(cumulative[i]))
		top := make(plotter.XYs, len(cumulative[i]))
		for j := range cumulative[i] {
			x := float64(dateRange[j].Unix())
			base[j].X = x
			top[j].X = x
			if i == 0 {
				base[j].Y = 0 // Bottom layer starts from 0
			} else {
				base[j].Y = cumulative[i-1][j]
			}
			top[j].Y = cumulative[i][j]
		}

		// Combine base and top to create a polygon
		points := make(plotter.XYs, 2*len(base))
		copy(points, top) // Add top points
		for j := len(base) - 1; j >= 0; j-- {
			points[len(base)+j] = base[j] // Add base points in reverse
		}

		// Create a polygon for the layer
		polygon, err := plotter.NewPolygon(points)
		if err != nil {
			return fmt.Errorf("error creating polygon: %v", err)
		}
		polygon.Color = color.RGBA{
			R: uint8(50 * i % 255),
			G: uint8(100 * i % 255),
			B: uint8(150 * i % 255),
			A: 100, // Set transparency
		}
		p.Add(polygon)
	}

	// Format x-axis as time
	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02"}

	// Save the plot to a file
	if err := p.Save(10*vg.Inch, 5*vg.Inch, output); err != nil {
		return fmt.Errorf("failed to save plot: %v", err)
	}

	return nil
}
