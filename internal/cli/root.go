package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	cacheDir string
	debug    bool
)

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "xrv",
		Short: "XRV - Exchange Rate Visualizer",
		Long: `XRV is a CLI application for visualizing historical exchange rates.

It provides interactive terminal visualizations, comprehensive statistics,
and support for long historical time ranges (back to 1999).`,
		Version: "0.1.0",
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.xrv/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&cacheDir, "cache-dir", "", "cache directory (default is $HOME/.xrv/cache)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")

	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("cache-dir", rootCmd.PersistentFlags().Lookup("cache-dir"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	rootCmd.AddCommand(NewVisualizeCommand())

	return rootCmd
}

func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
