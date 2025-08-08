package burndown

import (
	"fmt"
	"time"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// BurndownParameters matches Python's burndown parameters structure
type BurndownParameters struct {
	Sampling    int     // Sampling interval
	Granularity int     // Granularity parameter
	TickSize    float64 // Tick size in seconds
}

// BurndownHeader matches Python's header structure: (start, last, sampling, granularity, tick)
type BurndownHeader struct {
	Start       int64   // Start timestamp
	Last        int64   // End timestamp  
	Sampling    int     // Sampling interval
	Granularity int     // Granularity parameter
	TickSize    float64 // Tick size in seconds
}

// ProcessedBurndown represents the final processed burndown data ready for plotting
type ProcessedBurndown struct {
	Name            string      // Repository/entity name
	Matrix          [][]float64 // Final resampled matrix
	DateRange       []time.Time // Time series for x-axis
	Labels          []string    // Semantic labels for each band/layer
	Granularity     int         // Original granularity
	Sampling        int         // Original sampling
	ResampleMode    string      // Resampling mode used
}

// InterpolateBurndownMatrix converts sparse age-band data into a daily matrix with proper code persistence
// This implements burndown semantics: code persists until explicitly modified/deleted
func InterpolateBurndownMatrix(matrix [][]int, granularity, sampling int, progress bool) ([][]float64, error) {
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return [][]float64{}, fmt.Errorf("empty matrix")
	}

	rows := len(matrix)
	cols := len(matrix[0])

	// Create daily matrix: (matrix.shape[0] * granularity, matrix.shape[1] * sampling)
	dailyRows := rows * granularity
	dailyCols := cols * sampling
	daily := make([][]float64, dailyRows)
	for i := range daily {
		daily[i] = make([]float64, dailyCols)
	}

	fmt.Printf("DEBUG INTERPOLATION: Converting %dx%d sparse matrix to %dx%d daily matrix\n", 
		rows, cols, dailyRows, dailyCols)

	// Restore the original complex Python interpolation algorithm that creates smooth curves
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			// Skip if the future is zeros: y * granularity > (x + 1) * sampling
			if y*granularity > (x+1)*sampling {
				continue
			}

			// Define nested decay function (creates smooth exponential decay curves)
			decay := func(startIndex int, startVal float64) {
				if startVal == 0 {
					return
				}
				k := float64(matrix[y][x]) / startVal // k <= 1, creates decay rate
				scale := float64((x+1)*sampling - startIndex)
				
				for i := y * granularity; i < (y+1)*granularity; i++ {
					var initial float64
					if startIndex > 0 {
						initial = daily[i][startIndex-1]
					}
					// Create smooth exponential-like curves between points
					for j := startIndex; j < (x+1)*sampling; j++ {
						progress := float64(j-startIndex+1) / scale
						daily[i][j] = initial * (1 + (k-1)*progress)
					}
				}
			}

			// Define nested grow function (creates smooth growth curves)  
			grow := func(finishIndex int, finishVal float64) {
				var initial float64
				if x > 0 {
					initial = float64(matrix[y][x-1])
				}
				startIndex := x * sampling
				if startIndex < y*granularity {
					startIndex = y * granularity
				}
				if finishIndex == startIndex {
					return
				}
				// Average slope creates smooth linear growth
				avg := (finishVal - initial) / float64(finishIndex-startIndex)
				
				// Fill triangular region with smooth interpolation
				for j := x * sampling; j < finishIndex; j++ {
					for i := startIndex; i <= j; i++ {
						daily[i][j] = avg
					}
				}
				// Copy values to create smooth persistence
				for j := x * sampling; j < finishIndex; j++ {
					for i := y * granularity; i < x*sampling; i++ {
						if j > 0 {
							daily[i][j] = daily[i][j-1]
						}
					}
				}
			}

			// Main interpolation logic with complex conditional structure for smooth curves
			if (y+1)*granularity >= (x+1)*sampling {
				// Case: Current age band extends beyond current time sampling
				if y*granularity <= x*sampling {
					grow((x+1)*sampling, float64(matrix[y][x]))
				} else if (x+1)*sampling > y*granularity {
					grow((x+1)*sampling, float64(matrix[y][x]))
					// Smooth fill for overlapping region
					avg := float64(matrix[y][x]) / float64((x+1)*sampling-y*granularity)
					for j := y * granularity; j < (x+1)*sampling; j++ {
						for i := y * granularity; i <= j; i++ {
							daily[i][j] = avg
						}
					}
				}
			} else if (y+1)*granularity >= x*sampling {
				// Complex peak calculation case for smooth curves
				var v1, v2 float64
				if x > 0 {
					v1 = float64(matrix[y][x-1])
				}
				v2 = float64(matrix[y][x])
				delta := float64((y+1)*granularity - x*sampling)
				
				var previous float64
				var scale float64
				if x > 0 && (x-1)*sampling >= y*granularity {
					if x > 1 {
						previous = float64(matrix[y][x-2])
					}
					scale = float64(sampling)
				} else {
					if x == 0 {
						scale = float64(sampling)
					} else {
						scale = float64(x*sampling - y*granularity)
					}
				}
				
				// Calculate peak with smooth interpolation
				peak := v1 + (v1-previous)/scale*delta
				if v2 > peak {
					if x < cols-1 {
						k := (v2 - float64(matrix[y][x+1])) / float64(sampling)
						peak = float64(matrix[y][x]) + k*float64((x+1)*sampling-(y+1)*granularity)
					} else {
						peak = v2
					}
				}
				grow((y+1)*granularity, peak)
				decay((y+1)*granularity, peak)
			} else {
				// Case: Age band is completely in the past  
				if x > 0 {
					decay(x*sampling, float64(matrix[y][x-1]))
				}
			}
		}
	}

	return daily, nil
}

// FloorDateTime mimics Python's floor_datetime function
func FloorDateTime(dt time.Time, tickSize float64) time.Time {
	// This function should floor datetime according to tick size
	// For now, we'll implement a basic version
	return dt.Truncate(time.Duration(tickSize) * time.Second)
}

// LoadBurndown is the main function that replicates Python's load_burndown
func LoadBurndown(header BurndownHeader, name string, matrix [][]int, resample string, reportSurvival bool, interpolationProgress bool) (*ProcessedBurndown, error) {
	if header.Sampling <= 0 || header.Granularity <= 0 {
		return nil, fmt.Errorf("invalid sampling (%d) or granularity (%d)", header.Sampling, header.Granularity)
	}

	// DEBUG: Check raw input matrix
	fmt.Printf("DEBUG RAW INPUT MATRIX:\n")
	fmt.Printf("  Dimensions: %dx%d\n", len(matrix), len(matrix[0]))
	for i := 0; i < len(matrix); i++ {
		minVal, maxVal := matrix[i][0], matrix[i][0]
		nonZeroCount := 0
		for j := 0; j < len(matrix[i]); j++ {
			if matrix[i][j] < minVal { minVal = matrix[i][j] }
			if matrix[i][j] > maxVal { maxVal = matrix[i][j] }
			if matrix[i][j] != 0 { nonZeroCount++ }
		}
		fmt.Printf("  Row %d: min=%d, max=%d, non-zeros=%d\n", i, minVal, maxVal, nonZeroCount)
	}

	start := FloorDateTime(time.Unix(header.Start, 0), header.TickSize)
	last := time.Unix(header.Last, 0)
	
	// TODO: Implement survival analysis if reportSurvival is true
	// if reportSurvival {
	//     kmf := fitKaplanMeier(matrix)
	//     if kmf != nil {
	//         printSurvivalFunction(kmf, header.Sampling)
	//     }
	// }

	finish := start.Add(time.Duration(len(matrix[0])*header.Sampling) * time.Duration(header.TickSize) * time.Second)
	
	var finalMatrix [][]float64
	var dateRange []time.Time
	var labels []string
	
	if resample != "no" && resample != "raw" {
		fmt.Printf("resampling to %s, please wait...\n", resample)
		
		// Interpolate the day x day matrix
		daily, err := InterpolateBurndownMatrix(matrix, header.Granularity, header.Sampling, interpolationProgress)
		if err != nil {
			return nil, fmt.Errorf("interpolation failed: %v", err)
		}

		// Zero out data after 'last' timestamp
		lastDays := int(last.Sub(start).Hours() / 24)
		for i := range daily {
			for j := lastDays; j < len(daily[i]); j++ {
				daily[i][j] = 0
			}
		}

		// DEBUG: Analyze interpolated daily matrix before resampling
		fmt.Printf("DEBUG DAILY MATRIX ANALYSIS:\n")
		fmt.Printf("  Daily matrix dimensions: %dx%d\n", len(daily), len(daily[0]))
		for i := 0; i < min(len(daily), 3); i++ { // Show first 3 rows
			nonZeroCount := 0
			negCount := 0
			minVal, maxVal := daily[i][0], daily[i][0]
			for j := 0; j < len(daily[i]); j++ {
				if daily[i][j] != 0 { nonZeroCount++ }
				if daily[i][j] < 0 { negCount++ }
				if daily[i][j] < minVal { minVal = daily[i][j] }
				if daily[i][j] > maxVal { maxVal = daily[i][j] }
			}
			fmt.Printf("  Daily row %d: range=[%.2f to %.2f], non-zeros=%d, negatives=%d\n", 
				i, minVal, maxVal, nonZeroCount, negCount)
		}

		// Resample the bands - convert Python's pandas logic to Go
		dateRange, finalMatrix, labels, err = resampleBurndownData(daily, start, finish, resample)
		if err != nil {
			// Try fallback resampling like Python does
			if resample == "year" || resample == "A" {
				fmt.Println("too loose resampling - by year, trying by month")
				return LoadBurndown(header, name, matrix, "month", false, interpolationProgress)
			} else if resample == "month" || resample == "M" {
				fmt.Println("too loose resampling - by month, trying by day")
				return LoadBurndown(header, name, matrix, "day", false, interpolationProgress)
			}
			return nil, fmt.Errorf("too loose resampling: %s. Try finer", resample)
		}
	} else {
		// Raw mode - show age band labels
		finalMatrix = make([][]float64, len(matrix))
		for i := range matrix {
			finalMatrix[i] = make([]float64, len(matrix[i]))
			for j := range matrix[i] {
				finalMatrix[i][j] = float64(matrix[i][j])
			}
		}
		
		// Generate age band labels like Python does
		labels = make([]string, len(matrix))
		for i := range matrix {
			startTime := start.Add(time.Duration(i*header.Granularity) * time.Duration(header.TickSize) * time.Second)
			endTime := start.Add(time.Duration((i+1)*header.Granularity) * time.Duration(header.TickSize) * time.Second)
			labels[i] = fmt.Sprintf("%s - %s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
		}
		
		// Create date range for raw data
		dateRange = make([]time.Time, len(matrix[0]))
		for i := range dateRange {
			dateRange[i] = start.Add(time.Duration(i*header.Sampling) * time.Duration(header.TickSize) * time.Second)
		}
		
		resample = "M" // fake resampling type as Python does
	}

	// DEBUG: Analyze negative values without clamping to understand root cause
	negativeCount := 0
	minNegative := 0.0
	maxNegative := 0.0
	for i := range finalMatrix {
		for j := range finalMatrix[i] {
			if finalMatrix[i][j] < 0 {
				negativeCount++
				if finalMatrix[i][j] < minNegative {
					minNegative = finalMatrix[i][j]
				}
				if finalMatrix[i][j] > maxNegative || maxNegative == 0.0 {
					maxNegative = finalMatrix[i][j]
				}
			}
		}
	}
	if negativeCount > 0 {
		fmt.Printf("DEBUG NEGATIVE VALUES: Found %d negative values ranging from %.2f to %.2f\n", negativeCount, minNegative, maxNegative)
		fmt.Printf("  This indicates mathematical issues in interpolation that need fixing\n")
	}

	// DEBUG: Print detailed matrix information
	fmt.Printf("DEBUG FINAL RESULT:\n")
	fmt.Printf("  Name: %s\n", name)
	fmt.Printf("  Labels: %v\n", labels)
	fmt.Printf("  Matrix dimensions: %dx%d\n", len(finalMatrix), len(finalMatrix[0]))
	fmt.Printf("  DateRange: %d entries from %s to %s\n", len(dateRange), dateRange[0].Format("2006-01-02"), dateRange[len(dateRange)-1].Format("2006-01-02"))
	for i, label := range labels {
		nonZeroCount := 0
		negCount := 0
		sum := 0.0
		minVal, maxVal := finalMatrix[i][0], finalMatrix[i][0]
		for j := range finalMatrix[i] {
			if finalMatrix[i][j] > 0 {
				nonZeroCount++
				sum += finalMatrix[i][j]
			}
			if finalMatrix[i][j] < 0 {
				negCount++
			}
			if finalMatrix[i][j] < minVal { minVal = finalMatrix[i][j] }
			if finalMatrix[i][j] > maxVal { maxVal = finalMatrix[i][j] }
		}
		fmt.Printf("  Layer %d (%s): positives=%d, negatives=%d, sum=%.2f, range=[%.2f to %.2f]\n", 
			i, label, nonZeroCount, negCount, sum, minVal, maxVal)
	}

	return &ProcessedBurndown{
		Name:         name,
		Matrix:       finalMatrix,
		DateRange:    dateRange,
		Labels:       labels,
		Granularity:  header.Granularity,
		Sampling:     header.Sampling,
		ResampleMode: resample,
	}, nil
}

// resampleBurndownData implements pandas-like resampling logic
func resampleBurndownData(daily [][]float64, start, finish time.Time, resample string) ([]time.Time, [][]float64, []string, error) {
	// Convert resample aliases like Python does
	aliasMap := map[string]string{
		"year":  "A",
		"month": "M", 
		"day":   "D",
	}
	if alias, exists := aliasMap[resample]; exists {
		resample = alias
	}

	// Generate date range based on resampling frequency
	var dateGranularitySampling []time.Time
	switch resample {
	case "A": // Annual
		for year := start.Year(); year <= finish.Year(); year++ {
			yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, start.Location())
			yearEnd := time.Date(year+1, 1, 1, 0, 0, 0, 0, start.Location()).Add(-time.Second) // Dec 31 23:59:59
				
			// Include year if it OVERLAPS with our data timespan (not just if it starts after)
			// Year overlaps if: yearStart <= finish AND yearEnd >= start
			if yearStart.Before(finish) || yearStart.Equal(finish) {
				if yearEnd.After(start) || yearEnd.Equal(start) {
					dateGranularitySampling = append(dateGranularitySampling, yearStart)
				}
			}
		}
	case "M": // Monthly
		current := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
		if current.Before(start) {
			current = current.AddDate(0, 1, 0)
		}
		for current.Before(finish) || current.Equal(finish) {
			dateGranularitySampling = append(dateGranularitySampling, current)
			current = current.AddDate(0, 1, 0)
		}
	case "D": // Daily
		for current := start; current.Before(finish) || current.Equal(finish); current = current.AddDate(0, 0, 1) {
			dateGranularitySampling = append(dateGranularitySampling, current)
		}
	default:
		return nil, nil, nil, fmt.Errorf("unsupported resample mode: %s", resample)
	}

	if len(dateGranularitySampling) == 0 {
		return nil, nil, nil, fmt.Errorf("no valid resampling periods generated")
	}

	if dateGranularitySampling[0].After(finish) {
		return nil, nil, nil, fmt.Errorf("resampling period too loose")
	}

	// Create daily date range for sampling - start from actual data start, not year start
	dateRangeSampling := make([]time.Time, int(finish.Sub(start).Hours()/24)+1)
	for i := range dateRangeSampling {
		dateRangeSampling[i] = start.AddDate(0, 0, i)
	}

	// Fill the new resampled matrix
	resampledMatrix := make([][]float64, len(dateGranularitySampling))
	for i := range resampledMatrix {
		resampledMatrix[i] = make([]float64, len(dateRangeSampling))
	}

	for i, gdt := range dateGranularitySampling {
		var istart, ifinish int
		
		if i > 0 {
			// For overlapping years, use the end of previous year or data start, whichever is later
			prevYearEnd := time.Date(dateGranularitySampling[i-1].Year()+1, 1, 1, 0, 0, 0, 0, start.Location())
			if prevYearEnd.After(start) {
				istart = int(prevYearEnd.Sub(start).Hours() / 24)
			} else {
				istart = 0
			}
		}
		
		// For overlapping years, use the end of current year or data finish, whichever is earlier
		currentYearEnd := time.Date(gdt.Year()+1, 1, 1, 0, 0, 0, 0, start.Location())
		if currentYearEnd.Before(finish) {
			ifinish = int(currentYearEnd.Sub(start).Hours() / 24)
		} else {
			ifinish = int(finish.Sub(start).Hours() / 24)
		}
		
		fmt.Printf("DEBUG RESAMPLING: Layer %d (year %d)\n", i, gdt.Year())
		fmt.Printf("  gdt=%s, start=%s\n", gdt.Format("2006-01-02"), start.Format("2006-01-02"))
		fmt.Printf("  istart=%d, ifinish=%d\n", istart, ifinish)

		var j int
		for idx, sdt := range dateRangeSampling {
			if int(sdt.Sub(start).Hours()/24) >= istart {
				j = idx
				break
			}
		}

		// Sum the daily matrix data for this resampling period
		nonZeroDays := 0
		for k := j; k < len(dateRangeSampling); k++ {
			sdtDays := int(dateRangeSampling[k].Sub(start).Hours() / 24)
			
			// Skip negative days (dates before our start time)
			if sdtDays < 0 {
				continue
			}
			
			var sum float64
			for dailyRow := istart; dailyRow < ifinish && dailyRow < len(daily); dailyRow++ {
				if sdtDays < len(daily[dailyRow]) {
					sum += daily[dailyRow][sdtDays]
				}
			}
			resampledMatrix[i][k] = sum
			if sum > 0 { nonZeroDays++ }
		}
		fmt.Printf("  Processed %d days, %d with non-zero values\n", len(dateRangeSampling)-j, nonZeroDays)
	}

	// Generate labels based on resampling mode (matches Python exactly)
	var labels []string
	switch resample {
	case "A": // Year
		for _, dt := range dateGranularitySampling {
			labels = append(labels, fmt.Sprintf("%d", dt.Year()))
		}
	case "M": // Month
		for _, dt := range dateGranularitySampling {
			labels = append(labels, dt.Format("2006 January"))
		}
	default: // Day or other
		for _, dt := range dateGranularitySampling {
			labels = append(labels, dt.Format("2006-01-02"))
		}
	}

	return dateRangeSampling, resampledMatrix, labels, nil
}