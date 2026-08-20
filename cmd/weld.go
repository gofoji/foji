package cmd

import (
	"github.com/spf13/cobra"

	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/welder"
)

var weldCmd = &cobra.Command{
	Use:     "weld [list of processes]",
	Aliases: []string{"w"},
	Short:   "Runs the list of processes.",
	Long:    ``,
	Args:    cobra.MinimumNArgs(1),
	Run:     weld,
}

var weldFlags struct {
	simulate bool
}

func registerWeldFlags() {
	weldCmd.PersistentFlags().BoolVarP(&weldFlags.simulate, "simulate", "s", false,
		"simulates the processing, only displays files that would be generated")
}

func weld(_ *cobra.Command, args []string) {
	l := rootFlags.logger()

	c, err := cfg.Load(rootFlags.cfgFile, true)
	if err != nil {
		l.Fatal().Err(err).Msg("Unable to load config")
	}

	targets, err := c.Processes.Target(args)
	if err != nil {
		l.Fatal().Err(err).Msg("Getting targets")
	}

	err = welder.New(l, c, targets).Run(weldFlags.simulate)
	if err != nil {
		l.Fatal().Err(err).Msg("Welding")
	}
}
