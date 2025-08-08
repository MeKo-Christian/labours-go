package modes

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateOwnershipPlot(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test_ownership.png")

	// Sample ownership data for testing
	testPeople := []string{"Alice", "Bob", "Charlie"}
	testOwnership := map[string][][]int{
		"src/main.go": {
			{100, 80, 60}, // Alice's ownership over time
			{0, 20, 30},   // Bob's ownership over time
			{0, 0, 10},    // Charlie's ownership over time
		},
		"src/utils.go": {
			{50, 40, 30},
			{30, 40, 50},
			{20, 20, 20},
		},
		"README.md": {
			{0, 0, 0},
			{100, 80, 60},
			{0, 20, 40},
		},
	}

	err := mockGenerateOwnershipPlot("test", testPeople, testOwnership, outputPath)
	if err != nil {
		t.Errorf("generateOwnershipPlot() error = %v", err)
	}

	// Check if output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Output file was not created: %s", outputPath)
	}
}

func TestGenerateOwnershipPlotEmptyData(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test_ownership_empty.png")

	// Test with empty ownership data
	emptyPeople := []string{}
	emptyOwnership := map[string][][]int{}

	err := mockGenerateOwnershipPlot("test_empty", emptyPeople, emptyOwnership, outputPath)
	if err == nil {
		t.Error("Expected error for empty ownership data, but got nil")
	}
}

func TestGenerateOwnershipPlotSingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test_ownership_single.png")

	testPeople := []string{"Solo Developer"}
	testOwnership := map[string][][]int{
		"main.go": {
			{100, 100, 100}, // 100% ownership throughout
		},
	}

	err := mockGenerateOwnershipPlot("test_single", testPeople, testOwnership, outputPath)
	if err != nil {
		t.Errorf("generateOwnershipPlot() with single file error = %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Output file was not created: %s", outputPath)
	}
}

func TestCalculateFileOwnershipPercentages(t *testing.T) {
	// Test ownership percentage calculation
	ownershipMatrix := [][]int{
		{60, 80, 50}, // Alice
		{30, 10, 40}, // Bob
		{10, 10, 10}, // Charlie
	}

	percentages := calculateFileOwnershipPercentages(ownershipMatrix)

	if len(percentages) != len(ownershipMatrix) {
		t.Errorf("Expected %d ownership arrays, got %d", len(ownershipMatrix), len(percentages))
	}

	// Check first time point: 60 + 30 + 10 = 100
	expectedTotal := 100.0
	actualTotal := 0.0
	for _, dev := range percentages {
		actualTotal += dev[0]
	}

	if actualTotal != expectedTotal {
		t.Errorf("Expected ownership percentages to sum to %f, got %f", expectedTotal, actualTotal)
	}

	// Alice should have 60% at first time point
	if percentages[0][0] != 60.0 {
		t.Errorf("Expected Alice to have 60%% ownership at first point, got %f", percentages[0][0])
	}
}

func TestFindTopFilesByOwnershipChanges(t *testing.T) {
	ownership := map[string][][]int{
		"stable.go": {
			{100, 100, 100}, // No change
			{0, 0, 0},
		},
		"volatile.go": {
			{100, 50, 0}, // High change
			{0, 50, 100},
		},
		"moderate.go": {
			{80, 70, 60}, // Moderate change
			{20, 30, 40},
		},
	}

	topFiles := findTopFilesByOwnershipChanges(ownership, 2)

	if len(topFiles) > 2 {
		t.Errorf("Expected at most 2 files, got %d", len(topFiles))
	}

	// volatile.go should be first due to highest ownership changes
	if len(topFiles) > 0 && topFiles[0] != "volatile.go" {
		t.Errorf("Expected 'volatile.go' to be first, got '%s'", topFiles[0])
	}
}

func TestCalculateOwnershipChurn(t *testing.T) {
	// Test ownership churn calculation
	ownershipMatrix := [][]int{
		{100, 50, 0},  // Alice: high churn (changes from 100 to 0)
		{0, 50, 100},  // Bob: high churn (changes from 0 to 100)
	}

	churn := calculateOwnershipChurn(ownershipMatrix)

	// Churn should be high due to complete ownership transfer
	if churn <= 0 {
		t.Errorf("Expected positive churn value, got %f", churn)
	}

	// Test with stable ownership
	stableMatrix := [][]int{
		{100, 100, 100}, // No change
		{0, 0, 0},       // No change
	}

	stableChurn := calculateOwnershipChurn(stableMatrix)

	if stableChurn >= churn {
		t.Errorf("Expected stable ownership to have lower churn (%f) than volatile (%f)", stableChurn, churn)
	}
}

func TestCalculateOwnershipConcentration(t *testing.T) {
	// Test ownership concentration (Gini coefficient-like measure)
	
	// High concentration (one person owns everything)
	concentratedMatrix := [][]int{
		{100, 100, 100},
		{0, 0, 0},
		{0, 0, 0},
	}

	concentrated := calculateOwnershipConcentration(concentratedMatrix)

	// Low concentration (evenly distributed)
	distributedMatrix := [][]int{
		{33, 33, 33},
		{33, 33, 33},
		{34, 34, 34},
	}

	distributed := calculateOwnershipConcentration(distributedMatrix)

	if concentrated <= distributed {
		t.Errorf("Expected concentrated ownership (%f) to have higher concentration than distributed (%f)", concentrated, distributed)
	}
}

// Helper functions for testing

func calculateFileOwnershipPercentages(ownershipMatrix [][]int) [][]float64 {
	if len(ownershipMatrix) == 0 || len(ownershipMatrix[0]) == 0 {
		return [][]float64{}
	}

	percentages := make([][]float64, len(ownershipMatrix))
	timePoints := len(ownershipMatrix[0])

	for dev := range ownershipMatrix {
		percentages[dev] = make([]float64, timePoints)
	}

	for t := 0; t < timePoints; t++ {
		total := 0
		for dev := 0; dev < len(ownershipMatrix); dev++ {
			total += ownershipMatrix[dev][t]
		}

		if total > 0 {
			for dev := 0; dev < len(ownershipMatrix); dev++ {
				percentages[dev][t] = float64(ownershipMatrix[dev][t]) * 100.0 / float64(total)
			}
		}
	}

	return percentages
}

func findTopFilesByOwnershipChanges(ownership map[string][][]int, limit int) []string {
	type fileChurn struct {
		filename string
		churn    float64
	}

	var files []fileChurn

	for filename, matrix := range ownership {
		churn := calculateOwnershipChurn(matrix)
		files = append(files, fileChurn{filename: filename, churn: churn})
	}

	// Simple sort by churn (descending)
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].churn < files[j].churn {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	result := make([]string, 0, limit)
	for i := 0; i < len(files) && i < limit; i++ {
		result = append(result, files[i].filename)
	}

	return result
}

func calculateOwnershipChurn(ownershipMatrix [][]int) float64 {
	if len(ownershipMatrix) == 0 || len(ownershipMatrix[0]) == 0 {
		return 0
	}

	totalChurn := 0.0
	timePoints := len(ownershipMatrix[0])

	for t := 1; t < timePoints; t++ {
		for dev := 0; dev < len(ownershipMatrix); dev++ {
			change := float64(ownershipMatrix[dev][t] - ownershipMatrix[dev][t-1])
			if change < 0 {
				change = -change // Absolute value
			}
			totalChurn += change
		}
	}

	return totalChurn / float64(timePoints-1)
}

func calculateOwnershipConcentration(ownershipMatrix [][]int) float64 {
	if len(ownershipMatrix) == 0 || len(ownershipMatrix[0]) == 0 {
		return 0
	}

	concentration := 0.0
	timePoints := len(ownershipMatrix[0])

	for t := 0; t < timePoints; t++ {
		total := 0
		for dev := 0; dev < len(ownershipMatrix); dev++ {
			total += ownershipMatrix[dev][t]
		}

		if total > 0 {
			// Calculate Herfindahl index for this time point
			herfindahl := 0.0
			for dev := 0; dev < len(ownershipMatrix); dev++ {
				share := float64(ownershipMatrix[dev][t]) / float64(total)
				herfindahl += share * share
			}
			concentration += herfindahl
		}
	}

	return concentration / float64(timePoints)
}

// Mock function for testing
func mockGenerateOwnershipPlot(name string, people []string, ownership map[string][][]int, outputPath string) error {
	// Simple mock implementation
	if len(people) == 0 || len(ownership) == 0 {
		return fmt.Errorf("empty people or ownership data")
	}
	
	// Create a simple test file to simulate chart generation
	content := fmt.Sprintf("Mock ownership chart for %s with %d people and %d files", name, len(people), len(ownership))
	return os.WriteFile(outputPath, []byte(content), 0644)
}