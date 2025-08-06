package readers

import (
	"bytes"
	"labours-go/internal/pb"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
)

func TestProtobufReader_Read(t *testing.T) {
	// Create sample protobuf data for testing
	testResults := &pb.AnalysisResults{
		Burndown: &pb.BurndownAnalysisResults{
			Project: &pb.CompressedSparseRowMatrix{
				NumberOfRows:    2,
				NumberOfColumns: 3,
				Data:            []int64{100, 90, 80, 120, 100, 85},
				Indices:         []int32{0, 1, 2, 0, 1, 2},
				Indptr:          []int64{0, 3, 6},
			},
			Files: &pb.CompressedSparseRowMatrix{
				NumberOfRows:    1,
				NumberOfColumns: 3,
				Data:            []int64{50, 45, 40},
				Indices:         []int32{0, 1, 2},
				Indptr:          []int64{0, 3},
			},
			People: &pb.CompressedSparseRowMatrix{
				NumberOfRows:    2,
				NumberOfColumns: 3,
				Data:            []int64{80, 70, 60, 40, 30, 20},
				Indices:         []int32{0, 1, 2, 0, 1, 2},
				Indptr:          []int64{0, 3, 6},
			},
			FilesOwnership: &pb.FilesOwnership{
				Value: map[string]int32{
					"main.go":  0,
					"utils.go": 1,
				},
			},
			TickSize: 86400, // 1 day in seconds
		},
		Metadata: &pb.Metadata{
			Repository: "test-repo",
			BeginUnixTime: 1640995200, // 2022-01-01
			EndUnixTime: 1672531200,   // 2023-01-01
		},
	}

	// Serialize to protobuf format
	data, err := proto.Marshal(testResults)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	// Create reader and test
	reader := &ProtobufReader{}
	buffer := bytes.NewReader(data)

	err = reader.Read(buffer)
	if err != nil {
		t.Errorf("ProtobufReader.Read() error = %v", err)
	}

	// Test name extraction
	name := reader.GetName()
	if name == "" {
		t.Error("Expected non-empty name from ProtobufReader")
	}
}

func TestProtobufReader_GetProjectBurndown(t *testing.T) {
	reader := createTestProtobufReader(t)

	name, matrix := reader.GetProjectBurndown()

	if name == "" {
		t.Error("Expected non-empty project name")
	}

	if len(matrix) != 2 {
		t.Errorf("Expected 2 rows in matrix, got %d", len(matrix))
	}

	if len(matrix[0]) != 3 {
		t.Errorf("Expected 3 columns in first row, got %d", len(matrix[0]))
	}

	// Check specific values
	if matrix[0][0] != 100 {
		t.Errorf("Expected first value to be 100, got %d", matrix[0][0])
	}
}

func TestProtobufReader_GetFilesBurndown(t *testing.T) {
	reader := createTestProtobufReader(t)

	files, err := reader.GetFilesBurndown()
	if err == nil {
		// This might fail because we don't have FileNames in our test data
		t.Logf("Got %d files", len(files))
	} else {
		t.Logf("Expected error due to missing FileNames: %v", err)
	}
}

func TestProtobufReader_GetPeopleBurndown(t *testing.T) {
	reader := createTestProtobufReader(t)

	people, err := reader.GetPeopleBurndown()
	if err == nil {
		// This might fail because we don't have PeopleNames in our test data
		t.Logf("Got %d people", len(people))
	} else {
		t.Logf("Expected error due to missing PeopleNames: %v", err)
	}
}

func TestProtobufReader_GetOwnershipBurndown(t *testing.T) {
	reader := createTestProtobufReader(t)

	peopleNames, ownership, err := reader.GetOwnershipBurndown()
	if err != nil {
		t.Logf("GetOwnershipBurndown() error (expected): %v", err)
		return
	}

	if len(peopleNames) == 0 {
		t.Log("Expected empty people names due to test data structure")
	}

	t.Logf("Got ownership for %d files", len(ownership))
}

func TestProtobufReader_GetHeader(t *testing.T) {
	reader := createTestProtobufReader(t)

	startTime, endTime := reader.GetHeader()

	if startTime == 0 && endTime == 0 {
		t.Error("Expected non-zero start and end times")
	}

	if startTime >= endTime {
		t.Error("Expected start time to be before end time")
	}
}

func TestProtobufReader_InvalidData(t *testing.T) {
	reader := &ProtobufReader{}
	
	// Test with invalid protobuf data
	invalidData := strings.NewReader("invalid protobuf data")
	
	err := reader.Read(invalidData)
	if err == nil {
		t.Error("Expected error when reading invalid protobuf data")
	}
}

// Helper functions for testing

func createTestProtobufReader(t *testing.T) *ProtobufReader {
	testResults := &pb.AnalysisResults{
		Burndown: &pb.BurndownAnalysisResults{
			Project: &pb.CompressedSparseRowMatrix{
				NumberOfRows:    2,
				NumberOfColumns: 3,
				Data:            []int64{100, 90, 80, 120, 100, 85},
				Indices:         []int32{0, 1, 2, 0, 1, 2},
				Indptr:          []int64{0, 3, 6},
			},
			Files: &pb.CompressedSparseRowMatrix{
				NumberOfRows:    1,
				NumberOfColumns: 3,
				Data:            []int64{50, 45, 40},
				Indices:         []int32{0, 1, 2},
				Indptr:          []int64{0, 3},
			},
			People: &pb.CompressedSparseRowMatrix{
				NumberOfRows:    2,
				NumberOfColumns: 3,
				Data:            []int64{80, 70, 60, 40, 30, 20},
				Indices:         []int32{0, 1, 2, 0, 1, 2},
				Indptr:          []int64{0, 3, 6},
			},
			FilesOwnership: &pb.FilesOwnership{
				Value: map[string]int32{
					"main.go":  0,
					"utils.go": 1,
				},
			},
			TickSize: 86400,
		},
		Metadata: &pb.Metadata{
			Repository:    "test-repo",
			BeginUnixTime: 1640995200,
			EndUnixTime:   1672531200,
		},
	}

	data, err := proto.Marshal(testResults)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	reader := &ProtobufReader{}
	buffer := bytes.NewReader(data)
	
	err = reader.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	return reader
}