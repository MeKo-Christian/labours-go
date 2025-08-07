package readers

import (
	"fmt"
	"io"

	"github.com/spf13/viper"
	"google.golang.org/protobuf/proto"
	"labours-go/internal/burndown"
	"labours-go/internal/pb"
	"labours-go/internal/progress"
)

type ProtobufReader struct {
	data *pb.AnalysisResults
}

// Read loads the Protobuf data into the ProtobufReader structure
func (r *ProtobufReader) Read(file io.Reader) error {
	// Initialize progress tracking for file reading
	quiet := viper.GetBool("quiet")
	progEstimator := progress.NewProgressEstimator(!quiet)
	
	// Start reading operation
	progEstimator.StartOperation("Reading protobuf data", 2) // read + parse phases
	
	progEstimator.UpdateProgress(1)
	allBytes, err := io.ReadAll(file)
	if err != nil {
		progEstimator.FinishOperation()
		return fmt.Errorf("error reading Protobuf file: %v", err)
	}

	progEstimator.UpdateProgress(1)
	var results pb.AnalysisResults
	if err := proto.Unmarshal(allBytes, &results); err != nil {
		progEstimator.FinishOperation()
		return fmt.Errorf("error unmarshalling Protobuf: %v", err)
	}

	r.data = &results
	progEstimator.FinishOperation()
	return nil
}

// GetName retrieves the repository name from the Protobuf metadata
func (r *ProtobufReader) GetName() string {
	if r.data.Metadata != nil {
		return r.data.Metadata.Repository
	}
	return ""
}

// GetHeader retrieves the start and end timestamps from the Protobuf metadata
func (r *ProtobufReader) GetHeader() (int64, int64) {
	if r.data.Metadata != nil {
		return r.data.Metadata.BeginUnixTime, r.data.Metadata.EndUnixTime
	}
	return 0, 0
}

// GetProjectBurndown retrieves the project-level burndown matrix
func (r *ProtobufReader) GetProjectBurndown() (string, [][]int) {
	if r.data.Burndown == nil || r.data.Burndown.Project == nil {
		return "", nil
	}

	matrix := parseCompressedSparseRowMatrix(r.data.Burndown.Project)
	return r.GetName(), matrix
}

// GetFilesBurndown retrieves burndown data for files
func (r *ProtobufReader) GetFilesBurndown() ([]FileBurndown, error) {
	if r.data.Burndown == nil || r.data.Burndown.Files == nil {
		return nil, fmt.Errorf("no files burndown data found")
	}

	matrix := parseCompressedSparseRowMatrix(r.data.Burndown.Files)

	// Create individual file burndown entries
	var fileBurndowns []FileBurndown
	for i, filename := range r.data.FileNames {
		if i < len(matrix) {
			fileBurndowns = append(fileBurndowns, FileBurndown{
				Filename: filename,
				Matrix:   [][]int{matrix[i]}, // Each row represents one file
			})
		}
	}
	return fileBurndowns, nil
}

// GetPeopleBurndown retrieves burndown data for people
func (r *ProtobufReader) GetPeopleBurndown() ([]PeopleBurndown, error) {
	if r.data.Burndown == nil || r.data.Burndown.People == nil {
		return nil, fmt.Errorf("no people burndown data found")
	}

	matrix := parseCompressedSparseRowMatrix(r.data.Burndown.People)

	// Create individual people burndown entries
	var peopleBurndowns []PeopleBurndown
	for i, personName := range r.data.PeopleNames {
		if i < len(matrix) {
			peopleBurndowns = append(peopleBurndowns, PeopleBurndown{
				Person: personName,
				Matrix: [][]int{matrix[i]}, // Each row represents one person
			})
		}
	}
	return peopleBurndowns, nil
}

// GetOwnershipBurndown retrieves the ownership matrix and sequence
func (r *ProtobufReader) GetOwnershipBurndown() ([]string, map[string][][]int, error) {
	if r.data.Burndown == nil || r.data.Burndown.FilesOwnership == nil {
		return nil, nil, fmt.Errorf("no ownership data found")
	}

	peopleSequence := r.data.PeopleNames
	ownership := make(map[string][][]int)

	// Use the files ownership mapping to create ownership matrices
	for filename, ownerIndex := range r.data.Burndown.FilesOwnership.Value {
		if int(ownerIndex) < len(peopleSequence) {
			ownerName := peopleSequence[ownerIndex]
			if _, exists := ownership[ownerName]; !exists {
				ownership[ownerName] = [][]int{}
			}
			// For simplicity, create a basic matrix - this would need actual data
			ownership[ownerName] = append(ownership[ownerName], []int{1}) // Placeholder
		}
		_ = filename // Avoid unused variable warning
	}

	return peopleSequence, ownership, nil
}

// GetPeopleInteraction retrieves the interaction matrix for people
func (r *ProtobufReader) GetPeopleInteraction() ([]string, [][]int, error) {
	if r.data.Burndown == nil || r.data.Burndown.PeopleInteraction == nil {
		return nil, nil, fmt.Errorf("no people interaction data found")
	}

	matrix := parseCompressedSparseRowMatrix(r.data.Burndown.PeopleInteraction)
	return r.data.PeopleNames, matrix, nil
}

// GetFileCooccurrence retrieves file coupling data
func (r *ProtobufReader) GetFileCooccurrence() ([]string, [][]int, error) {
	if r.data.Couples == nil || r.data.Couples.FileCouples == nil {
		return nil, nil, fmt.Errorf("no file coupling data found")
	}

	matrix := parseCompressedSparseRowMatrix(r.data.Couples.FileCouples)
	return r.data.Couples.FileNames, matrix, nil
}

// GetPeopleCooccurrence retrieves people coupling data
func (r *ProtobufReader) GetPeopleCooccurrence() ([]string, [][]int, error) {
	if r.data.Couples == nil {
		return nil, nil, fmt.Errorf("no people coupling data found")
	}

	// For people coupling, we would need additional matrix data in the protobuf
	// For now, return empty data
	return r.data.PeopleNames, [][]int{}, nil
}

// GetShotnessCooccurrence retrieves shotness coupling data
func (r *ProtobufReader) GetShotnessCooccurrence() ([]string, [][]int, error) {
	// This would require additional protobuf structure for shotness data
	return []string{}, [][]int{}, fmt.Errorf("shotness data not implemented in current protobuf format")
}

// GetShotnessRecords retrieves shotness records
func (r *ProtobufReader) GetShotnessRecords() ([]ShotnessRecord, error) {
	if r.data.GetShotness() == nil || len(r.data.GetShotness().GetRecords()) == 0 {
		return []ShotnessRecord{}, fmt.Errorf("no shotness data found - ensure the input data contains shotness analysis results")
	}

	pbRecords := r.data.GetShotness().GetRecords()
	records := make([]ShotnessRecord, len(pbRecords))
	for i, pbRecord := range pbRecords {
		records[i] = ShotnessRecord{
			Type:     pbRecord.GetType(),
			Name:     pbRecord.GetName(),
			File:     pbRecord.GetFile(),
			Counters: pbRecord.GetCounters(),
		}
	}

	return records, nil
}

// GetDeveloperStats retrieves developer statistics
func (r *ProtobufReader) GetDeveloperStats() ([]DeveloperStat, error) {
	if len(r.data.DeveloperStats) == 0 {
		return nil, fmt.Errorf("no developer stats found")
	}

	stats := make([]DeveloperStat, len(r.data.DeveloperStats))
	for i, dev := range r.data.DeveloperStats {
		stats[i] = DeveloperStat{
			Name:          dev.Name,
			Commits:       int(dev.Commits),
			LinesAdded:    int(dev.LinesAdded),
			LinesRemoved:  int(dev.LinesRemoved),
			LinesModified: int(dev.LinesModified),
			FilesTouched:  int(dev.FilesTouched),
			Languages:     make(map[string]int),
		}
		// Copy language map
		for lang, count := range dev.Languages {
			stats[i].Languages[lang] = int(count)
		}
	}

	return stats, nil
}

// GetLanguageStats retrieves language statistics
func (r *ProtobufReader) GetLanguageStats() ([]LanguageStat, error) {
	if len(r.data.LanguageStats) == 0 {
		return nil, fmt.Errorf("no language stats found")
	}

	stats := make([]LanguageStat, len(r.data.LanguageStats))
	for i, lang := range r.data.LanguageStats {
		stats[i] = LanguageStat{
			Language: lang.Language,
			Lines:    int(lang.Lines),
		}
	}

	return stats, nil
}

// GetRuntimeStats retrieves runtime statistics
func (r *ProtobufReader) GetRuntimeStats() (map[string]float64, error) {
	if r.data.Metadata == nil {
		return nil, fmt.Errorf("no metadata found for runtime stats")
	}

	return map[string]float64{
		"total_runtime": float64(r.data.Metadata.RunTime),
	}, nil
}

// parseCompressedSparseRowMatrix converts protobuf CompressedSparseRowMatrix to dense matrix
func parseCompressedSparseRowMatrix(matrix *pb.CompressedSparseRowMatrix) [][]int {
	if matrix == nil {
		return [][]int{}
	}

	result := make([][]int, matrix.NumberOfRows)
	for i := range result {
		result[i] = make([]int, matrix.NumberOfColumns)
	}

	// Convert from CSR format to dense matrix with bounds checking
	for i := int32(0); i < matrix.NumberOfRows; i++ {
		if int(i+1) >= len(matrix.Indptr) {
			break
		}
		start := matrix.Indptr[i]
		end := matrix.Indptr[i+1]

		for j := start; j < end; j++ {
			if int(j) >= len(matrix.Indices) || int(j) >= len(matrix.Data) {
				break
			}
			col := matrix.Indices[j]
			if int(col) >= int(matrix.NumberOfColumns) {
				continue
			}
			value := matrix.Data[j]
			result[i][col] = int(value)
		}
	}

	return result
}

// GetBurndownParameters retrieves burndown parameters in Python-compatible format
func (r *ProtobufReader) GetBurndownParameters() (burndown.BurndownParameters, error) {
	if r.data.Burndown == nil {
		return burndown.BurndownParameters{}, fmt.Errorf("no burndown data found")
	}

	// Calculate appropriate tick size based on time span and matrix dimensions
	tickSize := float64(r.data.Burndown.TickSize) / 1e9 // Convert nanoseconds to seconds
	
	if r.data.Metadata != nil {
		// Calculate tick size from actual time span and expected data points
		timeSpan := float64(r.data.Metadata.EndUnixTime - r.data.Metadata.BeginUnixTime)
		
		// Get matrix dimensions to calculate appropriate tick size
		if r.data.Burndown.Project != nil {
			matrixCols := r.data.Burndown.Project.NumberOfColumns
			if matrixCols > 1 && timeSpan > 0 {
				// Calculate tick size as time span divided by number of time points
				calculatedTick := timeSpan / float64(matrixCols-1)
				
				// Use calculated tick size if it's reasonable, otherwise use original or fallback
				if calculatedTick > 0 && calculatedTick < timeSpan {
					tickSize = calculatedTick
				}
			}
		}
	}
	
	// Fallback if we still don't have a reasonable tick size
	if tickSize <= 0 || tickSize > 365*24*3600 { // More than a year per tick seems wrong
		tickSize = 86400 // Default to 1 day in seconds
	}

	// Debug output removed - tick size calculation working correctly
	
	return burndown.BurndownParameters{
		Sampling:    1,        // Daily sampling (1 day)
		Granularity: 1,        // 1 day granularity
		TickSize:    tickSize, // Calculated or fallback tick size
	}, nil
}

// GetProjectBurndownWithHeader retrieves project burndown with full header info
func (r *ProtobufReader) GetProjectBurndownWithHeader() (burndown.BurndownHeader, string, [][]int, error) {
	if r.data.Burndown == nil || r.data.Burndown.Project == nil {
		return burndown.BurndownHeader{}, "", nil, fmt.Errorf("no project burndown data found")
	}

	// Get header information
	start, last := r.GetHeader()
	params, err := r.GetBurndownParameters()
	if err != nil {
		return burndown.BurndownHeader{}, "", nil, err
	}

	header := burndown.BurndownHeader{
		Start:       start,
		Last:        last,
		Sampling:    params.Sampling,
		Granularity: params.Granularity,
		TickSize:    params.TickSize,
	}

	// Get matrix and name
	name, matrix := r.GetProjectBurndown()

	return header, name, matrix, nil
}
