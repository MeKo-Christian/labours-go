package modes

import (
	"fmt"
	"time"

	"labours-go/internal/readers"
)

// BurndownPerson generates burndown charts for individual people/developers.
func BurndownPerson(reader readers.Reader, output string, relative bool, startDate, endDate *time.Time, resample string) error {
	peopleBurndowns, err := reader.GetPeopleBurndown()
	if err != nil {
		return fmt.Errorf("failed to get people burndown data: %v", err)
	}

	// Generate a chart for each person
	for _, person := range peopleBurndowns {
		outputFile := fmt.Sprintf("%s_%s.png", output, person.Person)
		if err := generateBurndownPlot(person.Person, person.Matrix, outputFile, relative, startDate, endDate, resample); err != nil {
			return fmt.Errorf("failed to generate burndown for person %s: %v", person.Person, err)
		}
	}

	return nil
}
