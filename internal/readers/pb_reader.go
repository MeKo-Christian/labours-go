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
	if r.data.Header != nil {
		return r.data.Header.Repository
	}
	return ""
}

// GetHeader retrieves the start and end timestamps from the Protobuf metadata
func (r *ProtobufReader) GetHeader() (int64, int64) {
	if r.data.Header != nil {
		return r.data.Header.BeginUnixTime, r.data.Header.EndUnixTime
	}
	return 0, 0
}

// GetProjectBurndown retrieves the project-level burndown matrix
func (r *ProtobufReader) GetProjectBurndown() (string, [][]int) {
	// Parse burndown data from Contents
	burndownData := r.parseBurndownAnalysisResults()
	if burndownData == nil || burndownData.Project == nil {
		return "", nil
	}

	matrix := parseBurndownSparseMatrix(burndownData.Project)
	return r.GetName(), transposeMatrix(matrix)
}

// GetFilesBurndown retrieves burndown data for files
func (r *ProtobufReader) GetFilesBurndown() ([]FileBurndown, error) {
	burndownData := r.parseBurndownAnalysisResults()
	if burndownData == nil || len(burndownData.Files) == 0 {
		return nil, fmt.Errorf("no files burndown data found")
	}

	// Process each file's burndown matrix
	var fileBurndowns []FileBurndown
	for _, fileMatrix := range burndownData.Files {
		matrix := parseBurndownSparseMatrix(fileMatrix)
		transposed := transposeMatrix(matrix)
		fileBurndowns = append(fileBurndowns, FileBurndown{
			Filename: fileMatrix.Name,
			Matrix:   transposed,
		})
	}
	return fileBurndowns, nil
}

// GetPeopleBurndown retrieves burndown data for people
func (r *ProtobufReader) GetPeopleBurndown() ([]PeopleBurndown, error) {
	burndownData := r.parseBurndownAnalysisResults()
	if burndownData == nil || len(burndownData.People) == 0 {
		return nil, fmt.Errorf("no people burndown data found")
	}

	// Process each person's burndown matrix
	var peopleBurndowns []PeopleBurndown
	for _, personMatrix := range burndownData.People {
		matrix := parseBurndownSparseMatrix(personMatrix)
		transposed := transposeMatrix(matrix)
		peopleBurndowns = append(peopleBurndowns, PeopleBurndown{
			Person: personMatrix.Name,
			Matrix: transposed,
		})
	}
	return peopleBurndowns, nil
}

// GetOwnershipBurndown retrieves the ownership matrix and sequence
func (r *ProtobufReader) GetOwnershipBurndown() ([]string, map[string][][]int, error) {
	// Get people burndown data (matches Python behavior)
	peopleBurndowns, err := r.GetPeopleBurndown()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get people burndown data: %v", err)
	}

	// Extract people sequence (names) and build ownership map
	var peopleSequence []string
	ownership := make(map[string][][]int)

	for _, peopleBurndown := range peopleBurndowns {
		peopleSequence = append(peopleSequence, peopleBurndown.Person)
		
		// Transpose the matrix to match Python's .T behavior
		transposedMatrix := transposeMatrix(peopleBurndown.Matrix)
		ownership[peopleBurndown.Person] = transposedMatrix
	}

	return peopleSequence, ownership, nil
}

// GetPeopleInteraction retrieves the interaction matrix for people
func (r *ProtobufReader) GetPeopleInteraction() ([]string, [][]int, error) {
	burndownData := r.parseBurndownAnalysisResults()
	if burndownData == nil || burndownData.PeopleInteraction == nil {
		return nil, nil, fmt.Errorf("no people interaction data found")
	}

	matrix := parseCompressedSparseRowMatrix(burndownData.PeopleInteraction)
	
	// Extract people names from the burndown people data
	var peopleNames []string
	for _, person := range burndownData.People {
		peopleNames = append(peopleNames, person.Name)
	}
	
	return peopleNames, matrix, nil
}

// GetFileCooccurrence retrieves file coupling data
func (r *ProtobufReader) GetFileCooccurrence() ([]string, [][]int, error) {
	couplesData := r.parseCouplesAnalysisResults()
	if couplesData == nil || couplesData.FileCouples == nil || couplesData.FileCouples.Matrix == nil {
		return nil, nil, fmt.Errorf("no file coupling data found")
	}

	matrix := parseCompressedSparseRowMatrix(couplesData.FileCouples.Matrix)
	return couplesData.FileCouples.Index, matrix, nil
}

// GetPeopleCooccurrence retrieves people coupling data
func (r *ProtobufReader) GetPeopleCooccurrence() ([]string, [][]int, error) {
	couplesData := r.parseCouplesAnalysisResults()
	if couplesData == nil || couplesData.PeopleCouples == nil || couplesData.PeopleCouples.Matrix == nil {
		return nil, nil, fmt.Errorf("no people coupling data found")
	}

	matrix := parseCompressedSparseRowMatrix(couplesData.PeopleCouples.Matrix)
	return couplesData.PeopleCouples.Index, matrix, nil
}

// GetShotnessCooccurrence retrieves shotness coupling data
func (r *ProtobufReader) GetShotnessCooccurrence() ([]string, [][]int, error) {
	// This would require additional protobuf structure for shotness data
	return []string{}, [][]int{}, fmt.Errorf("shotness data not implemented in current protobuf format")
}

// GetShotnessRecords retrieves shotness records
func (r *ProtobufReader) GetShotnessRecords() ([]ShotnessRecord, error) {
	shotnessData := r.parseShotnessAnalysisResults()
	if shotnessData == nil || len(shotnessData.Records) == 0 {
		return []ShotnessRecord{}, fmt.Errorf("no shotness data found - ensure the input data contains shotness analysis results")
	}

	pbRecords := shotnessData.Records
	records := make([]ShotnessRecord, len(pbRecords))
	for i, pbRecord := range pbRecords {
		records[i] = ShotnessRecord{
			Type:     pbRecord.Type,
			Name:     pbRecord.Name,
			File:     pbRecord.File,
			Counters: pbRecord.Counters,
		}
	}

	return records, nil
}

// GetDeveloperStats retrieves developer statistics
func (r *ProtobufReader) GetDeveloperStats() ([]DeveloperStat, error) {
	devsData := r.parseDevsAnalysisResults()
	if devsData == nil || len(devsData.DevIndex) == 0 {
		return nil, fmt.Errorf("no developer stats found")
	}

	// Create synthetic developer stats from the available data
	stats := make([]DeveloperStat, len(devsData.DevIndex))
	for i, devName := range devsData.DevIndex {
		stats[i] = DeveloperStat{
			Name:          devName,
			Commits:       0, // Would need to aggregate from time series
			LinesAdded:    0,
			LinesRemoved:  0,
			LinesModified: 0,
			FilesTouched:  0,
			Languages:     make(map[string]int),
		}
	}

	return stats, nil
}

// GetLanguageStats retrieves language statistics
func (r *ProtobufReader) GetLanguageStats() ([]LanguageStat, error) {
	// Language stats might be part of other analysis results
	// For now, return empty as this data structure may not exist in protobuf
	return nil, fmt.Errorf("no language stats found in protobuf format")
}

// GetRuntimeStats retrieves runtime statistics
func (r *ProtobufReader) GetRuntimeStats() (map[string]float64, error) {
	if r.data.Header == nil {
		return nil, fmt.Errorf("no header found for runtime stats")
	}

	runtimeStats := make(map[string]float64)
	if r.data.Header.RunTimePerItem != nil {
		for key, value := range r.data.Header.RunTimePerItem {
			runtimeStats[key] = value
		}
	}

	return runtimeStats, nil
}

// GetDeveloperTimeSeriesData returns Python-compatible time series data for protobuf files
// This now parses real temporal data from DevsAnalysisResults.Ticks (matches Python's approach)
func (r *ProtobufReader) GetDeveloperTimeSeriesData() (*DeveloperTimeSeriesData, error) {
	// Parse real developer time series data from protobuf (like Python does)
	devsData := r.parseDevsAnalysisResults()
	if devsData == nil {
		return nil, fmt.Errorf("no developer analysis data found")
	}
	
	// Extract people list from dev_index (matches Python's people = list(self.contents["Devs"].dev_index))
	people := make([]string, len(devsData.DevIndex))
	copy(people, devsData.DevIndex)
	
	// Parse real time series data from ticks (matches Python's self.contents["Devs"].ticks.items())
	days := make(map[int]map[int]DevDay)
	
	// Iterate through all time ticks
	for tickKey, tickDevs := range devsData.Ticks {
		if tickDevs == nil {
			continue
		}
		
		// Create developer map for this time tick
		dayDevs := make(map[int]DevDay)
		
		// Iterate through all developers in this tick
		for devIndex, devTick := range tickDevs.Devs {
			if devTick == nil {
				continue
			}
			
			// Convert languages map from protobuf format to DevDay format
			languages := make(map[string][]int)
			if devTick.Languages != nil {
				for lang, langStats := range devTick.Languages {
					if langStats != nil {
						// Python format: {lang: [added, removed, changed]}
						languages[lang] = []int{
							int(langStats.Added),
							int(langStats.Removed), 
							int(langStats.Changed),
						}
					}
				}
			}
			
			// Convert protobuf DevTick to Go DevDay format (matches Python's DevDay structure)
			dayDevs[int(devIndex)] = DevDay{
				Commits:       int(devTick.Commits),
				LinesAdded:    int(devTick.Stats.Added),
				LinesRemoved:  int(devTick.Stats.Removed),
				LinesModified: int(devTick.Stats.Changed),
				Languages:     languages,
			}
		}
		
		// Store this day's data using the real time tick key
		days[int(tickKey)] = dayDevs
	}
	
	// Return the same format as Python: (people, days)
	return &DeveloperTimeSeriesData{
		People: people,
		Days:   days,
	}, nil
}

// parseBurndownSparseMatrix converts protobuf BurndownSparseMatrix to dense matrix
// This matches the Python _parse_burndown_matrix logic
func parseBurndownSparseMatrix(matrix *pb.BurndownSparseMatrix) [][]int {
	if matrix == nil {
		return [][]int{}
	}

	result := make([][]int, matrix.NumberOfRows)
	for i := range result {
		result[i] = make([]int, matrix.NumberOfColumns)
	}

	// Convert from row/column format to dense matrix (matches Python logic)
	for y, row := range matrix.Rows {
		if y >= int(matrix.NumberOfRows) {
			break
		}
		for x, value := range row.Columns {
			if x >= int(matrix.NumberOfColumns) {
				break
			}
			result[y][x] = int(value)
		}
	}

	return result
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

// parseBurndownAnalysisResults extracts and parses burndown data from the Contents map
func (r *ProtobufReader) parseBurndownAnalysisResults() *pb.BurndownAnalysisResults {
	if r.data == nil || r.data.Contents == nil {
		return nil
	}
	
	// Look for burndown data in Contents
	burndownBytes, exists := r.data.Contents["Burndown"]
	if !exists {
		return nil
	}
	
	// Parse the burndown data
	var burndownData pb.BurndownAnalysisResults
	if err := proto.Unmarshal(burndownBytes, &burndownData); err != nil {
		return nil
	}
	
	return &burndownData
}

// GetBurndownParameters retrieves burndown parameters in Python-compatible format
func (r *ProtobufReader) GetBurndownParameters() (burndown.BurndownParameters, error) {
	burndownData := r.parseBurndownAnalysisResults()
	if burndownData == nil {
		return burndown.BurndownParameters{}, fmt.Errorf("no burndown data found")
	}

	// Calculate appropriate tick size based on time span and matrix dimensions
	tickSize := float64(burndownData.TickSize) / 1e9 // Convert nanoseconds to seconds
	
	if r.data.Header != nil {
		// Calculate tick size from actual time span and expected data points
		timeSpan := float64(r.data.Header.EndUnixTime - r.data.Header.BeginUnixTime)
		
		// Get matrix dimensions to calculate appropriate tick size
		if burndownData.Project != nil {
			matrixCols := burndownData.Project.NumberOfColumns
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
	burndownData := r.parseBurndownAnalysisResults()
	if burndownData == nil || burndownData.Project == nil {
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

// parseCouplesAnalysisResults extracts and parses couples data from the Contents map
func (r *ProtobufReader) parseCouplesAnalysisResults() *pb.CouplesAnalysisResults {
	if r.data == nil || r.data.Contents == nil {
		return nil
	}
	
	// Look for couples data in Contents
	couplesBytes, exists := r.data.Contents["Couples"]
	if !exists {
		return nil
	}
	
	// Parse the couples data
	var couplesData pb.CouplesAnalysisResults
	if err := proto.Unmarshal(couplesBytes, &couplesData); err != nil {
		return nil
	}
	
	return &couplesData
}

// parseShotnessAnalysisResults extracts and parses shotness data from the Contents map
func (r *ProtobufReader) parseShotnessAnalysisResults() *pb.ShotnessAnalysisResults {
	if r.data == nil || r.data.Contents == nil {
		return nil
	}
	
	// Look for shotness data in Contents
	shotnessBytes, exists := r.data.Contents["Shotness"]
	if !exists {
		return nil
	}
	
	// Parse the shotness data
	var shotnessData pb.ShotnessAnalysisResults
	if err := proto.Unmarshal(shotnessBytes, &shotnessData); err != nil {
		return nil
	}
	
	return &shotnessData
}

// parseDevsAnalysisResults extracts and parses devs data from the Contents map
func (r *ProtobufReader) parseDevsAnalysisResults() *pb.DevsAnalysisResults {
	if r.data == nil || r.data.Contents == nil {
		return nil
	}
	
	// Look for devs data in Contents
	devsBytes, exists := r.data.Contents["Devs"]
	if !exists {
		return nil
	}
	
	// Parse the devs data
	var devsData pb.DevsAnalysisResults
	if err := proto.Unmarshal(devsBytes, &devsData); err != nil {
		return nil
	}
	
	return &devsData
}
