package cmd

import (
	"fmt"

	"github.com/gofoji/foji/cfg"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Args:    cobra.MaximumNArgs(0),
	Short:   "List all available processes",
	Run:     list,
}

func list(_ *cobra.Command, args []string) {
	l := getLogger(quiet, trace, verbose)

	c, err := cfg.Load(cfgFile, true)
	if err != nil {
		l.WithError(err).Fatal("Failed to load config")
	}

	fmt.Printf("Available Processes: %v\n", c.Processes.Keys()) // nolint
}
