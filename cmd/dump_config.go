package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/gofoji/foji/cfg"
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

	config, err := cfg.Load(cfgFile, includeDefaults || dumpWeld != "")
	if err != nil {
		l.Fatal().Err(err).Msg("Loading Config")
	}

	var out any = config

	if dumpWeld != "" {
		targets, err := config.Processes.Target([]string{dumpWeld})
		if err != nil {
			l.Fatal().Err(err).Msg("Getting welds")
		}

		if len(targets) != 1 {
			l.Fatal().Int("Found Welds", len(targets)).Msg("Invalid weld count")
		}

		out = targets[0]
	}

	err = yaml.NewEncoder(os.Stdout).Encode(out)
	if err != nil {
		l.Fatal().Err(err).Msg("Writing Yaml")
	}
}
