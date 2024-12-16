package modes

import (
	"fmt"
	"labours-go/internal/graphics"
	"os"
	"path/filepath"
	"time"

	"github.com/schollz/progressbar/v3"
)

// generateBurndownPlot creates the burndown plot with stacking, resampling, and survival ratio output.
func generateBurndownPlot(name string, matrix [][]int, output string, relative bool, startTime, endTime *time.Time, resample string) error {
	fmt.Println("Running: burndown-project")

	// Validate output path
	if output == "" {
		output = fmt.Sprintf("burndown_%s.png", name)
		fmt.Printf("Output not provided, using default: %s\n", output)
	}

	outputDir := filepath.Dir(output)
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory %s: %v", outputDir, err)
	}

	// Resampling logic
	if resample == "" {
		resample = "year"
	}
	fmt.Printf("resampling to %s, please wait...\n", resample)

	// Use default endTime if not provided
	if endTime == nil {
		now := time.Now()
		endTime = &now
	}

	// Use earliest time in the matrix if startTime is not provided
	if startTime == nil {
		tickSize := time.Duration(365*24) * time.Hour // Assuming yearly granularity by default
		if resample == "month" {
			tickSize = time.Duration(30*24) * time.Hour
		} else if resample == "day" {
			tickSize = 24 * time.Hour
		}
		earliest := findEarliestTime(matrix, tickSize, *endTime)
		startTime = &earliest
	}

	// Interpolation with progress bar
	interpolatedMatrix, dateRange := interpolateBurndownMatrix(matrix, *startTime, *endTime, resample)

	// Survival analysis
	survivalRatios := calculateSurvivalRatios(interpolatedMatrix, *startTime)
	printSurvivalRatios(survivalRatios)

	// Normalize if relative is true
	if relative {
		interpolatedMatrix = normalizeMatrix(interpolatedMatrix)
	}

	// Create plot
	if err := graphics.PlotStackedBurndown(interpolatedMatrix, dateRange, output, relative); err != nil {
		return fmt.Errorf("error creating burndown plot: %v", err)
	}

	fmt.Printf("Chart saved to %s\n", output)
	return nil
}

// ResampleDateRange creates a date range based on the given resampling interval.
func resampleDateRange(start, end time.Time, resample string) []time.Time {
	var step time.Duration
	switch resample {
	case "year":
		step = 365 * 24 * time.Hour
	case "month":
		step = 30 * 24 * time.Hour
	case "day":
		step = 24 * time.Hour
	default:
		step = 24 * time.Hour
	}

	var dates []time.Time
	for t := start; t.Before(end); t = t.Add(step) {
		dates = append(dates, t)
	}
	return dates
}

// interpolateBurndownMatrix interpolates the matrix and shows progress.
func interpolateBurndownMatrix(matrix [][]int, startTime, endTime time.Time, resample string) ([][]float64, []time.Time) {
	granularity := 1 // Assume 1-day granularity
	sampling := 1    // Default to 1-day sampling for now
	numBands := len(matrix)
	numTicks := len(matrix[0])

	// Calculate the total interpolated size
	daily := make([][]float64, numBands*granularity)
	for i := range daily {
		daily[i] = make([]float64, numTicks*sampling)
	}

	dateRange := resampleDateRange(startTime, endTime, resample)

	bar := progressbar.Default(int64(numBands), "Interpolating data")
	for y := 0; y < numBands; y++ {
		bar.Add(1)
		for x := 0; x < numTicks; x++ {
			for i := y * granularity; i < (y+1)*granularity; i++ {
				for j := x * sampling; j < (x+1)*sampling; j++ {
					daily[i][j] = float64(matrix[y][x])
				}
			}
		}
	}

	return daily, dateRange
}

// calculateSurvivalRatios computes survival ratios for the matrix.
func calculateSurvivalRatios(matrix [][]float64, startTime time.Time) map[int]float64 {
	survival := make(map[int]float64)
	total := 0.0

	for i := range matrix[0] { // Iterate over columns (time ticks)
		alive := 0.0
		for _, row := range matrix {
			if row[i] > 0 {
				alive += row[i]
			}
		}
		total += alive
		survival[i] = alive / total
	}

	return survival
}

// printSurvivalRatios prints the survival ratios to mimic the Python output.
func printSurvivalRatios(survival map[int]float64) {
	fmt.Println("           Ratio of survived lines")
	for days, ratio := range survival {
		fmt.Printf("%d days\t\t%.6f\n", days, ratio)
	}
}

// normalizeMatrix normalizes each column to sum to 1.
func normalizeMatrix(matrix [][]float64) [][]float64 {
	for j := 0; j < len(matrix[0]); j++ {
		sum := 0.0
		for i := 0; i < len(matrix); i++ {
			sum += matrix[i][j]
		}
		if sum == 0 {
			continue
		}
		for i := 0; i < len(matrix); i++ {
			matrix[i][j] /= sum
		}
	}
	return matrix
}

// findEarliestTime determines the earliest non-zero time entry in the data matrix.
func findEarliestTime(matrix [][]int, tickSize time.Duration, endTime time.Time) time.Time {
	for rowIndex, row := range matrix {
		for colIndex, val := range row {
			if val > 0 {
				// Calculate the corresponding time for this column
				earliestTime := endTime.Add(-tickSize * time.Duration(len(row)-colIndex))
				fmt.Printf("Earliest time found at row %d, col %d: %s\n", rowIndex, colIndex, earliestTime)
				return earliestTime
			}
		}
	}
	return time.Unix(0, 0) // Fallback, should never hit if matrix has data
}
