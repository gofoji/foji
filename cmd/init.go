package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofoji/foji/embed"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"i"},
	Short:   "Writes a sample config file to ./foji.yaml",
	Run:     initConfig,
}

func initConfig(_ *cobra.Command, _ []string) {
	info, err := os.Stat(cfgFile)
	if os.IsNotExist(err) {
		writeConfig()

		return
	}

	if err != nil {
		logrus.WithError(err).WithField("cfgFile", cfgFile).Fatal("unable to access cfgFile")
	}

	if info.IsDir() {
		logrus.WithField("cfgFile", cfgFile).Fatal("cfgFile is a directory")
	}

	if overwrite {
		logrus.WithField("cfgFile", cfgFile).Warn("Overwriting")
		writeConfig()
	} else {
		logrus.WithField("cfgFile", cfgFile).Fatal("cfgFile exists, specify `overwrite` to replace")
	}
}

func writeConfig() {
	err := WriteToFile(embed.InitDotYamlBytes, cfgFile)
	if err != nil {
		logrus.WithError(err).WithField("cfgFile", cfgFile).Fatal("Error saving file")
	}

	logrus.WithField("cfgFile", cfgFile).Info("wrote sample foji config file")
}

func WriteToFile(source []byte, file string) error {
	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		return fmt.Errorf("error creating output directory:%w", err)
	}

	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("error opening output file:%w", err)
	}

	_, err = f.Write(source)
	if closeErr := f.Close(); err != nil {
		return fmt.Errorf("error closing output file:%w", closeErr)
	}

	return err //nolint:wrapcheck
}
