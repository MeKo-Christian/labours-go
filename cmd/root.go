package cmd

import (
	"fmt"

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
}

func runLaboursCommand(cmd *cobra.Command, args []string) {
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
