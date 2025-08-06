package cmd

import (
	"fmt"
	"os"

	"labours-go/internal/graphics"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "labours",
	Short: "Labours CLI for analyzing git repository data",
	Long:  "Labours CLI for analyzing git repository data, visualizing trends, and generating reports.",
	Run:   runLaboursCommand,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	initializeFlags()
	bindFlagsToViper()
}

func initializeFlags() {
	rootCmd.PersistentFlags().StringP("output", "o", "", "Path to the output file/directory")
	rootCmd.PersistentFlags().StringP("input", "i", "-", "Path to the input file (- for stdin)")
	rootCmd.PersistentFlags().StringP("input-format", "f", "auto", "Input format (yaml, pb, auto)")
	rootCmd.PersistentFlags().Int("font-size", 12, "Size of the labels and legend")
	rootCmd.PersistentFlags().String("style", "ggplot", "Plot style to use")
	rootCmd.PersistentFlags().String("backend", "", "Matplotlib backend to use")
	rootCmd.PersistentFlags().String("background", "white", "Plot's general color scheme")
	rootCmd.PersistentFlags().String("size", "", "Axes' size in inches, e.g., '12,9'")
	rootCmd.PersistentFlags().Bool("relative", false, "Occupy 100% height for every measurement")
	rootCmd.PersistentFlags().String("tmpdir", "/tmp", "Temporary directory for intermediate files")
	rootCmd.PersistentFlags().StringSliceP("modes", "m", []string{}, "Modes to run (can be repeated)")
	rootCmd.PersistentFlags().String("resample", "year", "Resample interval for time series")
	rootCmd.PersistentFlags().String("start-date", "", "Start date for time-based plots")
	rootCmd.PersistentFlags().String("end-date", "", "End date for time-based plots")
	rootCmd.PersistentFlags().Bool("disable-projector", false, "Disable TensorFlow projector on couples")
	rootCmd.PersistentFlags().Int("max-people", 20, "Maximum number of developers in overwrites matrix and people plots.")
	rootCmd.PersistentFlags().Bool("order-ownership-by-time", false, "Sort developers in the ownership plot by their first appearance in the history.")
	rootCmd.PersistentFlags().Bool("sentiment", false, "Include sentiment analysis in the output (Python compatibility)")
}

func bindFlagsToViper() {
	viper.BindPFlags(rootCmd.PersistentFlags())
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.labours-go")

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println("No configuration file found, using defaults.")
	}

	// Load user themes from standard directories
	if err := graphics.LoadUserThemes(); err != nil {
		fmt.Printf("Warning: failed to load user themes: %v\n", err)
	}
}

func runLaboursCommand(cmd *cobra.Command, args []string) {
	// Handle theme-specific commands first
	if viper.GetBool("list-themes") {
		listThemes()
		return
	}

	if exportTheme := viper.GetString("export-theme"); exportTheme != "" {
		handleExportTheme(exportTheme)
		return
	}

	// Load custom theme if specified
	if loadTheme := viper.GetString("load-theme"); loadTheme != "" {
		if err := graphics.GlobalThemeManager.LoadThemeFromFile(loadTheme); err != nil {
			fmt.Printf("Failed to load custom theme: %v\n", err)
			os.Exit(1)
		}
	}

	// Set the selected theme
	themeName := viper.GetString("theme")
	if err := graphics.SetTheme(themeName); err != nil {
		fmt.Printf("Failed to set theme '%s': %v\n", themeName, err)
		fmt.Printf("Available themes: %v\n", graphics.ListThemes())
		os.Exit(1)
	}

	// Handle hercules integration if --from-repo is specified
	if repoPath := viper.GetString("from-repo"); repoPath != "" {
		handleHerculesIntegration(repoPath)
		return
	}

	input, inputFormat := viper.GetString("input"), viper.GetString("input-format")
	startDate, endDate := parseDates()
	validateDateRange(startDate, endDate)

	reader := detectAndReadInput(input, inputFormat)
	modes := resolveModes()

	// Handle Python compatibility: if --sentiment flag is set, add sentiment mode
	if viper.GetBool("sentiment") {
		modes = append(modes, "sentiment")
		fmt.Println("Added sentiment analysis mode (--sentiment flag)")
	}

	executeModes(modes, reader, viper.GetString("output"), startDate, endDate)
}

func listThemes() {
	fmt.Println("Available themes:")
	for _, theme := range graphics.ListThemes() {
		fmt.Printf("  - %s\n", theme)
	}
}

func handleExportTheme(themeName string) {
	outputPath := fmt.Sprintf("%s-theme.yaml", themeName)
	if err := graphics.GlobalThemeManager.ExportTheme(themeName, outputPath); err != nil {
		fmt.Printf("Failed to export theme '%s': %v\n", themeName, err)
		os.Exit(1)
	}
	fmt.Printf("Theme '%s' exported to %s\n", themeName, outputPath)
}

func handleHerculesIntegration(repoPath string) {
	// Auto-detect hercules binary
	herculesPath := viper.GetString("hercules")
	if herculesPath == "" {
		// Try common locations
		candidates := []string{
			"hercules",
			"./hercules",
			"/usr/local/bin/hercules",
			"/home/christian/Code/hercules/hercules",
		}

		for _, candidate := range candidates {
			if isExecutable(candidate) {
				herculesPath = candidate
				break
			}
		}

		if herculesPath == "" {
			fmt.Println("Error: hercules binary not found. Please install hercules or specify path with --hercules flag")
			os.Exit(1)
		}
	}

	fmt.Printf("Using hercules: %s\n", herculesPath)
	fmt.Printf("Analyzing repository: %s\n", repoPath)

	// Check if repository exists and is a git repo
	if !isGitRepository(repoPath) {
		fmt.Printf("Error: %s is not a git repository\n", repoPath)
		os.Exit(1)
	}

	modes := resolveModes()
	if len(modes) == 0 {
		modes = []string{"burndown-project", "devs"} // default modes
	}

	// Map labours-go modes to hercules analysis
	herculesAnalyses := mapModesToHerculesAnalyses(modes)

	for _, analysis := range herculesAnalyses {
		if err := runHerculesAndVisualize(herculesPath, repoPath, analysis); err != nil {
			fmt.Printf("Error running analysis '%s': %v\n", analysis, err)
		}
	}
}
