package graphics

import (
	"image/color"
)

// ColorPalette is the legacy color palette for backwards compatibility
// It's updated automatically when themes are changed
var ColorPalette = []color.Color{
	color.RGBA{R: 31, G: 119, B: 180, A: 255},  // Blue
	color.RGBA{R: 255, G: 127, B: 14, A: 255},  // Orange
	color.RGBA{R: 44, G: 160, B: 44, A: 255},   // Green
	color.RGBA{R: 214, G: 39, B: 40, A: 255},   // Red
	color.RGBA{R: 148, G: 103, B: 189, A: 255}, // Purple
	color.RGBA{R: 140, G: 86, B: 75, A: 255},   // Brown
	color.RGBA{R: 227, G: 119, B: 194, A: 255}, // Pink
	color.RGBA{R: 127, G: 127, B: 127, A: 255}, // Gray
	color.RGBA{R: 188, G: 189, B: 34, A: 255},  // Olive
	color.RGBA{R: 23, G: 190, B: 207, A: 255},  // Cyan
}

// HeatColor generates a color ranging from blue (cold, ratio=0) to red (hot, ratio=1)
// This is useful for heatmap-style visualizations where higher values should appear "hotter"
// This function now uses the current theme's heat color settings
func HeatColor(ratio float64) color.Color {
	return CurrentTheme.GetHeatColor(ratio)
}

// GetColor returns a color from the current theme's palette by index
func GetColor(index int) color.Color {
	palette := CurrentTheme.GetColorPalette()
	if len(palette) == 0 {
		// Fallback to default if somehow no colors are available
		return color.RGBA{R: 100, G: 100, B: 100, A: 255}
	}
	return palette[index%len(palette)]
}

// GetColorPalette returns the current theme's color palette
func GetColorPalette() []color.Color {
	return CurrentTheme.GetColorPalette()
}
