package readers

import "io"

type Reader interface {
	Read(file io.Reader) error
	GetName() string
	GetHeader() (int64, int64)
	GetProjectBurndown() (string, [][]int)
	GetFilesBurndown() ([]FileBurndown, error)
	GetPeopleBurndown() ([]PeopleBurndown, error)
	GetOwnershipBurndown() ([]string, map[string][][]int, error)
	GetPeopleInteraction() ([]string, [][]int, error)
	GetFileCooccurrence() ([]string, [][]int, error)
	GetPeopleCooccurrence() ([]string, [][]int, error)
	GetShotnessCooccurrence() ([]string, [][]int, error)
	GetShotnessRecords() ([]ShotnessRecord, error)
	GetDeveloperStats() ([]DeveloperStat, error)
	GetLanguageStats() ([]LanguageStat, error)
	GetRuntimeStats() (map[string]float64, error)
}

type FileBurndown struct {
	Filename string
	Matrix   [][]int
}

type PeopleBurndown struct {
	Person string
	Matrix [][]int
}

type DeveloperStat struct {
	Name          string
	Commits       int
	LinesAdded    int
	LinesRemoved  int
	LinesModified int
	FilesTouched  int
	Languages     map[string]int
}

type LanguageStat struct {
	Language string
	Lines    int
}

type ShotnessRecord struct {
	Type     string            // Type of structural unit (e.g., "function", "class")
	Name     string            // Name of the structural unit
	File     string            // File path containing the unit
	Counters map[int32]int32   // Time-based modification counters
}
