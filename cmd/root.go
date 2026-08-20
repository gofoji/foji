package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/gofoji/foji/cfg"
)

// rootFlagSet holds the flags declared on the root command. They are
// persistent, so every subcommand inherits them.
//
// Each command owns its own flag struct rather than sharing package-level
// variables: `--overwrite` is declared by three commands and `--dir` by two, and
// binding them all to one variable meant the commands aliased each other and
// could only work because exactly one runs per process.
type rootFlagSet struct {
	cfgFile string
	verbose bool
	trace   bool
	quiet   bool
}

var rootFlags rootFlagSet

// logger builds a logger at the verbosity selected by the root flags.
func (f rootFlagSet) logger() zerolog.Logger {
	return getLogger(f.quiet, f.trace, f.verbose)
}

var rootCmd = &cobra.Command{
	Use:   "foji",
	Short: "Fōji Generator for Postgres, SQL, and OpenAPI (Swagger)",
	Long: `Fōji reads your database, static sql files, OpenApi V3 and generates code using templates.  
The output templates are easily customized to your needs.  
https://github.com/gofoji/foji
Version: ` + cfg.Version(),
}

func Execute() {
	registerFlags()
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(copyTemplatesCmd)
	rootCmd.AddCommand(copyTemplateCmd)
	rootCmd.AddCommand(dumpConfigCmd)
	rootCmd.AddCommand(weldCmd)
	rootCmd.AddCommand(listCmd)

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err) //nolint
		os.Exit(1)
	}
}

// registerFlags binds the root flags and then delegates to each command, which
// registers its own.
func registerFlags() {
	rootCmd.PersistentFlags().StringVarP(&rootFlags.cfgFile, "config", "c", "foji.yaml", "config file")
	rootCmd.PersistentFlags().BoolVarP(&rootFlags.verbose, "verbose", "v", false, "include verbose logging")
	rootCmd.PersistentFlags().BoolVarP(&rootFlags.trace, "trace", "t", false, "include trace logging")
	rootCmd.PersistentFlags().BoolVarP(&rootFlags.quiet, "quiet", "q", false, "mutes all logging (overrides verbose)")

	registerInitFlags()
	registerDumpConfigFlags()
	registerCopyTemplateFlags()
	registerCopyTemplatesFlags()
	registerWeldFlags()
}
