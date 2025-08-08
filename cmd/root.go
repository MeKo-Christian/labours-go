package cmd

import (
	"fmt"
	"os"
	"strings"

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
	rootCmd.PersistentFlags().StringP("output", "o", "", "Path to output file/directory. JSON extension saves data instead of image")
	rootCmd.PersistentFlags().StringP("input", "i", "-", "Path to input file")
	rootCmd.PersistentFlags().StringP("input-format", "f", "auto", "Input format")
	rootCmd.PersistentFlags().Int("font-size", 12, "Size of labels and legend")
	rootCmd.PersistentFlags().String("style", "ggplot", "Plot style to use")
	rootCmd.PersistentFlags().String("backend", "", "Matplotlib backend")
	rootCmd.PersistentFlags().String("background", "white", "Plot's general color scheme")
	rootCmd.PersistentFlags().String("size", "", "Axes' size in inches, e.g. \"12,9\"")
	rootCmd.PersistentFlags().Bool("relative", false, "Occupy 100% height for every measurement")
	rootCmd.PersistentFlags().String("tmpdir", "", "Temporary directory for intermediate files")
	rootCmd.PersistentFlags().StringSliceP("modes", "m", []string{}, "What to plot, can be repeated")
	rootCmd.PersistentFlags().String("resample", "year", "Resample time series method")
	rootCmd.PersistentFlags().String("start-date", "", "Start date for time-based plots")
	rootCmd.PersistentFlags().String("end-date", "", "End date for time-based plots")
	rootCmd.PersistentFlags().Bool("disable-projector", false, "Do not run Tensorflow Projector")
	rootCmd.PersistentFlags().Int("max-people", 20, "Maximum developers in matrix and people plots")
	rootCmd.PersistentFlags().Bool("order-ownership-by-time", false, "Sort developers in the ownership plot by their first appearance in the history.")
	rootCmd.PersistentFlags().Bool("sentiment", false, "Include sentiment analysis in the output (Python compatibility)")

	// Progress and output control flags
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Disable progress bars and reduce output")
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output with detailed progress information")

	// Theme-related flags
	rootCmd.PersistentFlags().String("theme", "default", "Theme to use for visualization (default, dark, minimal, vibrant)")
	rootCmd.PersistentFlags().Bool("list-themes", false, "List all available themes and exit")
	rootCmd.PersistentFlags().String("export-theme", "", "Export a built-in theme to file for customization")
	rootCmd.PersistentFlags().String("load-theme", "", "Load custom theme from file")

	// Hercules integration flags
	rootCmd.PersistentFlags().String("hercules", "", "Path to hercules binary (empty for auto-detection)")
	rootCmd.PersistentFlags().String("from-repo", "", "Analyze git repository directly using hercules")
	rootCmd.PersistentFlags().String("hercules-flags", "", "Additional flags to pass to hercules")
}

func bindFlagsToViper() {
	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		fmt.Printf("Error binding flags: %v\n", err)
		os.Exit(1)
	}
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

	// Set the selected theme, with style-to-theme mapping
	themeName := viper.GetString("theme")
	styleName := viper.GetString("style")

	// Map matplotlib styles to themes for compatibility
	if styleName != "ggplot" && styleName != "" {
		mappedTheme := mapStyleToTheme(styleName)
		if mappedTheme != "" {
			themeName = mappedTheme
			if !viper.GetBool("quiet") {
				fmt.Printf("Mapping matplotlib style '%s' to theme '%s'\n", styleName, mappedTheme)
			}
		}
	}

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

// mapStyleToTheme maps matplotlib style names to labours-go theme names
func mapStyleToTheme(style string) string {
	styleToTheme := map[string]string{
		// Core matplotlib built-in styles
		"default":              "default", // matplotlib default
		"classic":              "default", // classic matplotlib -> default
		"ggplot":               "default", // ggplot is our default
		"dark_background":      "dark",    // dark background -> dark theme
		"grayscale":            "minimal", // grayscale -> minimal
		"bmh":                  "vibrant", // Bayesian Methods for Hackers -> vibrant
		"fivethirtyeight":      "vibrant", // FiveThirtyEight -> vibrant
		"fast":                 "default", // fast style -> default

		// Seaborn styles (original and v0.8+ variants)
		"seaborn":              "minimal", // seaborn-like -> minimal
		"seaborn-v0_8":         "minimal", // newer seaborn -> minimal
		"seaborn-bright":       "vibrant", // seaborn bright -> vibrant
		"seaborn-colorblind":   "default", // seaborn colorblind -> default
		"seaborn-dark":         "dark",    // seaborn dark -> dark
		"seaborn-darkgrid":     "dark",    // seaborn dark grid -> dark
		"seaborn-pastel":       "minimal", // seaborn pastel -> minimal
		"seaborn-white":        "minimal", // seaborn white -> minimal
		"seaborn-whitegrid":    "default", // seaborn white grid -> default
		"seaborn-paper":        "minimal", // seaborn paper -> minimal
		"seaborn-poster":       "vibrant", // seaborn poster -> vibrant
		"seaborn-talk":         "default", // seaborn talk -> default
		"seaborn-notebook":     "default", // seaborn notebook -> default
		"seaborn-muted":        "minimal", // seaborn muted -> minimal
		"seaborn-deep":         "dark",    // seaborn deep -> dark
		"seaborn-ticks":        "default", // seaborn ticks -> default

		// Tableau styles
		"tableau-colorblind10": "default", // tableau -> default
		"tab10":                "default", // tableau 10 colors -> default
		"tab20":                "vibrant", // tableau 20 colors -> vibrant
		"tab20b":               "vibrant", // tableau 20b -> vibrant
		"tab20c":               "minimal", // tableau 20c -> minimal

		// Solarized styles
		"Solarize_Light2":      "minimal", // Solarized light -> minimal
		"solarized":            "minimal", // general solarized -> minimal
		"solarized-light":      "minimal", // solarized light -> minimal
		"solarized-dark":       "dark",    // solarized dark -> dark

		// Additional matplotlib styles
		"cyberpunk":            "dark",    // cyberpunk style -> dark
		"science":              "minimal", // science style -> minimal
		"ieee":                 "minimal", // IEEE format -> minimal
		"nature":               "default", // nature format -> default
		"grid":                 "default", // with grid -> default
		"no-latex":             "default", // no LaTeX -> default

		// Common style variants and aliases (case-insensitive)
		"dark":         "dark",
		"light":        "default",
		"minimal":      "minimal",
		"vibrant":      "vibrant",
		"colorful":     "vibrant",
		"monochrome":   "minimal",
		"black":        "dark",
		"white":        "minimal",
		"bright":       "vibrant",
		"muted":        "minimal",
		"pastel":       "minimal",
		"deep":         "dark",
		"paper":        "minimal",
		"poster":       "vibrant",
		"talk":         "default",
		"notebook":     "default",
		"whitegrid":    "default",
		"darkgrid":     "dark",
		"ticks":        "default",

		// Color scheme aliases
		"blues":        "minimal",
		"greens":       "minimal",
		"greys":        "minimal",
		"oranges":      "vibrant",
		"purples":      "vibrant",
		"reds":         "vibrant",
	}

	return styleToTheme[strings.ToLower(style)]
}
