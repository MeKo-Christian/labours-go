package modes

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/viper"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"labours-go/internal/graphics"
	"labours-go/internal/progress"
	"labours-go/internal/readers"
)

func OwnershipBurndown(reader readers.Reader, output string) error {
	// Initialize progress tracking
	quiet := viper.GetBool("quiet")
	progEstimator := progress.NewProgressEstimator(!quiet)
	
	// Start multi-phase operation for ownership analysis
	totalPhases := 4 // validation, data extraction, processing, visualization
	progEstimator.StartMultiOperation(totalPhases, "Ownership Burndown Analysis")

	// Phase 1: Validate output path
	progEstimator.NextOperation("Validating output path")
	if output == "" {
		output = "ownership.png"
		if !quiet {
			fmt.Printf("Output not provided, using default: %s\n", output)
		}
	}

	outputDir := filepath.Dir(output)
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		progEstimator.FinishMultiOperation()
		return fmt.Errorf("failed to create output directory %s: %v", outputDir, err)
	}

	// Phase 2: Extract data from the reader
	progEstimator.NextOperation("Extracting ownership data")
	peopleSequence, ownershipData, err := reader.GetOwnershipBurndown()
	if err != nil {
		progEstimator.FinishMultiOperation()
		return fmt.Errorf("failed to get ownership burndown data: %v", err)
	}

	// Metadata for the timeline (hardcoded sampling for simplicity)
	sampling := 1                // Assume daily sampling
	startTime := time.Unix(0, 0) // Placeholder for start time; replace with actual value if needed
	lastTime := startTime.Add(time.Duration(len(ownershipData[peopleSequence[0]][0])*sampling) * 24 * time.Hour)

	// Phase 3: Process the data
	progEstimator.NextOperation("Processing ownership data")
	maxPeople := 20      // Maximum number of people to display
	orderByTime := false // Sort developers by their first appearance
	names, peopleMatrix, dateRange := processOwnershipBurndownWithProgress(
		startTime, lastTime, sampling, peopleSequence, ownershipData, maxPeople, orderByTime, progEstimator)

	// Phase 4: Generate output
	progEstimator.NextOperation("Generating visualization")
	
	// Check if JSON output is required
	if filepath.Ext(output) == ".json" {
		progEstimator.FinishMultiOperation()
		return saveOwnershipBurndownAsJSON(output, names, peopleMatrix, dateRange, lastTime)
	}

	// Visualize the data
	if err := plotOwnershipBurndown(names, peopleMatrix, dateRange, lastTime, output); err != nil {
		progEstimator.FinishMultiOperation()
		return fmt.Errorf("failed to plot ownership burndown: %v", err)
	}

	progEstimator.FinishMultiOperation()
	if !quiet {
		fmt.Println("Ownership burndown chart generated successfully.")
	}
	return nil
}

func processOwnershipBurndown(
	start, last time.Time, sampling int,
	sequence []string, data map[string][][]int,
	maxPeople int, orderByTime bool,
) ([]string, [][]float64, []time.Time) {
	// Aggregate the ownership data
	people := make([][]float64, len(sequence))
	for i, name := range sequence {
		rows := data[name]
		total := make([]float64, len(rows[0]))
		for _, row := range rows {
			for j, val := range row {
				total[j] += float64(val)
			}
		}
		people[i] = total
	}

	// Create a date range based on sampling
	dateRange := make([]time.Time, len(people[0]))
	for i := 0; i < len(dateRange); i++ {
		dateRange[i] = start.Add(time.Duration(i*sampling) * time.Hour * 24)
	}

	// Truncate to maxPeople
	if len(people) > maxPeople {
		sums := make([]float64, len(people))
		for i, row := range people {
			for _, val := range row {
				sums[i] += val
			}
		}

		indices := argsortDescending(sums)
		chosen := indices[:maxPeople]
		others := indices[maxPeople:]

		// Aggregate "others"
		othersTotal := make([]float64, len(people[0]))
		for _, idx := range others {
			for j, val := range people[idx] {
				othersTotal[j] += val
			}
		}

		// Update people and sequence
		truncatedPeople := make([][]float64, maxPeople+1)
		truncatedNames := make([]string, maxPeople+1)
		for i, idx := range chosen {
			truncatedPeople[i] = people[idx]
			truncatedNames[i] = sequence[idx]
		}
		truncatedPeople[maxPeople] = othersTotal
		truncatedNames[maxPeople] = "others"

		people = truncatedPeople
		sequence = truncatedNames
	}

	// Sort by first appearance or total ownership
	if orderByTime {
		appearances := make([]int, len(people))
		for i, row := range people {
			appearances[i] = findFirstNonZero(row)
		}
		indices := argsortAscending(appearances)
		people = reorder(people, indices)
		sequence = reorderStrings(sequence, indices)
	} else {
		totalOwnership := make([]float64, len(people))
		for i, row := range people {
			for _, val := range row {
				totalOwnership[i] += val
			}
		}
		indices := argsortDescending(totalOwnership)
		people = reorder(people, indices)
		sequence = reorderStrings(sequence, indices)
	}

	return sequence, people, dateRange
}

// processOwnershipBurndownWithProgress processes ownership data with progress tracking
func processOwnershipBurndownWithProgress(
	start, last time.Time, sampling int,
	sequence []string, data map[string][][]int,
	maxPeople int, orderByTime bool,
	progEstimator *progress.ProgressEstimator,
) ([]string, [][]float64, []time.Time) {
	// Start detailed progress for data processing
	totalSteps := len(sequence) + 2 // aggregation steps + sorting + date range creation
	progEstimator.StartOperation("Aggregating ownership data", totalSteps)
	
	// Aggregate the ownership data
	people := make([][]float64, len(sequence))
	for i, name := range sequence {
		progEstimator.UpdateProgress(1)
		rows := data[name]
		total := make([]float64, len(rows[0]))
		for _, row := range rows {
			for j, val := range row {
				total[j] += float64(val)
			}
		}
		people[i] = total
	}

	// Create a date range based on sampling
	progEstimator.UpdateProgress(1)
	dateRange := make([]time.Time, len(people[0]))
	for i := 0; i < len(dateRange); i++ {
		dateRange[i] = start.Add(time.Duration(i*sampling) * time.Hour * 24)
	}

	// Truncate to maxPeople
	if len(people) > maxPeople {
		sums := make([]float64, len(people))
		for i, row := range people {
			for _, val := range row {
				sums[i] += val
			}
		}

		indices := argsortDescending(sums)
		chosen := indices[:maxPeople]
		others := indices[maxPeople:]

		// Aggregate "others"
		othersTotal := make([]float64, len(people[0]))
		for _, idx := range others {
			for j, val := range people[idx] {
				othersTotal[j] += val
			}
		}

		// Update people and sequence
		truncatedPeople := make([][]float64, maxPeople+1)
		truncatedNames := make([]string, maxPeople+1)
		for i, idx := range chosen {
			truncatedPeople[i] = people[idx]
			truncatedNames[i] = sequence[idx]
		}
		truncatedPeople[maxPeople] = othersTotal
		truncatedNames[maxPeople] = "others"

		people = truncatedPeople
		sequence = truncatedNames
	}

	// Sort by first appearance or total ownership
	progEstimator.UpdateProgress(1)
	if orderByTime {
		appearances := make([]int, len(people))
		for i, row := range people {
			appearances[i] = findFirstNonZero(row)
		}
		indices := argsortAscending(appearances)
		people = reorder(people, indices)
		sequence = reorderStrings(sequence, indices)
	} else {
		totalOwnership := make([]float64, len(people))
		for i, row := range people {
			for _, val := range row {
				totalOwnership[i] += val
			}
		}
		indices := argsortDescending(totalOwnership)
		people = reorder(people, indices)
		sequence = reorderStrings(sequence, indices)
	}

	progEstimator.FinishOperation()
	return sequence, people, dateRange
}

func plotOwnershipBurndown(names []string, people [][]float64, dateRange []time.Time, lastTime time.Time, output string) error {
	// Create a plot
	p := plot.New()
	p.Title.Text = "Ownership Burndown"
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Ownership"

	// Convert people data into plotter.XYs
	stackData := make([]plotter.XYs, len(people))
	for i, row := range people {
		points := make(plotter.XYs, len(row))
		for j, val := range row {
			points[j].X = float64(dateRange[j].Unix())
			points[j].Y = val
		}
		stackData[i] = points
	}

	// Add stackplot layers
	for i, points := range stackData {
		line, err := plotter.NewLine(points)
		if err != nil {
			return fmt.Errorf("failed to create line plot: %v", err)
		}
		line.Color = graphics.ColorPalette[i%len(graphics.ColorPalette)]
		p.Add(line)
		p.Legend.Add(names[i], line)
	}

	// Save the plot
	width, height := graphics.GetPlotSize(graphics.ChartTypeCompact)
	if err := p.Save(width, height, output); err != nil {
		return fmt.Errorf("failed to save plot: %v", err)
	}

	return nil
}

func saveOwnershipBurndownAsJSON(output string, names []string, people [][]float64, dateRange []time.Time, lastTime time.Time) error {
	data := struct {
		Type      string      `json:"type"`
		Names     []string    `json:"names"`
		People    [][]float64 `json:"people"`
		DateRange []time.Time `json:"date_range"`
		Last      time.Time   `json:"last"`
	}{
		Type:      "ownership",
		Names:     names,
		People:    people,
		DateRange: dateRange,
		Last:      lastTime,
	}

	file, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create JSON output file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to write JSON data: %v", err)
	}

	fmt.Printf("JSON data saved to %s\n", output)
	return nil
}

func argsortDescending(data []float64) []int {
	indices := make([]int, len(data))
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(i, j int) bool {
		return data[indices[i]] > data[indices[j]]
	})
	return indices
}

func argsortAscending(data []int) []int {
	indices := make([]int, len(data))
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(i, j int) bool {
		return data[indices[i]] < data[indices[j]]
	})
	return indices
}

func findFirstNonZero(row []float64) int {
	for i, val := range row {
		if val > 0 {
			return i
		}
	}
	return math.MaxInt
}

func reorder(data [][]float64, indices []int) [][]float64 {
	reordered := make([][]float64, len(indices))
	for i, idx := range indices {
		reordered[i] = data[idx]
	}
	return reordered
}

func reorderStrings(data []string, indices []int) []string {
	reordered := make([]string, len(indices))
	for i, idx := range indices {
		reordered[i] = data[idx]
	}
	return reordered
}
