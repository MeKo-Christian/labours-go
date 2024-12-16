package modes

import (
	"fmt"
	"labours-go/internal/readers"
	"time"
)

// BurndownProject generates a burndown chart for the entire project.
func BurndownProject(reader readers.Reader, output string, relative bool, startTime, endTime *time.Time, resample string) error {
	repoName, burndownMatrix := reader.GetProjectBurndown()
	if len(burndownMatrix) == 0 {
		return fmt.Errorf("no burndown data available for project")
	}

	// Generate plot
	return generateBurndownPlot(repoName, burndownMatrix, output, relative, startTime, endTime, resample)
}
