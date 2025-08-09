package modes

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"gonum.org/v1/plot"
	"labours-go/internal/graphics"
	"labours-go/internal/readers"
)

func OverwritesMatrix(reader readers.Reader, output string) error {
	// Step 1: Extract data from the reader
	people, matrix, err := reader.GetPeopleInteraction()
	if err != nil {
		return fmt.Errorf("failed to get people interaction data: %v", err)
	}

	fmt.Println("Processing overwrites matrix...")

	// Step 2: Process the matrix
	maxPeople := 20 // This can be passed as a parameter or read from configuration
	people, normalizedMatrix := processOverwritesMatrix(people, matrix, maxPeople, true)

	// Step 3: Check if JSON output is required
	if strings.HasSuffix(output, ".json") {
		return saveMatrixAsJSON(output, people, normalizedMatrix)
	}

	// Step 4: Visualize the matrix
	if err := plotOverwritesMatrix(people, normalizedMatrix, output); err != nil {
		return fmt.Errorf("failed to plot overwrites matrix: %v", err)
	}

	fmt.Println("Overwrites matrix generated successfully.")
	return nil
}

func processOverwritesMatrix(people []string, matrix [][]int, maxPeople int, normalize bool) ([]string, [][]float64) {
	// Step 1: Truncate the matrix to the top `maxPeople` developers
	if len(people) > maxPeople {
		order := argsort(matrix)
		matrix = truncateMatrix(matrix, order[:maxPeople])
		people = truncatePeople(people, order[:maxPeople])
		fmt.Printf("Warning: truncated people to most productive %d\n", maxPeople)
	}

	// Step 2: Normalize the matrix to float64
	var normalizedMatrix [][]float64
	if normalize {
		normalizedMatrix = make([][]float64, len(matrix))
		for i := range matrix {
			sum := sumRow(matrix[i])
			normalizedMatrix[i] = make([]float64, len(matrix[i]))
			for j := range matrix[i] {
				if sum != 0 {
					normalizedMatrix[i][j] = float64(matrix[i][j]) / float64(sum)
				}
			}
		}
	} else {
		// Convert to float64 without normalization
		normalizedMatrix = make([][]float64, len(matrix))
		for i := range matrix {
			normalizedMatrix[i] = make([]float64, len(matrix[i]))
			for j := range matrix[i] {
				normalizedMatrix[i][j] = float64(matrix[i][j])
			}
		}
	}

	// Step 3: Invert the matrix (make values negative as in Python)
	for i := range normalizedMatrix {
		for j := range normalizedMatrix[i] {
			normalizedMatrix[i][j] = -normalizedMatrix[i][j]
		}
	}

	// Step 4: Truncate long names
	for i, name := range people {
		if len(name) > 40 {
			people[i] = name[:37] + "..."
		}
	}

	return people, normalizedMatrix
}

func plotOverwritesMatrix(people []string, matrix [][]float64, output string) error {
	// Create and configure the plot
	p := plot.New()
	p.Title.Text = "Overwrites Matrix"
	p.X.Label.Text = "Developers"
	p.Y.Label.Text = "Developers"

	// Ensure the X and Y axis have proper labels
	p.X.Tick.Label.Rotation = -45 // Rotate X-axis labels for readability
	p.NominalX(people...)         // Use `people` as X-axis labels
	p.NominalY(people...)         // Use `people` as Y-axis labels

	// Create the heatmap with your custom palette
	palette := &graphics.CustomPalette{
		Colors: graphics.ColorPalette, // Use your predefined palette
		Min:    -1.0,                  // Adjust Min and Max to align with normalized matrix values
		Max:    0.0,
	}
	heatmap := graphics.NewHeatMap(matrix, people, people, palette)

	// Add the heatmap to the plot
	p.Add(heatmap)

	// Save the plot
	width, height := graphics.GetPlotSize(graphics.ChartTypeSquare)
	if err := p.Save(width, height, output); err != nil {
		return fmt.Errorf("failed to save plot: %v", err)
	}
	return nil
}

func saveMatrixAsJSON(output string, people []string, matrix [][]float64) error {
	data := struct {
		Type   string      `json:"type"`
		People []string    `json:"people"`
		Matrix [][]float64 `json:"matrix"`
	}{
		Type:   "overwrites_matrix",
		People: people,
		Matrix: matrix,
	}

	file, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create JSON output file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func truncateMatrix(matrix [][]int, indices []int) [][]int {
	truncated := make([][]int, len(indices))
	for i, idx := range indices {
		if idx >= len(matrix) {
			continue // Skip invalid indices
		}
		truncated[i] = make([]int, len(indices))
		for j, jdx := range indices {
			if jdx < len(matrix[idx]) {
				truncated[i][j] = matrix[idx][jdx]
			}
		}
	}
	return truncated
}

func truncatePeople(people []string, indices []int) []string {
	truncated := make([]string, len(indices))
	for i, idx := range indices {
		truncated[i] = people[idx]
	}
	return truncated
}

func sumRow(row []int) int {
	sum := 0
	for _, val := range row {
		sum += val
	}
	return sum
}

func argsort(matrix [][]int) []int {
	scores := make([]int, len(matrix))
	for i, row := range matrix {
		scores[i] = row[0]
	}

	indices := make([]int, len(scores))
	for i := range indices {
		indices[i] = i
	}

	sort.Slice(indices, func(i, j int) bool {
		return scores[indices[i]] > scores[indices[j]]
	})

	return indices
}
