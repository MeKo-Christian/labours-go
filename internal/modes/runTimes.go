package modes

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/spf13/viper"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"labours-go/internal/graphics"
	"labours-go/internal/progress"
	"labours-go/internal/readers"
)

// RunTimes generates runtime analysis and visualization
func RunTimes(reader readers.Reader, output string) error {
	quiet := viper.GetBool("quiet")
	progEstimator := progress.NewProgressEstimator(!quiet)
	
	totalPhases := 3 // data extraction, analysis, plotting
	progEstimator.StartMultiOperation(totalPhases, "Runtime Analysis")

	// Phase 1: Extract runtime data
	progEstimator.NextOperation("Extracting runtime statistics")
	runtimeStats, err := reader.GetRuntimeStats()
	if err != nil {
		progEstimator.FinishMultiOperation()
		return fmt.Errorf("failed to get runtime stats: %v", err)
	}

	if len(runtimeStats) == 0 {
		progEstimator.FinishMultiOperation()
		if !quiet {
			fmt.Println("No runtime data available")
		}
		return nil
	}

	// Phase 2: Analyze runtime patterns
	progEstimator.NextOperation("Analyzing runtime patterns")
	runtimeAnalysis := analyzeRuntimeStats(runtimeStats)

	// Phase 3: Generate visualizations
	progEstimator.NextOperation("Generating visualization")
	if err := plotRuntimeAnalysis(runtimeAnalysis, output); err != nil {
		progEstimator.FinishMultiOperation()
		return fmt.Errorf("failed to generate runtime plots: %v", err)
	}

	progEstimator.FinishMultiOperation()
	if !quiet {
		fmt.Println("Runtime analysis completed successfully.")
	}
	return nil
}

// RuntimeMetric represents a single runtime measurement
type RuntimeMetric struct {
	Operation string
	TimeMs    float64
	Percentage float64
}

// RuntimeAnalysis represents the complete runtime analysis results
type RuntimeAnalysis struct {
	Metrics      []RuntimeMetric
	TotalTime    float64
	Statistics   RuntimeStatistics
}

// RuntimeStatistics provides summary statistics about runtime performance
type RuntimeStatistics struct {
	TotalOperations int
	TotalTimeMs     float64
	AverageTime     float64
	MaxTime         float64
	MinTime         float64
	SlowestOp       string
	FastestOp       string
}

// analyzeRuntimeStats performs analysis on runtime statistics
func analyzeRuntimeStats(runtimeStats map[string]float64) RuntimeAnalysis {
	var metrics []RuntimeMetric
	totalTime := 0.0
	maxTime := 0.0
	minTime := float64(^uint(0) >> 1) // Max float
	slowestOp := ""
	fastestOp := ""
	
	// Calculate total time first
	for _, time := range runtimeStats {
		totalTime += time
	}
	
	// Create metrics with percentages
	for operation, time := range runtimeStats {
		percentage := 0.0
		if totalTime > 0 {
			percentage = (time / totalTime) * 100
		}
		
		metrics = append(metrics, RuntimeMetric{
			Operation:  operation,
			TimeMs:     time,
			Percentage: percentage,
		})
		
		// Track min/max
		if time > maxTime {
			maxTime = time
			slowestOp = operation
		}
		if time < minTime {
			minTime = time
			fastestOp = operation
		}
	}
	
	// Sort by time (descending)
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].TimeMs > metrics[j].TimeMs
	})
	
	// Calculate average
	avgTime := 0.0
	if len(metrics) > 0 {
		avgTime = totalTime / float64(len(metrics))
	}
	
	return RuntimeAnalysis{
		Metrics:   metrics,
		TotalTime: totalTime,
		Statistics: RuntimeStatistics{
			TotalOperations: len(metrics),
			TotalTimeMs:     totalTime,
			AverageTime:     avgTime,
			MaxTime:         maxTime,
			MinTime:         minTime,
			SlowestOp:       slowestOp,
			FastestOp:       fastestOp,
		},
	}
}

// plotRuntimeAnalysis generates runtime visualization plots
func plotRuntimeAnalysis(analysis RuntimeAnalysis, output string) error {
	// Create bar chart of runtime breakdown
	if err := plotRuntimeBreakdown(analysis, output); err != nil {
		return err
	}
	
	// Create pie chart showing percentage breakdown
	if err := plotRuntimePieChart(analysis, output); err != nil {
		return err
	}
	
	return nil
}

// plotRuntimeBreakdown creates a bar chart showing runtime for each operation
func plotRuntimeBreakdown(analysis RuntimeAnalysis, output string) error {
	if len(analysis.Metrics) == 0 {
		return fmt.Errorf("no runtime metrics available")
	}
	
	p := plot.New()
	p.Title.Text = "Runtime Analysis Breakdown"
	p.X.Label.Text = "Operations (by time)"
	p.Y.Label.Text = "Time (milliseconds)"
	
	// Prepare data for bar chart (show top 15 operations)
	maxOps := len(analysis.Metrics)
	if maxOps > 15 {
		maxOps = 15
	}
	
	values := make(plotter.Values, maxOps)
	for i := 0; i < maxOps; i++ {
		values[i] = analysis.Metrics[i].TimeMs
	}
	
	// Create bar chart
	bars, err := plotter.NewBarChart(values, vg.Points(30))
	if err != nil {
		return fmt.Errorf("error creating bar chart: %v", err)
	}
	
	bars.Color = graphics.ColorPalette[5]
	p.Add(bars)
	
	// Create custom tick marks with operation names
	ticks := make([]plot.Tick, maxOps)
	for i := 0; i < maxOps; i++ {
		// Truncate operation names for readability
		opName := analysis.Metrics[i].Operation
		if len(opName) > 12 {
			opName = opName[:12] + "..."
		}
		ticks[i] = plot.Tick{
			Value: float64(i),
			Label: opName,
		}
	}
	p.X.Tick.Marker = plot.ConstantTicks(ticks)
	
	// Save the plot
	outputFile := filepath.Join(output, "runtime_breakdown.png")
	if err := p.Save(16*vg.Inch, 8*vg.Inch, outputFile); err != nil {
		return fmt.Errorf("failed to save runtime breakdown plot: %v", err)
	}
	
	fmt.Printf("Saved runtime breakdown plot to %s\n", outputFile)
	return nil
}

// plotRuntimePieChart creates a pie chart showing percentage breakdown of runtime
func plotRuntimePieChart(analysis RuntimeAnalysis, output string) error {
	if len(analysis.Metrics) == 0 {
		return fmt.Errorf("no runtime metrics available")
	}
	
	// Use a simple stacked bar chart as pie charts are complex in gonum/plot
	p := plot.New()
	p.Title.Text = "Runtime Percentage Distribution"
	p.X.Label.Text = "Cumulative Percentage"
	p.Y.Label.Text = "Operations"
	
	// Prepare data for stacked representation (top 10 operations)
	maxOps := len(analysis.Metrics)
	if maxOps > 10 {
		maxOps = 10
	}
	
	// Create horizontal bars showing percentages
	values := make(plotter.Values, maxOps)
	for i := 0; i < maxOps; i++ {
		values[i] = analysis.Metrics[i].Percentage
	}
	
	// Create horizontal bar chart
	bars, err := plotter.NewBarChart(values, vg.Points(25))
	if err != nil {
		return fmt.Errorf("error creating percentage chart: %v", err)
	}
	
	bars.Color = graphics.ColorPalette[6]
	bars.Horizontal = true
	p.Add(bars)
	
	// Create tick marks with operation names and percentages
	ticks := make([]plot.Tick, maxOps)
	for i := 0; i < maxOps; i++ {
		opName := analysis.Metrics[i].Operation
		if len(opName) > 15 {
			opName = opName[:15] + "..."
		}
		percentage := analysis.Metrics[i].Percentage
		ticks[i] = plot.Tick{
			Value: float64(i),
			Label: fmt.Sprintf("%s (%.1f%%)", opName, percentage),
		}
	}
	p.Y.Tick.Marker = plot.ConstantTicks(ticks)
	
	// Save the plot
	outputFile := filepath.Join(output, "runtime_percentage.png")
	if err := p.Save(16*vg.Inch, 10*vg.Inch, outputFile); err != nil {
		return fmt.Errorf("failed to save runtime percentage plot: %v", err)
	}
	
	fmt.Printf("Saved runtime percentage plot to %s\n", outputFile)
	
	// Print summary information
	fmt.Printf("Runtime Analysis Summary:\n")
	fmt.Printf("  Total operations: %d\n", analysis.Statistics.TotalOperations)
	fmt.Printf("  Total runtime: %.2f ms\n", analysis.Statistics.TotalTimeMs)
	fmt.Printf("  Average runtime per operation: %.2f ms\n", analysis.Statistics.AverageTime)
	fmt.Printf("  Slowest operation: %s (%.2f ms)\n", analysis.Statistics.SlowestOp, analysis.Statistics.MaxTime)
	fmt.Printf("  Fastest operation: %s (%.2f ms)\n", analysis.Statistics.FastestOp, analysis.Statistics.MinTime)
	
	return nil
}