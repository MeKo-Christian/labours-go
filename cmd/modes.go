package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"labours-go/internal/modes"
	"labours-go/internal/readers"
)

// Map of mode names to their handlers
var modeHandlers = map[string]func(reader readers.Reader, output string, startTime, endTime *time.Time) error{
	"burndown-project":  burndownProject,
	"burndown-file":     burndownFile,
	"burndown-person":   burndownPerson,
	"overwrites-matrix": overwritesMatrix,
	"ownership":         ownershipBurndown,
	"couples-files":     couplesFiles,
	"couples-people":    couplesPeople,
	"couples-shotness":  couplesShotness,
	"shotness":          shotness,
	"devs":              devs,
	"devs-efforts":      devsEfforts,
	"old-vs-new":        oldVsNew,
	"languages":         languages,
	"devs-parallel":     devsParallel,
	"run-times":         runTimes,
	"sentiment":         sentiment,
}

func executeModes(modes []string, reader readers.Reader, output string, startTime, endTime *time.Time) {
	for _, mode := range modes {
		fmt.Printf("Running mode: %s\n", mode)
		if modeFunc, ok := modeHandlers[mode]; ok {
			if err := modeFunc(reader, output, startTime, endTime); err != nil {
				fmt.Printf("Error in mode %s: %v\n", mode, err)
			}
		} else {
			fmt.Printf("Unknown mode: %s\n", mode)
		}
	}
}

func burndownProject(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	relative := viper.GetBool("relative")
	resample := viper.GetString("resample")
	return modes.BurndownProject(reader, output, relative, startTime, endTime, resample)
}

func burndownFile(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	relative := viper.GetBool("relative")
	resample := viper.GetString("resample")
	return modes.BurndownFile(reader, output, relative, startTime, endTime, resample)
}

func burndownPerson(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	relative := viper.GetBool("relative")
	resample := viper.GetString("resample")
	return modes.BurndownPerson(reader, output, relative, startTime, endTime, resample)
}

func overwritesMatrix(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	return modes.OverwritesMatrix(reader, output)
}

func ownershipBurndown(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	return modes.OwnershipBurndown(reader, output)
}

func couplesFiles(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	fmt.Println("Executing couples-files...")
	// Add logic for couples files
	return nil
}

func couplesPeople(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	fmt.Println("Executing couples-people...")
	// Add logic for couples people
	return nil
}

func couplesShotness(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	fmt.Println("Executing couples-shotness...")
	// Add logic for couples shotness
	return nil
}

func shotness(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	fmt.Println("Executing shotness...")
	// Add logic for shotness
	return nil
}

func devs(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	fmt.Println("Executing devs...")
	// Add logic for devs
	return nil
}

func devsEfforts(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	fmt.Println("Executing devs-efforts...")
	// Add logic for devs efforts
	return nil
}

func oldVsNew(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	fmt.Println("Executing old-vs-new...")
	// Add logic for old vs new
	return nil
}

func languages(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	fmt.Println("Executing languages...")
	// Add logic for languages
	return nil
}

func devsParallel(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	fmt.Println("Executing devs-parallel...")
	// Add logic for devs parallel
	return nil
}

func runTimes(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	fmt.Println("Executing run-times...")
	// Add logic for run times
	return nil
}
