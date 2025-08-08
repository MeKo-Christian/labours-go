package modes

import (
	"fmt"
	"image/color"
	"path/filepath"
	"strconv"

	"github.com/spf13/viper"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"labours-go/internal/graphics"
	"labours-go/internal/progress"
	"labours-go/internal/readers"
)

// CouplesShotness generates shotness-based coupling analysis and visualization
func CouplesShotness(reader readers.Reader, output string) error {
	quiet := viper.GetBool("quiet")
	progEstimator := progress.NewProgressEstimator(!quiet)
	
	totalPhases := 3 // data extraction, analysis, plotting
	progEstimator.StartMultiOperation(totalPhases, "Shotness Coupling Analysis")

	// Phase 1: Extract shotness coupling data
	progEstimator.NextOperation("Extracting shotness coupling data")
	entityNames, couplingMatrix, err := reader.GetShotnessCooccurrence()
	if err != nil {
		progEstimator.FinishMultiOperation()
		return fmt.Errorf("failed to get shotness coupling data: %v", err)
	}

	if len(entityNames) == 0 {
		progEstimator.FinishMultiOperation()
		if !quiet {
			fmt.Println("No shotness coupling data available")
		}
		return nil
	}

	// Phase 2: Analyze coupling patterns
	progEstimator.NextOperation("Analyzing shotness coupling patterns")
	couplingAnalysis := analyzeShotnessCoupling(entityNames, couplingMatrix)

	// Phase 3: Generate visualizations
	progEstimator.NextOperation("Generating visualization")
	if err := plotShotnessCoupling(couplingAnalysis, output); err != nil {
		progEstimator.FinishMultiOperation()
		return fmt.Errorf("failed to generate shotness coupling plots: %v", err)
	}

	progEstimator.FinishMultiOperation()
	if !quiet {
		fmt.Println("Shotness coupling analysis completed successfully.")
	}
	return nil
}

// ShotnessCouplingPair represents a coupling relationship between two shotness entities
type ShotnessCouplingPair struct {
	Entity1       string
	Entity2       string
	CouplingScore float64
	CooccuranceCount int
}

// ShotnessCouplingAnalysis represents the complete shotness coupling analysis results
type ShotnessCouplingAnalysis struct {
	EntityNames    []string
	CouplingMatrix [][]int
	TopCoupling    []ShotnessCouplingPair
	Statistics     ShotnessCouplingStatistics
}

// ShotnessCouplingStatistics provides summary statistics about shotness coupling
type ShotnessCouplingStatistics struct {
	TotalEntities     int
	TotalCouplings    int
	AverageCoupling   float64
	MaxCoupling       int
	MinCoupling       int
}

// analyzeShotnessCoupling performs analysis on shotness coupling data
func analyzeShotnessCoupling(entityNames []string, couplingMatrix [][]int) ShotnessCouplingAnalysis {
	analysis := ShotnessCouplingAnalysis{
		EntityNames:    entityNames,
		CouplingMatrix: couplingMatrix,
	}
	
	// Calculate coupling pairs and statistics
	var pairs []ShotnessCouplingPair
	totalCoupling := 0
	maxCoupling := 0
	minCoupling := int(^uint(0) >> 1) // Max int
	
	for i := 0; i < len(entityNames); i++ {
		for j := i + 1; j < len(entityNames); j++ {
			if i < len(couplingMatrix) && j < len(couplingMatrix[i]) {
				coupling := couplingMatrix[i][j]
				totalCoupling += coupling
				
				if coupling > maxCoupling {
					maxCoupling = coupling
				}
				if coupling < minCoupling && coupling > 0 {
					minCoupling = coupling
				}
				
				if coupling > 0 {
					pairs = append(pairs, ShotnessCouplingPair{
						Entity1:          entityNames[i],
						Entity2:          entityNames[j],
						CouplingScore:    float64(coupling),
						CooccuranceCount: coupling,
					})
				}
			}
		}
	}
	
	// Sort pairs by coupling score (descending)
	for i := 0; i < len(pairs)-1; i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[i].CouplingScore < pairs[j].CouplingScore {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
	
	// Take top 25 couples for visualization (shotness can be more detailed)
	if len(pairs) > 25 {
		analysis.TopCoupling = pairs[:25]
	} else {
		analysis.TopCoupling = pairs
	}
	
	// Calculate statistics
	avgCoupling := 0.0
	if len(pairs) > 0 {
		avgCoupling = float64(totalCoupling) / float64(len(pairs))
	}
	
	analysis.Statistics = ShotnessCouplingStatistics{
		TotalEntities:   len(entityNames),
		TotalCouplings:  totalCoupling,
		AverageCoupling: avgCoupling,
		MaxCoupling:     maxCoupling,
		MinCoupling:     minCoupling,
	}
	
	return analysis
}

// plotShotnessCoupling generates coupling visualization plots
func plotShotnessCoupling(analysis ShotnessCouplingAnalysis, output string) error {
	// Create heatmap for shotness entities
	if err := plotShotnessCouplingHeatmap(analysis, output); err != nil {
		return err
	}
	
	// Create bar chart of top coupling pairs
	if err := plotTopShotnessCouplingPairs(analysis, output); err != nil {
		return err
	}
	
	return nil
}

// plotShotnessCouplingHeatmap creates a heatmap of shotness coupling relationships
func plotShotnessCouplingHeatmap(analysis ShotnessCouplingAnalysis, output string) error {
	if len(analysis.CouplingMatrix) == 0 {
		return fmt.Errorf("no coupling matrix data available")
	}
	
	// Create heatmap data
	heatmapData := make([][]float64, len(analysis.CouplingMatrix))
	maxVal := 0.0
	minVal := float64(analysis.Statistics.MaxCoupling)
	
	for i, row := range analysis.CouplingMatrix {
		heatmapData[i] = make([]float64, len(row))
		for j, val := range row {
			heatmapData[i][j] = float64(val)
			if float64(val) > maxVal {
				maxVal = float64(val)
			}
			if float64(val) < minVal && val > 0 {
				minVal = float64(val)
			}
		}
	}
	
	// Create custom palette for heatmap (green theme for shotness)
	palette := &graphics.CustomPalette{
		Colors: []color.Color{
			color.RGBA{255, 255, 255, 255}, // White for no coupling
			color.RGBA{200, 255, 200, 255}, // Light green
			color.RGBA{100, 255, 100, 255}, // Medium green
			color.RGBA{0, 200, 0, 255},     // Dark green for high coupling
		},
		Min: minVal,
		Max: maxVal,
	}
	
	// Create plot
	p := plot.New()
	p.Title.Text = "Shotness Coupling Heatmap"
	
	// Create heatmap
	heatmap := graphics.NewHeatMap(heatmapData, analysis.EntityNames, analysis.EntityNames, palette)
	p.Add(heatmap)
	
	// Save the plot
	outputFile := filepath.Join(output, "shotness_coupling_heatmap.png")
	if err := p.Save(12*vg.Inch, 12*vg.Inch, outputFile); err != nil {
		return fmt.Errorf("failed to save heatmap: %v", err)
	}
	
	fmt.Printf("Saved shotness coupling heatmap to %s\n", outputFile)
	return nil
}

// plotTopShotnessCouplingPairs creates a bar chart of the most coupled shotness entities
func plotTopShotnessCouplingPairs(analysis ShotnessCouplingAnalysis, output string) error {
	if len(analysis.TopCoupling) == 0 {
		return fmt.Errorf("no coupling pairs data available")
	}
	
	p := plot.New()
	p.Title.Text = "Top Shotness Coupling Pairs"
	p.X.Label.Text = "Coupling Pair Rank"
	p.Y.Label.Text = "Shotness Coupling Score"
	
	// Prepare data for bar chart
	maxPairs := len(analysis.TopCoupling)
	if maxPairs > 20 {
		maxPairs = 20 // Show top 20 pairs
	}
	
	values := make(plotter.Values, maxPairs)
	for i := 0; i < maxPairs; i++ {
		values[i] = analysis.TopCoupling[i].CouplingScore
	}
	
	// Create bar chart
	bars, err := plotter.NewBarChart(values, vg.Points(25))
	if err != nil {
		return fmt.Errorf("error creating bar chart: %v", err)
	}
	
	bars.Color = graphics.ColorPalette[4]
	p.Add(bars)
	
	// Add x-axis labels with entity pair names (truncated)
	labels := make([]string, maxPairs)
	for i := 0; i < maxPairs; i++ {
		pair := analysis.TopCoupling[i]
		// Truncate entity names for readability
		entity1 := pair.Entity1
		entity2 := pair.Entity2
		if len(entity1) > 8 {
			entity1 = entity1[:8] + "..."
		}
		if len(entity2) > 8 {
			entity2 = entity2[:8] + "..."
		}
		labels[i] = entity1 + "-" + entity2
	}
	
	// Create custom tick marks
	ticks := make([]plot.Tick, maxPairs)
	for i := range ticks {
		ticks[i] = plot.Tick{
			Value: float64(i),
			Label: strconv.Itoa(i + 1), // Just show rank numbers
		}
	}
	p.X.Tick.Marker = plot.ConstantTicks(ticks)
	
	// Save the plot
	outputFile := filepath.Join(output, "top_shotness_coupling_pairs.png")
	if err := p.Save(16*vg.Inch, 8*vg.Inch, outputFile); err != nil {
		return fmt.Errorf("failed to save coupling pairs plot: %v", err)
	}
	
	fmt.Printf("Saved top shotness coupling pairs plot to %s\n", outputFile)
	
	// Print summary information
	fmt.Printf("Shotness Coupling Analysis Summary:\n")
	fmt.Printf("  Total entities: %d\n", analysis.Statistics.TotalEntities)
	fmt.Printf("  Total coupling relationships: %d\n", len(analysis.TopCoupling))
	fmt.Printf("  Average coupling score: %.2f\n", analysis.Statistics.AverageCoupling)
	fmt.Printf("  Max coupling score: %d\n", analysis.Statistics.MaxCoupling)
	
	return nil
}