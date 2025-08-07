package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"labours-go/internal/modes"
	"labours-go/internal/progress"
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
	// Initialize progress tracking for multiple modes
	quiet := viper.GetBool("quiet")
	progEstimator := progress.NewProgressEstimator(!quiet)
	
	if len(modes) > 1 {
		// Start multi-mode progress tracking
		progEstimator.StartMultiOperation(len(modes), "Analysis Modes")
		
		for _, mode := range modes {
			progEstimator.NextOperation(fmt.Sprintf("Running %s", mode))
			
			if !quiet {
				fmt.Printf("Running mode: %s\n", mode)
			}
			
			if modeFunc, ok := modeHandlers[mode]; ok {
				if err := modeFunc(reader, output, startTime, endTime); err != nil {
					fmt.Printf("Error in mode %s: %v\n", mode, err)
				}
			} else {
				fmt.Printf("Unknown mode: %s\n", mode)
			}
		}
		
		progEstimator.FinishMultiOperation()
	} else {
		// Single mode - let the individual mode handle its own progress
		for _, mode := range modes {
			if !quiet {
				fmt.Printf("Running mode: %s\n", mode)
			}
			
			if modeFunc, ok := modeHandlers[mode]; ok {
				if err := modeFunc(reader, output, startTime, endTime); err != nil {
					fmt.Printf("Error in mode %s: %v\n", mode, err)
				}
			} else {
				fmt.Printf("Unknown mode: %s\n", mode)
			}
		}
	}
}

func burndownProject(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	relative := viper.GetBool("relative")
	resample := viper.GetString("resample")
	// Use Python-compatible implementation
	return modes.GenerateBurndownProjectPython(reader, output, relative, resample)
}

func burndownFile(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	relative := viper.GetBool("relative")
	resample := viper.GetString("resample")
	// Use Python-compatible implementation  
	return modes.GenerateBurndownFilePython(reader, output, relative, resample)
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
	return modes.Shotness(reader, output)
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
	resample := viper.GetString("resample")
	return modes.OldVsNew(reader, output, startTime, endTime, resample)
}

func languages(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	return modes.Languages(reader, output)
}

func devsParallel(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	return modes.DevsParallel(reader, output)
}

func runTimes(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	fmt.Println("Executing run-times...")
	// Add logic for run times
	return nil
}

func sentiment(reader readers.Reader, output string, startTime, endTime *time.Time) error {
	return modes.Sentiment(reader, output)
}
