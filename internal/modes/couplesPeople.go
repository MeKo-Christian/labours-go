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

// CouplesPeople generates people coupling analysis and visualization
func CouplesPeople(reader readers.Reader, output string) error {
	quiet := viper.GetBool("quiet")
	progEstimator := progress.NewProgressEstimator(!quiet)
	
	totalPhases := 3 // data extraction, analysis, plotting
	progEstimator.StartMultiOperation(totalPhases, "People Coupling Analysis")

	// Phase 1: Extract people coupling data
	progEstimator.NextOperation("Extracting people coupling data")
	peopleNames, couplingMatrix, err := reader.GetPeopleCooccurrence()
	if err != nil {
		progEstimator.FinishMultiOperation()
		return fmt.Errorf("failed to get people coupling data: %v", err)
	}

	if len(peopleNames) == 0 {
		progEstimator.FinishMultiOperation()
		if !quiet {
			fmt.Println("No people coupling data available")
		}
		return nil
	}

	// Phase 2: Analyze coupling patterns
	progEstimator.NextOperation("Analyzing coupling patterns")
	couplingAnalysis := analyzePeopleCoupling(peopleNames, couplingMatrix)

	// Phase 3: Generate visualizations
	progEstimator.NextOperation("Generating visualization")
	if err := plotPeopleCoupling(couplingAnalysis, output); err != nil {
		progEstimator.FinishMultiOperation()
		return fmt.Errorf("failed to generate people coupling plots: %v", err)
	}

	progEstimator.FinishMultiOperation()
	if !quiet {
		fmt.Println("People coupling analysis completed successfully.")
	}
	return nil
}

// PeopleCouplingPair represents a coupling relationship between two developers
type PeopleCouplingPair struct {
	Person1          string
	Person2          string
	CouplingScore    float64
	CollaborationCount int
}

// PeopleCouplingAnalysis represents the complete coupling analysis results
type PeopleCouplingAnalysis struct {
	PeopleNames    []string
	CouplingMatrix [][]int
	TopCoupling    []PeopleCouplingPair
	Statistics     PeopleCouplingStatistics
}

// PeopleCouplingStatistics provides summary statistics about people coupling
type PeopleCouplingStatistics struct {
	TotalPeople       int
	TotalCollaborations int
	AverageCoupling   float64
	MaxCoupling       int
	MinCoupling       int
}

// analyzePeopleCoupling performs analysis on people coupling data
func analyzePeopleCoupling(peopleNames []string, couplingMatrix [][]int) PeopleCouplingAnalysis {
	analysis := PeopleCouplingAnalysis{
		PeopleNames:    peopleNames,
		CouplingMatrix: couplingMatrix,
	}
	
	// Calculate coupling pairs and statistics
	var pairs []PeopleCouplingPair
	totalCoupling := 0
	maxCoupling := 0
	minCoupling := int(^uint(0) >> 1) // Max int
	
	for i := 0; i < len(peopleNames); i++ {
		for j := i + 1; j < len(peopleNames); j++ {
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
					pairs = append(pairs, PeopleCouplingPair{
						Person1:            peopleNames[i],
						Person2:            peopleNames[j],
						CouplingScore:      float64(coupling),
						CollaborationCount: coupling,
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
	
	// Take top 20 couples for visualization
	maxPeople := viper.GetInt("max-people")
	if maxPeople <= 0 {
		maxPeople = 20
	}
	if len(pairs) > maxPeople {
		analysis.TopCoupling = pairs[:maxPeople]
	} else {
		analysis.TopCoupling = pairs
	}
	
	// Calculate statistics
	avgCoupling := 0.0
	if len(pairs) > 0 {
		avgCoupling = float64(totalCoupling) / float64(len(pairs))
	}
	
	analysis.Statistics = PeopleCouplingStatistics{
		TotalPeople:         len(peopleNames),
		TotalCollaborations: totalCoupling,
		AverageCoupling:     avgCoupling,
		MaxCoupling:         maxCoupling,
		MinCoupling:         minCoupling,
	}
	
	return analysis
}

// plotPeopleCoupling generates coupling visualization plots
func plotPeopleCoupling(analysis PeopleCouplingAnalysis, output string) error {
	// Create heatmap for people collaborations
	if err := plotPeopleCouplingHeatmap(analysis, output); err != nil {
		return err
	}
	
	// Create bar chart of top coupling pairs
	if err := plotTopPeopleCouplingPairs(analysis, output); err != nil {
		return err
	}
	
	return nil
}

// plotPeopleCouplingHeatmap creates a heatmap of people coupling relationships
func plotPeopleCouplingHeatmap(analysis PeopleCouplingAnalysis, output string) error {
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
	
	// Create custom palette for heatmap (blue theme for people collaboration)
	palette := &graphics.CustomPalette{
		Colors: []color.Color{
			color.RGBA{255, 255, 255, 255}, // White for no collaboration
			color.RGBA{200, 220, 255, 255}, // Light blue
			color.RGBA{100, 150, 255, 255}, // Medium blue
			color.RGBA{0, 100, 200, 255},   // Dark blue for strong collaboration
		},
		Min: minVal,
		Max: maxVal,
	}
	
	// Create plot
	p := plot.New()
	p.Title.Text = "People Collaboration Heatmap"
	
	// Create heatmap
	heatmap := graphics.NewHeatMap(heatmapData, analysis.PeopleNames, analysis.PeopleNames, palette)
	p.Add(heatmap)
	
	// Save the plot
	outputFile := filepath.Join(output, "people_coupling_heatmap.png")
	if err := p.Save(12*vg.Inch, 12*vg.Inch, outputFile); err != nil {
		return fmt.Errorf("failed to save heatmap: %v", err)
	}
	
	fmt.Printf("Saved people coupling heatmap to %s\n", outputFile)
	return nil
}

// plotTopPeopleCouplingPairs creates a bar chart of the most coupled people pairs
func plotTopPeopleCouplingPairs(analysis PeopleCouplingAnalysis, output string) error {
	if len(analysis.TopCoupling) == 0 {
		return fmt.Errorf("no coupling pairs data available")
	}
	
	p := plot.New()
	p.Title.Text = "Top Developer Collaboration Pairs"
	p.X.Label.Text = "Collaboration Pair Rank"
	p.Y.Label.Text = "Collaboration Score"
	
	// Prepare data for bar chart
	maxPairs := len(analysis.TopCoupling)
	if maxPairs > 15 {
		maxPairs = 15 // Show top 15 pairs
	}
	
	values := make(plotter.Values, maxPairs)
	for i := 0; i < maxPairs; i++ {
		values[i] = analysis.TopCoupling[i].CouplingScore
	}
	
	// Create bar chart
	bars, err := plotter.NewBarChart(values, vg.Points(30))
	if err != nil {
		return fmt.Errorf("error creating bar chart: %v", err)
	}
	
	bars.Color = graphics.ColorPalette[3]
	p.Add(bars)
	
	// Add x-axis labels with people pair names
	labels := make([]string, maxPairs)
	for i := 0; i < maxPairs; i++ {
		pair := analysis.TopCoupling[i]
		// Truncate long names for readability
		person1 := pair.Person1
		person2 := pair.Person2
		if len(person1) > 10 {
			person1 = person1[:10]
		}
		if len(person2) > 10 {
			person2 = person2[:10]
		}
		labels[i] = person1 + "-" + person2
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
	outputFile := filepath.Join(output, "top_people_coupling_pairs.png")
	if err := p.Save(16*vg.Inch, 8*vg.Inch, outputFile); err != nil {
		return fmt.Errorf("failed to save coupling pairs plot: %v", err)
	}
	
	fmt.Printf("Saved top people coupling pairs plot to %s\n", outputFile)
	
	// Print summary information
	fmt.Printf("People Coupling Analysis Summary:\n")
	fmt.Printf("  Total developers: %d\n", analysis.Statistics.TotalPeople)
	fmt.Printf("  Total collaboration pairs: %d\n", len(analysis.TopCoupling))
	fmt.Printf("  Average collaboration score: %.2f\n", analysis.Statistics.AverageCoupling)
	fmt.Printf("  Max collaboration score: %d\n", analysis.Statistics.MaxCoupling)
	
	return nil
}