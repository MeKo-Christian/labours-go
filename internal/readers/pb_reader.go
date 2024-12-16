package readers

import (
	"fmt"
	"io"

	"labours-go/internal/pb"

	"github.com/gogo/protobuf/proto"
)

type ProtobufReader struct {
	data *pb.AnalysisResults
}

// Read loads the Protobuf data into the ProtobufReader structure
func (r *ProtobufReader) Read(file io.Reader) error {
	allBytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading Protobuf file: %v", err)
	}

	var results pb.AnalysisResults
	if err := proto.Unmarshal(allBytes, &results); err != nil {
		return fmt.Errorf("error unmarshalling Protobuf: %v", err)
	}

	r.data = &results
	return nil
}

// GetName retrieves the repository name from the Protobuf header
func (r *ProtobufReader) GetName() string {
	if r.data.Header != nil {
		return r.data.Header.Repository
	}
	return ""
}

// GetHeader retrieves the start and end timestamps from the Protobuf header
func (r *ProtobufReader) GetHeader() (int64, int64) {
	if r.data.Header != nil {
		return r.data.Header.BeginUnixTime, r.data.Header.EndUnixTime
	}
	return 0, 0
}

// GetProjectBurndown retrieves the project-level burndown matrix
func (r *ProtobufReader) GetProjectBurndown() (string, [][]int) {
	if r.data.Contents == nil {
		return "", nil
	}

	// Parse the "Burndown" data
	burndownBytes, ok := r.data.Contents["Burndown"]
	if !ok {
		return "", nil
	}

	var burndown pb.BurndownAnalysisResults
	if err := proto.Unmarshal(burndownBytes, &burndown); err != nil {
		fmt.Printf("error parsing Burndown data: %v\n", err)
		return "", nil
	}

	matrix := parseSparseMatrix(burndown.Project)
	return r.GetName(), matrix
}

// GetFilesBurndown retrieves burndown data for files
func (r *ProtobufReader) GetFilesBurndown() ([]FileBurndown, error) {
	if r.data.Contents == nil {
		return nil, fmt.Errorf("no content found in Protobuf data")
	}

	burndownBytes, ok := r.data.Contents["Burndown"]
	if !ok {
		return nil, fmt.Errorf("no Burndown data found")
	}

	var burndown pb.BurndownAnalysisResults
	if err := proto.Unmarshal(burndownBytes, &burndown); err != nil {
		return nil, fmt.Errorf("error parsing Burndown data: %v", err)
	}

	var fileBurndowns []FileBurndown
	for _, file := range burndown.Files {
		matrix := parseSparseMatrix(file)
		fileBurndowns = append(fileBurndowns, FileBurndown{
			Filename: file.Name,
			Matrix:   matrix,
		})
	}
	return fileBurndowns, nil
}

// GetPeopleBurndown retrieves burndown data for people
func (r *ProtobufReader) GetPeopleBurndown() ([]PeopleBurndown, error) {
	if r.data.Contents == nil {
		return nil, fmt.Errorf("no content found in Protobuf data")
	}

	burndownBytes, ok := r.data.Contents["Burndown"]
	if !ok {
		return nil, fmt.Errorf("no Burndown data found")
	}

	var burndown pb.BurndownAnalysisResults
	if err := proto.Unmarshal(burndownBytes, &burndown); err != nil {
		return nil, fmt.Errorf("error parsing Burndown data: %v", err)
	}

	var peopleBurndowns []PeopleBurndown
	for _, person := range burndown.People {
		matrix := parseSparseMatrix(person)
		peopleBurndowns = append(peopleBurndowns, PeopleBurndown{
			Person: person.Name,
			Matrix: matrix,
		})
	}
	return peopleBurndowns, nil
}

// GetOwnershipBurndown retrieves the ownership matrix and sequence
func (r *ProtobufReader) GetOwnershipBurndown() ([]string, map[string][][]int, error) {
	if r.data.Contents == nil {
		return nil, nil, fmt.Errorf("no content found in Protobuf data")
	}

	burndownBytes, ok := r.data.Contents["Burndown"]
	if !ok {
		return nil, nil, fmt.Errorf("no Burndown data found")
	}

	var burndown pb.BurndownAnalysisResults
	if err := proto.Unmarshal(burndownBytes, &burndown); err != nil {
		return nil, nil, fmt.Errorf("error parsing Burndown data: %v", err)
	}

	peopleSequence := []string{}
	ownership := make(map[string][][]int)

	for _, person := range burndown.People {
		matrix := parseSparseMatrix(person)
		ownership[person.Name] = matrix
		peopleSequence = append(peopleSequence, person.Name)
	}

	return peopleSequence, ownership, nil
}

// GetPeopleInteraction retrieves the interaction matrix for people
func (r *ProtobufReader) GetPeopleInteraction() ([]string, [][]int, error) {
	if r.data.Contents == nil {
		return nil, nil, fmt.Errorf("no content found in Protobuf data")
	}

	burndownBytes, ok := r.data.Contents["Burndown"]
	if !ok {
		return nil, nil, fmt.Errorf("no Burndown data found")
	}

	var burndown pb.BurndownAnalysisResults
	if err := proto.Unmarshal(burndownBytes, &burndown); err != nil {
		return nil, nil, fmt.Errorf("error parsing Burndown data: %v", err)
	}

	matrix := parseCompressedSparseRowMatrix(burndown.PeopleInteraction)
	return r.GetPeopleSequence(), matrix, nil
}

// GetPeopleSequence retrieves the sequence of people
func (r *ProtobufReader) GetPeopleSequence() []string {
	if r.data.Contents == nil {
		return nil
	}

	burndownBytes, ok := r.data.Contents["Burndown"]
	if !ok {
		return nil
	}

	var burndown pb.BurndownAnalysisResults
	if err := proto.Unmarshal(burndownBytes, &burndown); err != nil {
		return nil
	}

	sequence := []string{}
	for _, person := range burndown.People {
		sequence = append(sequence, person.Name)
	}
	return sequence
}

func (r *ProtobufReader) GetFileCooccurrence() ([]string, [][]int, error) {
	if r.data.Contents == nil {
		return nil, nil, fmt.Errorf("no content found in Protobuf data")
	}

	couplesBytes, ok := r.data.Contents["Couples"]
	if !ok {
		return nil, nil, fmt.Errorf("no Couples data found")
	}

	var couples pb.CouplesAnalysisResults
	if err := proto.Unmarshal(couplesBytes, &couples); err != nil {
		return nil, nil, fmt.Errorf("error parsing Couples data: %v", err)
	}

	matrix := parseCompressedSparseRowMatrix(couples.FileCouples.Matrix)
	return couples.FileCouples.Index, matrix, nil
}

func (r *ProtobufReader) GetPeopleCooccurrence() ([]string, [][]int, error) {
	if r.data.Contents == nil {
		return nil, nil, fmt.Errorf("no content found in Protobuf data")
	}

	couplesBytes, ok := r.data.Contents["Couples"]
	if !ok {
		return nil, nil, fmt.Errorf("no Couples data found")
	}

	var couples pb.CouplesAnalysisResults
	if err := proto.Unmarshal(couplesBytes, &couples); err != nil {
		return nil, nil, fmt.Errorf("error parsing Couples data: %v", err)
	}

	matrix := parseCompressedSparseRowMatrix(couples.PeopleCouples.Matrix)
	return couples.PeopleCouples.Index, matrix, nil
}

func (r *ProtobufReader) GetShotnessCooccurrence() ([]string, [][]int, error) {
	if r.data.Contents == nil {
		return nil, nil, fmt.Errorf("no content found in Protobuf data")
	}

	shotnessBytes, ok := r.data.Contents["Shotness"]
	if !ok {
		return nil, nil, fmt.Errorf("no Shotness data found")
	}

	var shotness pb.ShotnessAnalysisResults
	if err := proto.Unmarshal(shotnessBytes, &shotness); err != nil {
		return nil, nil, fmt.Errorf("error parsing Shotness data: %v", err)
	}

	records := shotness.Records
	names := make([]string, len(records))
	matrix := make([][]int, len(records))

	for i, record := range records {
		names[i] = fmt.Sprintf("%s:%s:%s", record.Type, record.Name, record.File)
		row := make([]int, len(record.Counters))
		for k, v := range record.Counters {
			row[k] = int(v)
		}
		matrix[i] = row
	}
	return names, matrix, nil
}

func (r *ProtobufReader) GetShotnessStats() ([][]int, error) {
	if r.data.Contents == nil {
		return nil, fmt.Errorf("no content found in Protobuf data")
	}

	shotnessBytes, ok := r.data.Contents["Shotness"]
	if !ok {
		return nil, fmt.Errorf("no Shotness data found")
	}

	var shotness pb.ShotnessAnalysisResults
	if err := proto.Unmarshal(shotnessBytes, &shotness); err != nil {
		return nil, fmt.Errorf("error parsing Shotness data: %v", err)
	}

	stats := make([][]int, len(shotness.Records))
	for i, record := range shotness.Records {
		row := make([]int, len(record.Counters))
		for _, value := range record.Counters {
			row = append(row, int(value))
		}
		stats[i] = row
	}
	return stats, nil
}

func (r *ProtobufReader) GetDeveloperStats() ([]DeveloperStat, error) {
	if r.data == nil || r.data.Contents == nil {
		return nil, fmt.Errorf("no content found in Protobuf data")
	}

	devsBytes, ok := r.data.Contents["Devs"]
	if !ok {
		return nil, fmt.Errorf("no Developer data found")
	}

	var devs pb.DevsAnalysisResults
	if err := proto.Unmarshal(devsBytes, &devs); err != nil {
		return nil, fmt.Errorf("error parsing Developer data: %v", err)
	}

	if devs.DevIndex == nil || len(devs.DevIndex) == 0 {
		return nil, fmt.Errorf("Developer index is missing or empty")
	}

	// Aggregate developer stats
	devStats := make(map[string]*DeveloperStat)
	for _, tick := range devs.Ticks {
		for devID, dev := range tick.Devs {
			name := devs.DevIndex[devID]
			if _, exists := devStats[name]; !exists {
				devStats[name] = &DeveloperStat{
					Name: name,
				}
			}
			stat := devStats[name]
			stat.Commits += int(dev.Commits)
			stat.LinesAdded += int(dev.Stats.GetAdded())
			stat.LinesRemoved += int(dev.Stats.GetRemoved())
			stat.LinesModified += int(dev.Stats.GetChanged())
		}
	}

	for _, tick := range devs.Ticks {
		for devID, dev := range tick.Devs {
			name := devs.DevIndex[devID]
			if _, exists := devStats[name]; !exists {
				devStats[name] = &DeveloperStat{
					Name:      name,
					Languages: make(map[string]int),
				}
			}
			stat := devStats[name]
			stat.Commits += int(dev.Commits)
			stat.LinesAdded += int(dev.Stats.GetAdded())
			stat.LinesRemoved += int(dev.Stats.GetRemoved())
			stat.LinesModified += int(dev.Stats.GetChanged())

			// Language stats aggregation
			for lang, langStat := range dev.Languages {
				stat.Languages[lang] += int(langStat.GetAdded()) + int(langStat.GetRemoved()) + int(langStat.GetChanged())
			}
		}
	}

	// Convert map to slice for return
	results := make([]DeveloperStat, 0, len(devStats))
	for _, stat := range devStats {
		results = append(results, *stat)
	}

	return results, nil
}

func (r *ProtobufReader) GetLanguageStats() ([]LanguageStat, error) {
	if r.data.Contents == nil {
		return nil, fmt.Errorf("no content found in Protobuf data")
	}

	devsBytes, ok := r.data.Contents["Devs"]
	if !ok {
		return nil, fmt.Errorf("no Developer data found")
	}

	var devs pb.DevsAnalysisResults
	if err := proto.Unmarshal(devsBytes, &devs); err != nil {
		return nil, fmt.Errorf("error parsing Developer data: %v", err)
	}

	langStats := map[string]int{}
	for _, tick := range devs.Ticks {
		for _, dev := range tick.Devs {
			for lang, stats := range dev.Languages {
				langStats[lang] += int(stats.Added) + int(stats.Removed) + int(stats.Changed)
			}
		}
	}

	var result []LanguageStat
	for lang, lines := range langStats {
		result = append(result, LanguageStat{
			Language: lang,
			Lines:    lines,
		})
	}
	return result, nil
}

func (r *ProtobufReader) GetRuntimeStats() (map[string]float64, error) {
	if r.data.Header == nil {
		return nil, fmt.Errorf("no header data found")
	}
	return r.data.Header.RunTimePerItem, nil
}

func parseSparseMatrix(matrix *pb.BurndownSparseMatrix) [][]int {
	rows := make([][]int, matrix.NumberOfRows)
	for i, row := range matrix.Rows {
		rows[i] = make([]int, matrix.NumberOfColumns)
		for j, col := range row.Columns {
			rows[i][j] = int(col)
		}
	}
	return rows
}

func parseCompressedSparseRowMatrix(matrix *pb.CompressedSparseRowMatrix) [][]int {
	result := make([][]int, matrix.NumberOfRows)
	for i := range result {
		result[i] = make([]int, matrix.NumberOfColumns)
	}
	for i, data := range matrix.Data {
		row := int(matrix.Indptr[i])
		col := int(matrix.Indices[i])
		result[row][col] = int(data)
	}
	return result
}
