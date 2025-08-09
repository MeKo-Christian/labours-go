package modes

import (
	"fmt"
	"path/filepath"
	"sort"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"labours-go/internal/graphics"
	"labours-go/internal/readers"
)

// ShotnessResult represents a processed shotness record with aggregated statistics
type ShotnessResult struct {
	Type            string
	Name            string
	File            string
	TotalHits       int32   // Total number of modifications
	AvgHitsPerTime  float64 // Average modifications per time period
	TimeSpan        int32   // Number of different time periods with modifications
	FirstHit        int32   // First time period with modifications
	LastHit         int32   // Last time period with modifications
}

// Shotness generates code hotspot analysis showing which structural
// units (functions, classes, etc.) have been modified most frequently.
// Provides both text-based statistics (primary) and visualization (optional).
func Shotness(reader readers.Reader, output string) error {
	// Step 1: Read shotness records
	records, err := reader.GetShotnessRecords()
	if err != nil {
		fmt.Printf("Warning: No shotness data available - %v\n", err)
		fmt.Println("To generate shotness analysis, run hercules with the --shotness flag:")
		fmt.Println("  hercules --burndown --shotness <repo> | labours -m shotness")
		return nil
	}

	if len(records) == 0 {
		fmt.Println("No shotness records found in the data.")
		fmt.Println("To generate shotness analysis, run hercules with the --shotness flag:")
		fmt.Println("  hercules --burndown --shotness <repo> | labours -m shotness")
		return nil
	}

	// Step 2: Process and aggregate shotness data
	results := processShotnessRecords(records)

	// Step 3: Print Python-compatible text statistics (primary output)
	printShotnessStats(results)

	// Step 4: Generate visualization (optional - only if output directory specified)
	if output != "" {
		if err := plotShotness(results, output); err != nil {
			fmt.Printf("Warning: Failed to generate shotness plot: %v\n", err)
		}
	}

	return nil
}

// processShotnessRecords processes raw shotness records and calculates aggregate statistics
func processShotnessRecords(records []readers.ShotnessRecord) []ShotnessResult {
	results := make([]ShotnessResult, len(records))
	
	for i, record := range records {
		var totalHits int32
		var firstHit, lastHit int32 = -1, -1
		
		// Calculate statistics from counters
		for timePoint, count := range record.Counters {
			totalHits += count
			
			if firstHit == -1 || timePoint < firstHit {
				firstHit = timePoint
			}
			if lastHit == -1 || timePoint > lastHit {
				lastHit = timePoint
			}
		}
		
		timeSpan := int32(len(record.Counters))
		var avgHits float64
		if timeSpan > 0 {
			avgHits = float64(totalHits) / float64(timeSpan)
		}
		
		results[i] = ShotnessResult{
			Type:            record.Type,
			Name:            record.Name,
			File:            record.File,
			TotalHits:       totalHits,
			AvgHitsPerTime:  avgHits,
			TimeSpan:        timeSpan,
			FirstHit:        firstHit,
			LastHit:         lastHit,
		}
	}
	
	// Sort by total hits (descending) to identify the hottest spots
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalHits > results[j].TotalHits
	})
	
	return results
}

// plotShotness creates a bar chart showing the hottest code spots by modification frequency
func plotShotness(results []ShotnessResult, output string) error {
	// Limit to top 20 hottest spots for better visualization
	maxItems := 20
	if len(results) > maxItems {
		results = results[:maxItems]
	}
	
	// Create a new plot
	p := plot.New()
	p.Title.Text = "Code Hotspots (Most Frequently Modified Structural Units)"
	p.X.Label.Text = "Structural Units"
	p.Y.Label.Text = "Total Modifications"
	
	// Prepare data for the bar chart
	names := make([]string, len(results))
	values := make(plotter.Values, len(results))
	
	for i, result := range results {
		// Create a short label combining type, name, and file
		shortFile := filepath.Base(result.File)
		if len(shortFile) > 15 {
			shortFile = shortFile[:12] + "..."
		}
		
		label := fmt.Sprintf("%s:%s\n(%s)", result.Type, result.Name, shortFile)
		if len(label) > 30 {
			label = label[:27] + "..."
		}
		
		names[i] = label
		values[i] = float64(result.TotalHits)
	}
	
	// Create bar chart
	bars, err := plotter.NewBarChart(values, vg.Points(40))
	if err != nil {
		return fmt.Errorf("failed to create bar chart: %v", err)
	}
	
	// Style the bars with gradient coloring (hottest = red, cooler = blue)
	for i := range bars.Values {
		if len(results) > 1 {
			// Create a heat gradient from red (hottest) to blue (coolest)
			ratio := float64(i) / float64(len(results)-1)
			bars.Color = graphics.HeatColor(1.0 - ratio) // Invert so first (hottest) gets 1.0
		} else {
			bars.Color = graphics.ColorPalette[0]
		}
	}
	
	p.Add(bars)
	
	// Create custom labels for X axis
	p.NominalX(names...)
	
	// Always rotate labels for shotness charts as names tend to be long
	p.X.Tick.Label.Rotation = 0.785398 // 45 degrees in radians
	p.X.Tick.Label.XAlign = -0.5
	p.X.Tick.Label.YAlign = -0.5
	
	// Save the plot with dynamic sizing
	width, height := graphics.GetPlotSize(graphics.ChartTypeWide)
	outputFile := filepath.Join(output, "shotness.png")
	if err := p.Save(width, height, outputFile); err != nil {
		return fmt.Errorf("failed to save shotness plot: %v", err)
	}
	
	// Also create an SVG version
	svgFile := filepath.Join(output, "shotness.svg")
	if err := p.Save(width, height, svgFile); err != nil {
		return fmt.Errorf("failed to save shotness SVG: %v", err)
	}
	
	fmt.Printf("Shotness charts saved to %s and %s\n", outputFile, svgFile)
	
	// Print text summary
	printShotnessSummary(results)
	
	return nil
}

// printShotnessStats prints shotness statistics in Python-compatible format
// Matches the format from Python's show_shotness_stats function:
// "%8d  %s:%s [%s]" % (count, r.file, r.name, r.internal_role)
func printShotnessStats(results []ShotnessResult) {
	fmt.Println("Shotness Analysis - Code Hotspots:")
	
	if len(results) == 0 {
		fmt.Println("No hotspots found.")
		return
	}

	// Print in Python-compatible format: count  file:name [type]
	for _, result := range results {
		fmt.Printf("%8d  %s:%s [%s]\n", 
			result.TotalHits, 
			result.File, 
			result.Name, 
			result.Type)
	}
	
	fmt.Printf("\nTotal: %d hotspots analyzed\n", len(results))
}

// printShotnessSummary prints a detailed text summary of the shotness analysis
func printShotnessSummary(results []ShotnessResult) {
	fmt.Println("\nCode Hotspot Analysis (Shotness):")
	fmt.Println("==================================")
	
	if len(results) == 0 {
		fmt.Println("No hotspots found.")
		return
	}
	
	totalModifications := int32(0)
	typeCount := make(map[string]int)
	
	for _, result := range results {
		totalModifications += result.TotalHits
		typeCount[result.Type]++
	}
	
	fmt.Printf("Total structural units analyzed: %d\n", len(results))
	fmt.Printf("Total modifications tracked: %d\n", totalModifications)
	fmt.Println("\nStructural unit types:")
	for unitType, count := range typeCount {
		fmt.Printf("  %-12s: %d units\n", unitType, count)
	}
	
	fmt.Println("\nTop Hotspots:")
	fmt.Println("Rank | Type       | Name                    | File                     | Hits | Avg/Time | Span")
	fmt.Println("-----|------------|-------------------------|--------------------------|------|----------|-----")
	
	maxDisplay := 15
	if len(results) < maxDisplay {
		maxDisplay = len(results)
	}
	
	for i := 0; i < maxDisplay; i++ {
		result := results[i]
		
		// Truncate long names and file paths for display
		name := result.Name
		if len(name) > 23 {
			name = name[:20] + "..."
		}
		
		file := filepath.Base(result.File)
		if len(file) > 24 {
			file = file[:21] + "..."
		}
		
		fmt.Printf("%4d | %-10s | %-23s | %-24s | %4d | %8.1f | %4d\n",
			i+1,
			result.Type,
			name,
			file,
			result.TotalHits,
			result.AvgHitsPerTime,
			result.TimeSpan,
		)
	}
	
	if len(results) > maxDisplay {
		fmt.Printf("\n... and %d more hotspots\n", len(results)-maxDisplay)
	}
	
	// Summary statistics
	if len(results) > 0 {
		hottest := results[0]
		fmt.Printf("\nHottest spot: %s '%s' in %s (%d modifications)\n",
			hottest.Type, hottest.Name, filepath.Base(hottest.File), hottest.TotalHits)
			
		avgModifications := float64(totalModifications) / float64(len(results))
		fmt.Printf("Average modifications per unit: %.1f\n", avgModifications)
	}
}