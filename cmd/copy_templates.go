package cmd

import (
	"regexp"

	"github.com/gofoji/foji/cfg"
	"github.com/gofoji/foji/embed"
	"github.com/gofoji/foji/stringlist"
	"github.com/spf13/cobra"
)

var copyTemplatesCmd = &cobra.Command{
	Use:     "copyProcessTemplates [list of processes]",
	Aliases: []string{"cpt"},
	Short:   "Copy embedded templates to a local directory",
	Long:    "By default it uses './foji' as the destination directory.  Use 'all' to dump all embedded templates.",
	Args:    cobra.MinimumNArgs(1),
	Run:     copyTemplates,
}

func copyTemplates(_ *cobra.Command, args []string) {
	l := getLogger(quiet, trace, verbose)

	c, err := cfg.Load(cfgFile, true)
	if err != nil {
		l.WithError(err).Fatal("Failed to load config")
	}

	var templates stringlist.Strings

	if len(args) == 1 && args[0] == "all" {
		templates = embed.List()
		templates = templates.Filter(templateRegex.MatchString)
	} else {
		targets, err := c.Processes.Target(args)
		if err != nil {
			l.WithError(err).Fatal("Failed to process targets")
		}

		if len(targets) == 0 {
			l.WithField("processes", c.Processes.String()).WithField("targets", args).Fatal("No valid targets defined.")
		}

		templateMaps := stringlist.StringMap{}

		for _, p := range targets {
			templateMaps = cfg.MergeTypesMaps(templateMaps, p.All())
		}

		templates = templateMaps.Values()
	}

	for _, v := range templates {
		err = writeTemplate(l, dir, v, stdout, overwrite)
		if err != nil {
			l.WithField("template", v).WithError(err).Fatal("Failed to Write")
		}
	}
}

var templateRegex = regexp.MustCompile("^foji/")
