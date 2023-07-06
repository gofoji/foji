package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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

const (
	permRWXUser = 0o700
	permRWUser  = 0o600
)

func WriteToFile(source []byte, file string) error {
	if err := os.MkdirAll(filepath.Dir(file), permRWXUser); err != nil {
		return fmt.Errorf("create output directory:%w", err)
	}

	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, permRWUser)
	if err != nil {
		return fmt.Errorf("open file:%w", err)
	}

	_, err = f.Write(source)

	if closeErr := f.Close(); closeErr != nil {
		return fmt.Errorf("closing file:%w", closeErr)
	}

	if err != nil {
		return fmt.Errorf("writing file:%w", err)
	}

	return nil
}
