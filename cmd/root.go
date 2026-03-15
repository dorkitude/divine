package cmd

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed ../VERSION
var versionFile string

var (
	version   = strings.TrimSpace(versionFile)
	commit    = "unknown"
	buildDate = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "divine",
	Short: "A general-purpose divination CLI",
	Long:  "Draw cards from tarot, I Ching, creative prompts, and more.",
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate("divine {{.Version}}\n")
	rootCmd.AddCommand(versionCmd)
}
