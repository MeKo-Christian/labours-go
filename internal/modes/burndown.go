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

// resampleDateRange creates a date range based on the given resampling interval.
func resampleDateRange(start, end time.Time, resample string) []time.Time {
	var dates []time.Time
	
	switch resample {
	case "year":
		// Yearly samples - start of each year
		for year := start.Year(); year <= end.Year(); year++ {
			yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, start.Location())
			if yearStart.After(end) {
				break
			}
			if yearStart.After(start) || yearStart.Equal(start) {
				dates = append(dates, yearStart)
			}
		}
		
	case "month", "M":
		// Monthly samples - start of each month
		current := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
		if current.Before(start) {
			current = current.AddDate(0, 1, 0)
		}
		
		for current.Before(end) || current.Equal(end) {
			dates = append(dates, current)
			current = current.AddDate(0, 1, 0)
		}
		
	case "week", "W":
		// Weekly samples - start of each week (Monday)
		// Find the first Monday on or after start
		current := start
		for current.Weekday() != time.Monday {
			current = current.AddDate(0, 0, 1)
		}
		
		for current.Before(end) || current.Equal(end) {
			dates = append(dates, current)
			current = current.AddDate(0, 0, 7)
		}
		
	case "day", "D":
		// Daily samples
		for current := start; current.Before(end) || current.Equal(end); current = current.AddDate(0, 0, 1) {
			dates = append(dates, current)
		}
		
	default:
		// Default to daily sampling
		for current := start; current.Before(end) || current.Equal(end); current = current.AddDate(0, 0, 1) {
			dates = append(dates, current)
		}
	}
	
	// Ensure we have at least two points for interpolation
	if len(dates) == 0 {
		dates = append(dates, start, end)
	} else if len(dates) == 1 {
		if !dates[0].Equal(end) {
			dates = append(dates, end)
		}
	}
	
	return dates
}

// interpolateBurndownMatrix interpolates and resamples the matrix according to the specified interval
func interpolateBurndownMatrix(matrix [][]int, startTime, endTime time.Time, resample string) ([][]float64, []time.Time) {
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return [][]float64{}, []time.Time{}
	}

	numBands := len(matrix)
	originalTicks := len(matrix[0])

	// Generate the target date range based on resampling
	dateRange := resampleDateRange(startTime, endTime, resample)
	targetTicks := len(dateRange)

	// Create interpolated matrix
	interpolated := make([][]float64, numBands)
	for i := range interpolated {
		interpolated[i] = make([]float64, targetTicks)
	}

	bar := progressbar.Default(int64(numBands), "Interpolating burndown data")
	
	// Interpolate each band (developer/file/etc)
	for band := 0; band < numBands; band++ {
		bar.Add(1)
		
		// If target resolution matches original, direct copy
		if targetTicks == originalTicks {
			for tick := 0; tick < originalTicks; tick++ {
				interpolated[band][tick] = float64(matrix[band][tick])
			}
			continue
		}

		// Interpolate between original data points
		for targetTick := 0; targetTick < targetTicks; targetTick++ {
			// Map target tick to original tick space
			originalPos := float64(targetTick) * float64(originalTicks-1) / float64(targetTicks-1)
			
			// Find surrounding original ticks
			leftTick := int(originalPos)
			rightTick := leftTick + 1
			
			// Handle boundary cases
			if leftTick >= originalTicks-1 {
				interpolated[band][targetTick] = float64(matrix[band][originalTicks-1])
				continue
			}
			if rightTick >= originalTicks {
				interpolated[band][targetTick] = float64(matrix[band][leftTick])
				continue
			}
			
			// Linear interpolation
			fraction := originalPos - float64(leftTick)
			leftValue := float64(matrix[band][leftTick])
			rightValue := float64(matrix[band][rightTick])
			
			interpolated[band][targetTick] = leftValue + fraction*(rightValue-leftValue)
		}
	}

	return interpolated, dateRange
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
