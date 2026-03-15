package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	version   = "0.0.2"
	commit    = "unknown"
	buildDate = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "divine",
	Short: "A general-purpose divination CLI",
	Long:  "Draw cards from tarot, I Ching, creative prompts, and more.",
	Version: "0.0.2",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	if storedVersion := defaultVersionFromFile(); storedVersion != "" {
		version = storedVersion
	}
	rootCmd.Version = version

	rootCmd.SetVersionTemplate("divine {{.Version}}\n")
	rootCmd.AddCommand(versionCmd)
}

func defaultVersionFromFile() string {
	versionPath, err := findVersionFile()
	if err != nil {
		return ""
	}

	value, err := os.ReadFile(versionPath)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(value))
}
