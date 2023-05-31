package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/gofoji/foji/cfg"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Args:    cobra.MaximumNArgs(0),
	Short:   "List all available processes",
	Run:     list,
}

func list(_ *cobra.Command, _ []string) {
	l := getLogger(quiet, trace, verbose)

	config, err := cfg.Load(cfgFile, true)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to load config")
	}

	fmt.Printf("Available Processes: %v\n", config.Processes.Keys()) //nolint
}
