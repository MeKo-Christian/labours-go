package graphics

import (
	"image/color"
	"math"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// CustomPalette represents a mapping of values to a predefined set of colors.
type CustomPalette struct {
	Colors []color.Color
	Min    float64
	Max    float64
}

// At maps a value to a corresponding color in the palette.
func (p *CustomPalette) At(value float64) color.Color {
	// Normalize the value to the range [0, 1].
	normalized := (value - p.Min) / (p.Max - p.Min)
	if normalized < 0 {
		normalized = 0
	} else if normalized > 1 {
		normalized = 1
	}

	// Scale the normalized value to the palette size.
	index := int(math.Round(normalized * float64(len(p.Colors)-1)))
	return p.Colors[index]
}

// HeatMap represents a heatmap plotter for a 2D matrix.
type HeatMap struct {
	Matrix  [][]float64
	Rows    []string
	Cols    []string
	Palette *CustomPalette
}

// NewHeatMap creates a new HeatMap with a custom palette.
func NewHeatMap(matrix [][]float64, rows, cols []string, palette *CustomPalette) *HeatMap {
	return &HeatMap{
		Matrix:  matrix,
		Rows:    rows,
		Cols:    cols,
		Palette: palette,
	}
}

// Plot draws the heatmap onto the plot canvas.
func (hm *HeatMap) Plot(c draw.Canvas, p *plot.Plot) {
	r := c.Rectangle.Size()
	cellWidth := r.X / vg.Length(len(hm.Cols))
	cellHeight := r.Y / vg.Length(len(hm.Rows))

	for rowIdx, row := range hm.Matrix {
		for colIdx, value := range row {
			x := vg.Length(colIdx) * cellWidth
			y := vg.Length(len(hm.Rows)-1-rowIdx) * cellHeight // Invert rows for correct orientation

			// Map value to a color using the custom palette.
			clr := hm.Palette.At(value)

			// Define the coordinates for the cell.
			xMin := c.Rectangle.Min.X + x
			xMax := xMin + cellWidth
			yMin := c.Rectangle.Min.Y + y
			yMax := yMin + cellHeight

			// Create a path for the rectangle.
			path := vg.Path{
				{Type: vg.MoveComp, Pos: vg.Point{X: xMin, Y: yMin}},
				{Type: vg.LineComp, Pos: vg.Point{X: xMax, Y: yMin}},
				{Type: vg.LineComp, Pos: vg.Point{X: xMax, Y: yMax}},
				{Type: vg.LineComp, Pos: vg.Point{X: xMin, Y: yMax}},
				{Type: vg.CloseComp},
			}

			// Set the fill color and fill the rectangle.
			c.SetColor(clr)
			c.Fill(path)
		}
	}
}

// DataRange returns the minimum and maximum data range of the heatmap.
func (hm *HeatMap) DataRange() (xmin, xmax, ymin, ymax float64) {
	xmin, ymin = 0, 0
	xmax = float64(len(hm.Cols))
	ymax = float64(len(hm.Rows))
	return
}
