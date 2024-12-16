package readers

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type YamlReader struct {
	data map[string]interface{}
}

func (r *YamlReader) Read(file io.Reader) error {
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&r.data); err != nil {
		return fmt.Errorf("error decoding YAML: %v", err)
	}
	return nil
}

func (r *YamlReader) GetName() string {
	herculesData, ok := r.data["hercules"].(map[string]interface{})
	if !ok {
		return ""
	}
	return herculesData["repository"].(string)
}

func (r *YamlReader) GetHeader() (int64, int64) {
	herculesData, ok := r.data["hercules"].(map[string]interface{})
	if !ok {
		return 0, 0
	}
	begin := int64(herculesData["begin_unix_time"].(int))
	end := int64(herculesData["end_unix_time"].(int))
	return begin, end
}

func (r *YamlReader) GetProjectBurndown() (string, [][]int) {
	burndownData, ok := r.data["Burndown"].(map[string]interface{})
	if !ok {
		return "", nil
	}
	repo := r.GetName()
	matrix := parseBurndownMatrix(burndownData["project"].(string))
	return repo, matrix
}

func (r *YamlReader) GetFilesBurndown() ([]FileBurndown, error) {
	burndownData, ok := r.data["Burndown"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing Burndown data in YAML")
	}
	filesData, ok := burndownData["files"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing files data in Burndown")
	}

	var fileBurndowns []FileBurndown
	for filename, matrixData := range filesData {
		matrix := parseBurndownMatrix(matrixData.(string))
		fileBurndowns = append(fileBurndowns, FileBurndown{
			Filename: filename,
			Matrix:   matrix,
		})
	}
	return fileBurndowns, nil
}

func (r *YamlReader) GetPeopleBurndown() ([]PeopleBurndown, error) {
	burndownData, ok := r.data["Burndown"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing Burndown data in YAML")
	}
	peopleData, ok := burndownData["people"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing people data in Burndown")
	}

	var peopleBurndowns []PeopleBurndown
	for person, matrixData := range peopleData {
		matrix := parseBurndownMatrix(matrixData.(string))
		peopleBurndowns = append(peopleBurndowns, PeopleBurndown{
			Person: person,
			Matrix: matrix,
		})
	}
	return peopleBurndowns, nil
}

func (r *YamlReader) GetOwnershipBurndown() ([]string, map[string][][]int, error) {
	burndownData, ok := r.data["Burndown"].(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("missing Burndown data in YAML")
	}
	peopleSequence, ok := burndownData["people_sequence"].([]string)
	if !ok {
		return nil, nil, fmt.Errorf("missing people_sequence in Burndown")
	}

	peopleData, ok := burndownData["people"].(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("missing people data in Burndown")
	}

	ownership := make(map[string][][]int)
	for person, matrixData := range peopleData {
		matrix := parseBurndownMatrix(matrixData.(string))
		ownership[person] = matrix
	}

	return peopleSequence, ownership, nil
}

func (r *YamlReader) GetPeopleInteraction() ([]string, [][]int, error) {
	burndownData, ok := r.data["Burndown"].(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("missing Burndown data in YAML")
	}
	peopleSequence, ok := burndownData["people_sequence"].([]string)
	if !ok {
		return nil, nil, fmt.Errorf("missing people_sequence in Burndown")
	}
	interactionData, ok := burndownData["people_interaction"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("missing people_interaction data")
	}

	matrix := parseBurndownMatrix(interactionData)
	return peopleSequence, matrix, nil
}

func (r *YamlReader) GetFileCooccurrence() ([]string, [][]int, error) {
	couplesData, ok := r.data["Couples"].(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("missing Couples data in YAML")
	}
	fileIndex, ok := couplesData["file_couples_index"].([]string)
	if !ok {
		return nil, nil, fmt.Errorf("missing file_couples_index in Couples")
	}
	matrixData, ok := couplesData["file_couples_matrix"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("missing file_couples_matrix in Couples")
	}

	matrix := parseBurndownMatrix(matrixData)
	return fileIndex, matrix, nil
}

func (r *YamlReader) GetPeopleCooccurrence() ([]string, [][]int, error) {
	couplesData, ok := r.data["Couples"].(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("missing Couples data in YAML")
	}
	peopleIndex, ok := couplesData["people_couples_index"].([]string)
	if !ok {
		return nil, nil, fmt.Errorf("missing people_couples_index in Couples")
	}
	matrixData, ok := couplesData["people_couples_matrix"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("missing people_couples_matrix in Couples")
	}

	matrix := parseBurndownMatrix(matrixData)
	return peopleIndex, matrix, nil
}

func (r *YamlReader) GetShotnessCooccurrence() ([]string, [][]int, error) {
	shotnessData, ok := r.data["Shotness"].([]map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("missing Shotness data in YAML")
	}

	var names []string
	var matrix [][]int
	for _, record := range shotnessData {
		name := fmt.Sprintf("%s:%s:%s", record["type"], record["name"], record["file"])
		names = append(names, name)
		row := make([]int, len(record["counters"].(map[string]int)))
		for _, value := range record["counters"].(map[string]int) {
			row = append(row, value)
		}
		matrix = append(matrix, row)
	}
	return names, matrix, nil
}

func (r *YamlReader) GetShotnessStats() ([][]int, error) {
	shotnessData, ok := r.data["Shotness"].([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing Shotness data in YAML")
	}

	var stats [][]int
	for _, record := range shotnessData {
		row := make([]int, len(record["counters"].(map[string]int)))
		for _, value := range record["counters"].(map[string]int) {
			row = append(row, value)
		}
		stats = append(stats, row)
	}
	return stats, nil
}

func (r *YamlReader) GetDeveloperStats() ([]DeveloperStat, error) {
	devsData, ok := r.data["Devs"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing Devs data in YAML")
	}
	devIndex, ok := devsData["dev_index"].([]string)
	if !ok {
		return nil, fmt.Errorf("missing dev_index in Devs")
	}
	ticks, ok := devsData["ticks"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing ticks in Devs")
	}

	var devStats []DeveloperStat
	for _, tick := range ticks {
		// Assert the tick as a map
		tickMap, ok := tick.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid tick format")
		}

		// Get the "devs" map from the tick
		devs, ok := tickMap["devs"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("missing 'devs' in tick")
		}

		// Iterate over devs map
		for id, stats := range devs {
			// Cast stats to a map
			statsMap, ok := stats.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid stats format for developer %s", id)
			}

			// Parse developer stats
			commits, _ := statsMap["commits"].(int)       // Safely cast commits
			linesAdded, _ := statsMap["added"].(int)      // Safely cast linesAdded
			linesRemoved, _ := statsMap["removed"].(int)  // Safely cast linesRemoved
			linesModified, _ := statsMap["changed"].(int) // Safely cast linesModified

			index, err := strconv.Atoi(id)
			if err != nil {
				return nil, fmt.Errorf("invalid developer index: %v", err)
			}

			// Append developer stats
			devStats = append(devStats, DeveloperStat{
				Name:          devIndex[index],
				Commits:       commits,
				LinesAdded:    linesAdded,
				LinesRemoved:  linesRemoved,
				LinesModified: linesModified,
			})
		}
	}

	return devStats, nil
}

func (r *YamlReader) GetLanguageStats() ([]LanguageStat, error) {
	// Stub: Language stats data is typically not present in YAML files.
	return nil, fmt.Errorf("language stats not implemented for YAML")
}

func (r *YamlReader) GetRuntimeStats() (map[string]float64, error) {
	// Stub: Runtime stats are typically not present in YAML files.
	return nil, fmt.Errorf("runtime stats not implemented for YAML")
}

// Helper function to parse burndown matrices
func parseBurndownMatrix(data string) [][]int {
	lines := strings.Split(data, "\n")
	var matrix [][]int
	for _, line := range lines {
		if line == "" {
			continue
		}
		numbers := strings.Fields(line)
		var row []int
		for _, num := range numbers {
			val, err := strconv.Atoi(num)
			if err != nil {
				continue
			}
			row = append(row, val)
		}
		matrix = append(matrix, row)
	}
	return matrix
}
