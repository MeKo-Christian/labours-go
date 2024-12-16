package cmd

import (
	"fmt"
	"labours-go/internal/readers"
	"os"
	"time"

	"github.com/araddon/dateparse"
	"github.com/spf13/viper"
)

// parseFlexibleDate parses a date string into a time.Time object.
// Returns an error if the date cannot be parsed.
func parseFlexibleDate(dateStr string) (time.Time, error) {
	parsedDate, err := dateparse.ParseAny(dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format: %v", err)
	}
	return parsedDate, nil
}

func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

func parseDates() (startTime *time.Time, endTime *time.Time) {
	if startTimeStr := viper.GetString("start-date"); startTimeStr != "" {
		parsedStartTime, err := parseFlexibleDate(startTimeStr)
		if err != nil {
			fmt.Printf("Error parsing start date: %v\n", err)
			os.Exit(1)
		}
		startTime = &parsedStartTime
	}

	if endTimeStr := viper.GetString("end-date"); endTimeStr != "" {
		parsedEndTime, err := parseFlexibleDate(endTimeStr)
		if err != nil {
			fmt.Printf("Error parsing end date: %v\n", err)
			os.Exit(1)
		}
		endTime = &parsedEndTime
	}

	return startTime, endTime
}

func validateDateRange(startTime, endTime *time.Time) {
	if startTime != nil && endTime != nil && endTime.Before(*startTime) {
		fmt.Println("Error: end date must be after start date")
		os.Exit(1)
	}
}

func detectAndReadInput(input, inputFormat string) readers.Reader {
	reader, err := readers.DetectAndReadInput(input, inputFormat)
	if err != nil {
		fmt.Printf("Error detecting or reading input: %v\n", err)
		os.Exit(1)
	}
	return reader
}

func resolveModes() []string {
	modes := viper.GetStringSlice("modes")
	if len(modes) == 0 {
		fmt.Println("No modes specified. Use --modes to specify what to run.")
		os.Exit(1)
	}

	if contains(modes, "all") {
		modes = []string{
			"burndown-project", "overwrites-matrix", "ownership",
			"couples-files", "couples-people", "couples-shotness",
			"shotness", "devs", "devs-efforts",
		}
	}
	return modes
}
