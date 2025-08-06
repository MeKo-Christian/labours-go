package main

import (
	"fmt"
	"labours-go/internal/pb"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/proto"
)

// This utility creates sample test data files for testing
func main() {
	testDir := "test/testdata"
	
	// Ensure testdata directory exists
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		fmt.Printf("Failed to create testdata directory: %v\n", err)
		os.Exit(1)
	}

	// Create simple burndown data
	fmt.Println("Creating simple burndown test data...")
	simpleBurndown := generateSimpleBurndownData()
	simpleData, err := proto.Marshal(simpleBurndown)
	if err != nil {
		fmt.Printf("Failed to serialize simple data: %v\n", err)
		os.Exit(1)
	}

	simplePath := filepath.Join(testDir, "simple_burndown.pb")
	err = os.WriteFile(simplePath, simpleData, 0644)
	if err != nil {
		fmt.Printf("Failed to write simple data: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created %s (%d bytes)\n", simplePath, len(simpleData))

	// Create realistic burndown data
	fmt.Println("Creating realistic burndown test data...")
	realisticBurndown := generateRealisticBurndownData()
	realisticData, err := proto.Marshal(realisticBurndown)
	if err != nil {
		fmt.Printf("Failed to serialize realistic data: %v\n", err)
		os.Exit(1)
	}

	realisticPath := filepath.Join(testDir, "realistic_burndown.pb")
	err = os.WriteFile(realisticPath, realisticData, 0644)
	if err != nil {
		fmt.Printf("Failed to write realistic data: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created %s (%d bytes)\n", realisticPath, len(realisticData))

	// Create a README for the test data
	readmePath := filepath.Join(testDir, "README.md")
	readme := `# Test Data Files

This directory contains sample hercules protobuf files for testing labours-go.

## Files

- **simple_burndown.pb**: Small-scale test data with sample project data
- **realistic_burndown.pb**: Large-scale test data with more comprehensive metrics
  
## Generated Data Characteristics

### Simple Burndown Data
- Basic project burndown matrix
- Simple file and people data
- Metadata with timestamps

### Realistic Burndown Data  
- More comprehensive burndown analysis
- Multiple developers and files
- Extended time range

## Usage in Tests

These files are used by:
- Unit tests for reader functionality
- Integration tests for end-to-end workflows
- Visual regression tests for chart consistency
- Performance benchmarks

## Regeneration

To regenerate this test data, run:
` + "```bash" + `
go run test/create_sample_data.go
` + "```" + `
`

	err = os.WriteFile(readmePath, []byte(readme), 0644)
	if err != nil {
		fmt.Printf("Failed to write README: %v\n", err)
	} else {
		fmt.Printf("Created %s\n", readmePath)
	}

	fmt.Println("Sample test data creation completed successfully!")
}

func generateSimpleBurndownData() *pb.AnalysisResults {
	return &pb.AnalysisResults{
		Burndown: &pb.BurndownAnalysisResults{
			Project: &pb.CompressedSparseRowMatrix{
				NumberOfRows:    2,
				NumberOfColumns: 10,
				Data:            generateSequentialData(20),
				Indices:         generateSequentialIndices(20),
				Indptr:          []int64{0, 10, 20},
			},
			Files: &pb.CompressedSparseRowMatrix{
				NumberOfRows:    3,
				NumberOfColumns: 10,
				Data:            generateSequentialData(30),
				Indices:         generateSequentialIndices(30),
				Indptr:          []int64{0, 10, 20, 30},
			},
			People: &pb.CompressedSparseRowMatrix{
				NumberOfRows:    3,
				NumberOfColumns: 10,
				Data:            generateSequentialData(30),
				Indices:         generateSequentialIndices(30),
				Indptr:          []int64{0, 10, 20, 30},
			},
			FilesOwnership: &pb.FilesOwnership{
				Value: map[string]int32{
					"main.go":    0,
					"utils.go":   1,
					"handler.go": 2,
				},
			},
			TickSize: 86400, // 1 day in seconds
		},
		Metadata: &pb.Metadata{
			Repository:    "test-repository",
			BeginUnixTime: 1640995200, // 2022-01-01
			EndUnixTime:   1672531200, // 2023-01-01
			Version:       1,
		},
		FileNames:   []string{"main.go", "utils.go", "handler.go"},
		PeopleNames: []string{"Alice", "Bob", "Charlie"},
	}
}

func generateRealisticBurndownData() *pb.AnalysisResults {
	return &pb.AnalysisResults{
		Burndown: &pb.BurndownAnalysisResults{
			Project: &pb.CompressedSparseRowMatrix{
				NumberOfRows:    5,
				NumberOfColumns: 50,
				Data:            generateSequentialData(250),
				Indices:         generateSequentialIndices(250),
				Indptr:          []int64{0, 50, 100, 150, 200, 250},
			},
			Files: &pb.CompressedSparseRowMatrix{
				NumberOfRows:    10,
				NumberOfColumns: 50,
				Data:            generateSequentialData(500),
				Indices:         generateSequentialIndices(500),
				Indptr:          generateIndptr(10, 50),
			},
			People: &pb.CompressedSparseRowMatrix{
				NumberOfRows:    5,
				NumberOfColumns: 50,
				Data:            generateSequentialData(250),
				Indices:         generateSequentialIndices(250),
				Indptr:          []int64{0, 50, 100, 150, 200, 250},
			},
			FilesOwnership: &pb.FilesOwnership{
				Value: generateFileOwnership(10),
			},
			TickSize: 86400,
		},
		Metadata: &pb.Metadata{
			Repository:    "realistic-repository",
			BeginUnixTime: 1640995200,
			EndUnixTime:   1672531200,
			Version:       1,
		},
		FileNames:   generateFileNames(10),
		PeopleNames: []string{"Alice", "Bob", "Charlie", "Diana", "Eve"},
	}
}

func generateSequentialData(count int) []int64 {
	data := make([]int64, count)
	for i := 0; i < count; i++ {
		data[i] = int64(1000 - i*5) // Decreasing pattern
		if data[i] < 0 {
			data[i] = 0
		}
	}
	return data
}

func generateSequentialIndices(count int) []int32 {
	indices := make([]int32, count)
	for i := 0; i < count; i++ {
		indices[i] = int32(i % 50) // Cycle through columns
	}
	return indices
}

func generateIndptr(rows, cols int) []int64 {
	indptr := make([]int64, rows+1)
	for i := 0; i <= rows; i++ {
		indptr[i] = int64(i * cols)
	}
	return indptr
}

func generateFileOwnership(numFiles int) map[string]int32 {
	ownership := make(map[string]int32)
	for i := 0; i < numFiles; i++ {
		filename := fmt.Sprintf("file_%d.go", i)
		ownership[filename] = int32(i % 5) // Assign to one of 5 people
	}
	return ownership
}

func generateFileNames(count int) []string {
	names := make([]string, count)
	for i := 0; i < count; i++ {
		names[i] = fmt.Sprintf("file_%d.go", i)
	}
	return names
}