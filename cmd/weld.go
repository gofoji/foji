package cmd

import (
	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/welder"
	"github.com/spf13/cobra"
)

var weldCmd = &cobra.Command{
	Use:     "weld [list of processes]",
	Aliases: []string{"w"},
	Short:   "Runs the list of processes.",
	Long:    ``,
	Args:    cobra.MinimumNArgs(1),
	Run:     weld,
}

func weld(_ *cobra.Command, args []string) {
	l := getLogger(quiet, trace, verbose)

	c, err := cfg.Load(cfgFile, true)
	if err != nil {
		l.WithError(err).Fatal("loading config")
	}

	targets, err := c.Processes.Target(args)
	if err != nil {
		l.WithError(err).Fatal("getting targets")
	}

	err = welder.New(l, c, targets).Run(simulate)
	if err != nil {
		l.WithError(err).Fatal("welding")
	}
}
