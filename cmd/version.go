package cmd

import (
	"fmt"

	"github.com/gofoji/foji/cfg"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Display version",
	Run:     version,
}

func version(_ *cobra.Command, _ []string) {
	fmt.Println(cfg.Version())
}
