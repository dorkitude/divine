package cmd

import (
	"fmt"

	"github.com/alan-botts/divine/internal/deck"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var decksCmd = &cobra.Command{
	Use:   "decks",
	Short: "List all available decks",
	RunE:  runDecks,
}

func init() {
	rootCmd.AddCommand(decksCmd)
}

func runDecks(cmd *cobra.Command, args []string) error {
	decks, err := deck.LoadAll()
	if err != nil {
		return err
	}

	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))

	for _, d := range decks {
		fmt.Println(nameStyle.Render(fmt.Sprintf("%s: %s", d.DirName, d.Meta.Name)))
		fmt.Println(descStyle.Render(d.Meta.Description))
		fmt.Println(metaStyle.Render(fmt.Sprintf(
			"  %d cards | %s | %s",
			len(d.Cards), d.Meta.Author, d.Meta.LicenseType,
		)))
		fmt.Println()
	}

	return nil
}
