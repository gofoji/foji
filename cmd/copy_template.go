package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/gofoji/foji/foji"
)

var copyTemplateCmd = &cobra.Command{
	Use:     "copy [template name]",
	Aliases: []string{"cp"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Dump a copy of the default templates to a local directory (./foji by default)",
	Run:     copyTemplate,
}

func copyTemplate(_ *cobra.Command, args []string) {
	l := getLogger(quiet, trace, verbose)

	err := writeTemplate(l, dir, args[0], stdout, overwrite)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to Write Template")
	}
}

func writeTemplate(l zerolog.Logger, dir, filename string, useStdout, overwrite bool) error {
	b, err := foji.Default(filename)
	if err != nil {
		return fmt.Errorf("failed to read template:%w", err)
	}

	if useStdout {
		_, err = os.Stdout.Write(b)

		return err //nolint:wrapcheck
	}

	if dir != "" {
		filename = ChangeDirectory(dir, "foji", filename)
	}

	l = l.With().Str("template", filename).Logger()

	if useStdout || overwrite || !FileExists(filename) {
		l.Debug().Msg("Writing")

		err = WriteToFile(b, filename)
		if err != nil {
			return fmt.Errorf("failed to write template:%w", err)
		}
	} else {
		l.Warn().Msg("Skipping, specify `overwrite` to replace")
	}

	return nil
}
