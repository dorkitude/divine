package cmd

import (
	"fmt"

	"github.com/alan-botts/divine/internal/deck"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Developer tools",
}

var validateDecksCmd = &cobra.Command{
	Use:   "validate-decks",
	Short: "Validate all deck structures",
	RunE:  runValidateDecks,
}

func init() {
	toolsCmd.AddCommand(validateDecksCmd)
	rootCmd.AddCommand(toolsCmd)
}

func runValidateDecks(cmd *cobra.Command, args []string) error {
	decks, err := deck.LoadAll()
	if err != nil {
		return fmt.Errorf("failed to load decks: %w", err)
	}

	okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	nameStyle := lipgloss.NewStyle().Bold(true)

	hasErrors := false
	for _, d := range decks {
		errs := d.Validate()
		if len(errs) == 0 {
			fmt.Printf("%s %s (%d cards)\n",
				okStyle.Render("OK"),
				nameStyle.Render(d.Meta.Name),
				len(d.Cards),
			)
		} else {
			hasErrors = true
			fmt.Printf("%s %s\n",
				errStyle.Render("ERR"),
				nameStyle.Render(d.Meta.Name),
			)
			for _, e := range errs {
				fmt.Printf("    %s\n", errStyle.Render(e))
			}
		}
	}

	if hasErrors {
		return fmt.Errorf("validation failed")
	}

	fmt.Printf("\nAll %d decks valid.\n", len(decks))
	return nil
}
