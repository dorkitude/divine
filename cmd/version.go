package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of divine",
	Run: func(cmd *cobra.Command, args []string) {
		if commit == "unknown" && buildDate == "unknown" {
			fmt.Printf("divine %s\n", version)
			return
		}
		fmt.Printf("divine %s (commit=%s, date=%s)\n", version, commit, buildDate)
	},
}
