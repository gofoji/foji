package cmd

import (
	"os"

	"github.com/gofoji/foji/cfg"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var dumpConfigCmd = &cobra.Command{
	Use:     "dumpConfig",
	Aliases: []string{"dump"},
	Short:   "Dump the config",
	Long:    `Parses the config.  Writes yaml to stdout.  Used to validate syntax`,
	Run:     dumpConfig,
}

func dumpConfig(_ *cobra.Command, _ []string) {
	l := getLogger(quiet, trace, verbose)

	c, err := cfg.Load(cfgFile, includeDefaults)
	if err != nil {
		l.Fatal().Err(err).Msg("Loading Config")
	}

	err = yaml.NewEncoder(os.Stdout).Encode(c)
	if err != nil {
		l.Fatal().Err(err).Msg("Writing Yaml")
	}
}
