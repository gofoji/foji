package cmd

import (
	"os"

	"github.com/gofoji/foji/cfg"
	"github.com/sirupsen/logrus"
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
	c, err := cfg.Load(cfgFile, includeDefaults)

	if err != nil {
		logrus.WithError(err).Fatal("Loading Config")
	}

	err = yaml.NewEncoder(os.Stdout).Encode(c)
	if err != nil {
		logrus.WithError(err).Fatal("Getting Database Schema")
	}
}
