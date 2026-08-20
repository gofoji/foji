package cmd

import (
	"github.com/spf13/cobra"

	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/foji"
	"github.com/gofoji/foji/stringlist"
)

var copyTemplatesCmd = &cobra.Command{
	Use:     "copyProcessTemplates [list of processes]",
	Aliases: []string{"cpt"},
	Short:   "Copy embedded templates to a local directory",
	Long:    "By default it uses './foji' as the destination directory.  Use 'all' to dump all embedded templates.",
	Args:    cobra.MinimumNArgs(1),
	Run:     copyTemplates,
}

var copyTemplatesFlags struct {
	dir       string
	overwrite bool
}

func registerCopyTemplatesFlags() {
	copyTemplatesCmd.PersistentFlags().StringVarP(&copyTemplatesFlags.dir, "dir", "d", "foji", "output directory")
	copyTemplatesCmd.PersistentFlags().BoolVarP(&copyTemplatesFlags.overwrite, "overwrite", "y", false,
		"force overwrite an existing output file")
}

func copyTemplates(_ *cobra.Command, args []string) {
	l := rootFlags.logger()

	config, err := cfg.Load(rootFlags.cfgFile, true)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to load config")
	}

	var templates stringlist.Strings

	if len(args) == 1 && args[0] == "all" {
		templates, err = foji.AllTemplates()
		if err != nil {
			l.Fatal().Err(err).Msg("Failed to load templates")
		}
	} else {
		targets, err := config.Processes.Target(args)
		if err != nil {
			l.Fatal().Err(err).Msg("Failed to process targets")
		}

		if len(targets) == 0 {
			l.Fatal().Str("processes", config.Processes.String()).Strs("targets", args).Msg("No valid targets defined.")
		}

		templateMaps := cfg.OrderedMap{}

		for _, p := range targets {
			templateMaps.SetAll(p.All())
		}

		templates = templateMaps.Values()
	}

	for _, v := range templates {
		// This command declares no --stdout flag, so it always writes to disk.
		err = writeTemplate(l, copyTemplatesFlags.dir, v, false, copyTemplatesFlags.overwrite)
		if err != nil {
			l.Fatal().Str("template", v).Err(err).Msg("Failed to Write")
		}
	}
}
