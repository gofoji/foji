package cmd

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/gofoji/foji/foji"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"i"},
	Short:   "Writes a sample config file to ./foji.yaml",
	Run:     initConfig,
}

var initFlags struct {
	overwrite bool
}

func registerInitFlags() {
	initCmd.PersistentFlags().BoolVarP(&initFlags.overwrite, "overwrite", "y", false,
		"force overwrite an existing output file")
}

func initConfig(_ *cobra.Command, _ []string) {
	l := log.With().Str("cfgFile", rootFlags.cfgFile).Logger()

	info, err := os.Stat(rootFlags.cfgFile)
	if os.IsNotExist(err) {
		writeConfig()

		return
	}

	if err != nil {
		l.Fatal().Err(err).Msg("unable to access cfgFile")
	}

	if info.IsDir() {
		l.Fatal().Msg("cfgFile is a directory")
	}

	if !initFlags.overwrite {
		l.Fatal().Msg("cfgFile exists, specify `overwrite` to replace")
	}

	l.Info().Msg("overwriting")
	writeConfig()
}

func writeConfig() {
	l := log.With().Str("cfgFile", rootFlags.cfgFile).Logger()

	err := WriteToFile(foji.InitConfig, rootFlags.cfgFile)
	if err != nil {
		l.Fatal().Err(err).Msg("saving file")
	}

	l.Info().Msg("wrote sample foji config file")
}
