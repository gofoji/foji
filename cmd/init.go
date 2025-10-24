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

func initConfig(_ *cobra.Command, _ []string) {
	l := log.With().Str("cfgFile", cfgFile).Logger()

	info, err := os.Stat(cfgFile)
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

	if !overwrite {
		l.Fatal().Msg("cfgFile exists, specify `overwrite` to replace")
	}

	l.Info().Msg("overwriting")
	writeConfig()
}

func writeConfig() {
	l := log.With().Str("cfgFile", cfgFile).Logger()

	err := WriteToFile(foji.InitConfig, cfgFile)
	if err != nil {
		l.Fatal().Err(err).Msg("saving file")
	}

	l.Info().Msg("wrote sample foji config file")
}
