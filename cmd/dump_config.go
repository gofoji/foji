package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"

	"github.com/gofoji/foji/cfg"
)

var dumpConfigCmd = &cobra.Command{
	Use:     "dumpConfig",
	Aliases: []string{"dump"},
	Short:   "Dump the config",
	Long:    `Parses the config.  Writes yaml to stdout.  Used to validate syntax`,
	Run:     dumpConfig,
}

var dumpConfigFlags struct {
	includeDefaults bool
	weld            string
}

func registerDumpConfigFlags() {
	dumpConfigCmd.PersistentFlags().BoolVarP(&dumpConfigFlags.includeDefaults, "includeDefaults", "d", false,
		"Include evaluated Fōji defaults in the dump")
	dumpConfigCmd.PersistentFlags().StringVarP(&dumpConfigFlags.weld, "weld", "w", "",
		"limits output to only the specified weld config (includes defaults")
}

func dumpConfig(_ *cobra.Command, _ []string) {
	l := rootFlags.logger()

	config, err := cfg.Load(rootFlags.cfgFile, dumpConfigFlags.includeDefaults || dumpConfigFlags.weld != "")
	if err != nil {
		l.Fatal().Err(err).Msg("Loading Config")
	}

	var out any = config

	if dumpConfigFlags.weld != "" {
		targets, err := config.Processes.Target([]string{dumpConfigFlags.weld})
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
