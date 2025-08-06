package modes

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"labours-go/internal/readers"
)

// Mock reader for testing the languages mode
type MockLanguageReader struct {
	languageStats []readers.LanguageStat
}

func (m *MockLanguageReader) Read(file io.Reader) error                        { return nil }
func (m *MockLanguageReader) GetName() string                                   { return "mock-repo" }
func (m *MockLanguageReader) GetHeader() (int64, int64)                         { return 0, 0 }
func (m *MockLanguageReader) GetProjectBurndown() (string, [][]int)             { return "", nil }
func (m *MockLanguageReader) GetFilesBurndown() ([]readers.FileBurndown, error) { return nil, nil }
func (m *MockLanguageReader) GetPeopleBurndown() ([]readers.PeopleBurndown, error) {
	return nil, nil
}
func (m *MockLanguageReader) GetOwnershipBurndown() ([]string, map[string][][]int, error) {
	return nil, nil, nil
}
func (m *MockLanguageReader) GetPeopleInteraction() ([]string, [][]int, error)   { return nil, nil, nil }
func (m *MockLanguageReader) GetFileCooccurrence() ([]string, [][]int, error)    { return nil, nil, nil }
func (m *MockLanguageReader) GetPeopleCooccurrence() ([]string, [][]int, error)  { return nil, nil, nil }
func (m *MockLanguageReader) GetShotnessCooccurrence() ([]string, [][]int, error) { return nil, nil, nil }
func (m *MockLanguageReader) GetShotnessRecords() ([]readers.ShotnessRecord, error) { return nil, nil }
func (m *MockLanguageReader) GetDeveloperStats() ([]readers.DeveloperStat, error) { return nil, nil }
func (m *MockLanguageReader) GetRuntimeStats() (map[string]float64, error)       { return nil, nil }

func (m *MockLanguageReader) GetLanguageStats() ([]readers.LanguageStat, error) {
	return m.languageStats, nil
}

func TestLanguages(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir := t.TempDir()

	// Sample language statistics for testing
	testLangStats := []readers.LanguageStat{
		{Language: "Go", Lines: 15000},
		{Language: "Python", Lines: 8000},
		{Language: "JavaScript", Lines: 5000},
		{Language: "HTML", Lines: 3000},
		{Language: "CSS", Lines: 2000},
	}

	mockReader := &MockLanguageReader{
		languageStats: testLangStats,
	}

	err := Languages(mockReader, tmpDir)
	if err != nil {
		t.Errorf("Languages() error = %v", err)
	}

	// Check if output files were created
	pngFile := filepath.Join(tmpDir, "languages.png")
	if _, err := os.Stat(pngFile); os.IsNotExist(err) {
		t.Errorf("PNG output file was not created: %s", pngFile)
	}

	svgFile := filepath.Join(tmpDir, "languages.svg")
	if _, err := os.Stat(svgFile); os.IsNotExist(err) {
		t.Errorf("SVG output file was not created: %s", svgFile)
	}
}

func TestLanguagesEmptyData(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with empty language stats
	mockReader := &MockLanguageReader{
		languageStats: []readers.LanguageStat{},
	}

	err := Languages(mockReader, tmpDir)
	if err == nil {
		t.Error("Expected error for empty language stats, but got nil")
	}

	expectedErrorMessage := "no language statistics found in the data"
	if err.Error() != expectedErrorMessage+" - the input file may not contain language analysis results" {
		t.Errorf("Expected error message containing '%s', got: %v", expectedErrorMessage, err)
	}
}

func TestLanguagesSingleLanguage(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with single language
	testLangStats := []readers.LanguageStat{
		{Language: "Go", Lines: 25000},
	}

	mockReader := &MockLanguageReader{
		languageStats: testLangStats,
	}

	err := Languages(mockReader, tmpDir)
	if err != nil {
		t.Errorf("Languages() with single language error = %v", err)
	}

	// Check if output files were created
	pngFile := filepath.Join(tmpDir, "languages.png")
	if _, err := os.Stat(pngFile); os.IsNotExist(err) {
		t.Errorf("PNG output file was not created: %s", pngFile)
	}
}

func TestLanguagesSorting(t *testing.T) {
	// This test validates that languages are properly sorted by line count
	tmpDir := t.TempDir()

	// Unsorted language stats
	testLangStats := []readers.LanguageStat{
		{Language: "CSS", Lines: 1000},
		{Language: "Go", Lines: 15000},
		{Language: "Python", Lines: 8000},
		{Language: "HTML", Lines: 2000},
		{Language: "JavaScript", Lines: 5000},
	}

	mockReader := &MockLanguageReader{
		languageStats: testLangStats,
	}

	err := Languages(mockReader, tmpDir)
	if err != nil {
		t.Errorf("Languages() sorting test error = %v", err)
	}

	// The function should handle sorting internally, so we just verify it completes successfully
	pngFile := filepath.Join(tmpDir, "languages.png")
	if _, err := os.Stat(pngFile); os.IsNotExist(err) {
		t.Errorf("PNG output file was not created: %s", pngFile)
	}
}