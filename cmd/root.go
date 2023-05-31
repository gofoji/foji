package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gofoji/foji/cfg"
)

var (
	cfgFile         string
	verbose         bool
	trace           bool
	quiet           bool
	overwrite       bool
	includeDefaults bool
	stdout          bool
	simulate        bool
	dir             string
	dumpWeld        string
)

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

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err) //nolint
		os.Exit(1)
	}
}

func registerFlags() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "foji.yaml", "config file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "include verbose logging")
	rootCmd.PersistentFlags().BoolVarP(&trace, "trace", "t", false, "include trace logging")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "mutes all logging (overrides verbose)")

	initCmd.PersistentFlags().BoolVarP(&overwrite, "overwrite", "y", false,
		"force overwrite an existing output file")

	dumpConfigCmd.PersistentFlags().BoolVarP(&includeDefaults, "includeDefaults", "d", false,
		"Include evaluated Fōji defaults in the dump")
	dumpConfigCmd.PersistentFlags().StringVarP(&dumpWeld, "weld", "w", "",
		"limits output to only the specified weld config (includes defaults")

	copyTemplateCmd.PersistentFlags().BoolVarP(&stdout, "stdout", "o", false, "write to stdout")
	copyTemplateCmd.PersistentFlags().BoolVarP(&overwrite, "overwrite", "y", false,
		"force overwrite an existing output file")
	copyTemplateCmd.PersistentFlags().StringVarP(&dir, "dir", "d", "foji", "output directory")

	copyTemplatesCmd.PersistentFlags().StringVarP(&dir, "dir", "d", "foji", "output directory")
	copyTemplatesCmd.PersistentFlags().BoolVarP(&overwrite, "overwrite", "y", false,
		"force overwrite an existing output file")

	weldCmd.PersistentFlags().BoolVarP(&simulate, "simulate", "s", false,
		"simulates the processing, only displays files that would be generated")
}
