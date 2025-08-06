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

// Devs generates plots for individual developers' contributions over time.
func Devs(reader readers.Reader, output string, maxPeople int) error {
	// Step 1: Extract developer statistics
	developerStats, err := reader.GetDeveloperStats()
	if err != nil {
		return fmt.Errorf("failed to get developer stats: %v", err)
	}

	if len(developerStats) > maxPeople {
		fmt.Printf("Picking top %d developers by commit count.\n", maxPeople)
		developerStats = selectTopDevelopers(developerStats, maxPeople)
	}

	// Step 2: Generate time series data for each developer
	devSeries := generateTimeSeries(developerStats)

	// Step 3: Cluster developers by contribution patterns (placeholder logic)
	clusters := clusterDevelopers(devSeries)

	// Step 4: Plot the developer contributions
	if err := plotDevs(developerStats, devSeries, clusters, output); err != nil {
		return fmt.Errorf("failed to generate developer plots: %v", err)
	}

	fmt.Println("Developer plots generated successfully.")
	return nil
}

// selectTopDevelopers selects the top developers by commit count.
func selectTopDevelopers(stats []readers.DeveloperStat, maxPeople int) []readers.DeveloperStat {
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Commits > stats[j].Commits
	})
	if len(stats) > maxPeople {
		return stats[:maxPeople]
	}
	return stats
}

// generateTimeSeries generates synthetic time series data for each developer.
func generateTimeSeries(stats []readers.DeveloperStat) map[string][]float64 {
	devSeries := make(map[string][]float64)
	for _, stat := range stats {
		// Generate a synthetic time series based on commit activity
		// In a real implementation, this would come from daily or weekly data
		series := make([]float64, 52) // 52 weeks in a year
		commitsPerWeek := float64(stat.Commits) / 52.0
		for i := 0; i < len(series); i++ {
			// Add random variation to simulate real activity
			series[i] = commitsPerWeek + float64(i%5)*0.1*commitsPerWeek
		}
		devSeries[stat.Name] = series
	}
	return devSeries
}

// clusterDevelopers clusters developers based on their contribution patterns (placeholder logic).
func clusterDevelopers(devSeries map[string][]float64) map[string]int {
	// Placeholder logic: assign developers to arbitrary clusters
	clusters := make(map[string]int)
	i := 0
	for dev := range devSeries {
		clusters[dev] = i % 3 // Assign developers to 3 clusters
		i++
	}
	return clusters
}

// plotDevs generates plots for developers' contributions.
func plotDevs(developerStats []readers.DeveloperStat, devSeries map[string][]float64, clusters map[string]int, output string) error {
	// Create a new plot
	p := plot.New()
	p.Title.Text = "Developer Contributions Over Time"
	p.X.Label.Text = "Weeks"
	p.Y.Label.Text = "Commits"

	// Plot each developer's time series
	for _, dev := range developerStats {
		series := devSeries[dev.Name]
		pts := make(plotter.XYs, len(series))
		for i, val := range series {
			pts[i].X = float64(i)
			pts[i].Y = val
		}

		line, err := plotter.NewLine(pts)
		if err != nil {
			return fmt.Errorf("error creating plot line for developer %s: %v", dev.Name, err)
		}

		line.Color = graphics.ColorPalette[0] // Use the first color for now
		p.Add(line)
		p.Legend.Add(dev.Name, line)
	}

	// Save the plot
	outputFile := filepath.Join(output, "developer_contributions.png")
	if err := p.Save(16*vg.Inch, 8*vg.Inch, outputFile); err != nil {
		return fmt.Errorf("failed to save plot: %v", err)
	}

	fmt.Printf("Saved developer plot to %s\n", outputFile)
	return nil
}
