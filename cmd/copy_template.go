package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofoji/foji/embed"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
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
	c, err := embed.Get(filename)
	if err != nil {
		return fmt.Errorf("failed to read template:%w", err)
	}

	if useStdout {
		_, err = os.Stdout.Write(c)

		return err //nolint:wrapcheck
	}

	if dir != "" {
		filename = changeDirectory(dir, "foji", filename)
	}

	l = l.With().Str("template", filename).Logger()

	if useStdout || overwrite || !fileExists(filename) {
		l.Debug().Msg("Writing")

		err = WriteToFile(c, filename)
		if err != nil {
			return fmt.Errorf("failed to write template:%w", err)
		}
	} else {
		l.Warn().Msg("Skipping, specify `overwrite` to replace")
	}

	return nil
}

func fileExists(filename string) bool {
	fileInfo, err := os.Stat(filename)

	return err == nil && fileInfo.Mode().IsRegular()
}

func changeDirectory(dir, swapDir, filename string) string {
	path := strings.Split(filename, string(os.PathSeparator))
	if len(path) == 0 {
		return filename
	}

	if path[0] == swapDir {
		path[0] = dir
	} else {
		path = append([]string{dir}, path...)
	}

	return filepath.Join(path...)
}
