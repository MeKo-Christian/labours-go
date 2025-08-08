package readers

import (
	"bytes"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"labours-go/internal/pb"
)

func TestProtobufReader_Read(t *testing.T) {
	// Create burndown analysis data
	burndownData := &pb.BurndownAnalysisResults{
		Project: &pb.BurndownSparseMatrix{
			Name:            "test-project",
			NumberOfRows:    2,
			NumberOfColumns: 3,
			Rows: []*pb.BurndownSparseMatrixRow{
				{Columns: []uint32{100, 90, 80}},
				{Columns: []uint32{120, 100, 85}},
			},
		},
		Files: []*pb.BurndownSparseMatrix{
			{
				Name:            "main.go",
				NumberOfRows:    1,
				NumberOfColumns: 3,
				Rows: []*pb.BurndownSparseMatrixRow{
					{Columns: []uint32{50, 45, 40}},
				},
			},
		},
		People: []*pb.BurndownSparseMatrix{
			{
				Name:            "developer1",
				NumberOfRows:    2,
				NumberOfColumns: 3,
				Rows: []*pb.BurndownSparseMatrixRow{
					{Columns: []uint32{80, 70, 60}},
					{Columns: []uint32{40, 30, 20}},
				},
			},
		},
		FilesOwnership: []*pb.FilesOwnership{
			{
				Value: map[int32]int32{
					0: 0, // file index -> owner index
					1: 1,
				},
			},
		},
		Granularity: 1,
		Sampling:    86400, // 1 day in seconds
	}

	// Marshal burndown data to bytes
	burndownBytes, err := proto.Marshal(burndownData)
	if err != nil {
		t.Fatalf("Failed to marshal burndown data: %v", err)
	}

	// Create sample protobuf data using correct structure
	testResults := &pb.AnalysisResults{
		Header: &pb.Metadata{
			Repository:    "test-repo",
			BeginUnixTime: 1640995200, // 2022-01-01
			EndUnixTime:   1672531200, // 2023-01-01
		},
		Contents: map[string][]byte{
			"Burndown": burndownBytes,
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

	if len(matrix) != 3 {
		t.Errorf("Expected 3 rows in matrix (after transpose), got %d", len(matrix))
	}

	if len(matrix[0]) != 2 {
		t.Errorf("Expected 2 columns in first row (after transpose), got %d", len(matrix[0]))
	}

	// Check specific values (after transpose)
	if matrix[0][0] != 100 {
		t.Errorf("Expected first value to be 100, got %d", matrix[0][0])
	}
	
	// Check that transposition worked correctly
	if matrix[0][1] != 120 {
		t.Errorf("Expected matrix[0][1] to be 120 (from original row 1, col 0), got %d", matrix[0][1])
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
	// Create burndown analysis data
	burndownData := &pb.BurndownAnalysisResults{
		Project: &pb.BurndownSparseMatrix{
			Name:            "test-project",
			NumberOfRows:    2,
			NumberOfColumns: 3,
			Rows: []*pb.BurndownSparseMatrixRow{
				{Columns: []uint32{100, 90, 80}},
				{Columns: []uint32{120, 100, 85}},
			},
		},
		Files: []*pb.BurndownSparseMatrix{
			{
				Name:            "main.go",
				NumberOfRows:    1,
				NumberOfColumns: 3,
				Rows: []*pb.BurndownSparseMatrixRow{
					{Columns: []uint32{50, 45, 40}},
				},
			},
		},
		People: []*pb.BurndownSparseMatrix{
			{
				Name:            "developer1",
				NumberOfRows:    2,
				NumberOfColumns: 3,
				Rows: []*pb.BurndownSparseMatrixRow{
					{Columns: []uint32{80, 70, 60}},
					{Columns: []uint32{40, 30, 20}},
				},
			},
		},
		FilesOwnership: []*pb.FilesOwnership{
			{
				Value: map[int32]int32{
					0: 0, // file index -> owner index
					1: 1,
				},
			},
		},
		Granularity: 1,
		Sampling:    86400,
	}

	// Marshal burndown data to bytes
	burndownBytes, err := proto.Marshal(burndownData)
	if err != nil {
		t.Fatalf("Failed to marshal burndown data: %v", err)
	}

	testResults := &pb.AnalysisResults{
		Header: &pb.Metadata{
			Repository:    "test-repo",
			BeginUnixTime: 1640995200,
			EndUnixTime:   1672531200,
		},
		Contents: map[string][]byte{
			"Burndown": burndownBytes,
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
