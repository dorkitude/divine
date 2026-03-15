package cmd

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alan-botts/divine/internal/deck"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	drawDeck    string
	drawCount   int
	drawAllDecks bool
)

var drawCmd = &cobra.Command{
	Use:   "draw",
	Short: "Draw one or more cards",
	Long:  "Randomly draw cards from available decks.",
	RunE:  runDraw,
}

func init() {
	drawCmd.Flags().StringVar(&drawDeck, "deck", "", "Draw from a specific deck (by directory name)")
	drawCmd.Flags().IntVarP(&drawCount, "n", "n", 1, "Number of cards to draw")
	drawCmd.Flags().BoolVar(&drawAllDecks, "all-decks", true, "Draw from all decks (default)")
	rootCmd.AddCommand(drawCmd)
}

func runDraw(cmd *cobra.Command, args []string) error {
	rand.Seed(time.Now().UnixNano())

	decksDir, err := deck.FindDecksDir()
	if err != nil {
		return err
	}

	decks, err := deck.LoadAll(decksDir)
	if err != nil {
		return err
	}

	if len(decks) == 0 {
		return fmt.Errorf("no decks found in %s", decksDir)
	}

	// Filter to specific deck if requested
	if drawDeck != "" {
		var filtered []deck.Deck
		for _, d := range decks {
			if d.DirName == drawDeck {
				filtered = append(filtered, d)
			}
		}
		if len(filtered) == 0 {
			fmt.Fprintf(os.Stderr, "Unknown deck %q. Available decks:\n", drawDeck)
			for _, d := range decks {
				fmt.Fprintf(os.Stderr, "  %s: %s\n", d.DirName, d.Meta.Name)
			}
			return fmt.Errorf("deck not found: %q", drawDeck)
		}
		decks = filtered
	}

	// Collect all cards with deck info
	type cardWithDeck struct {
		card     deck.Card
		deckName string
	}
	var pool []cardWithDeck
	for _, d := range decks {
		for _, c := range d.Cards {
			pool = append(pool, cardWithDeck{card: c, deckName: d.Meta.Name})
		}
	}

	if drawCount > len(pool) {
		drawCount = len(pool)
	}

	// Shuffle and pick
	indices := rand.Perm(len(pool))[:drawCount]

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("219"))

	deckStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("248"))

	keywordStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("180"))

	dividerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		MarginBottom(1)

	contentWidth := terminalContentWidth()

	for i, idx := range indices {
		cwd := pool[idx]

		var content strings.Builder
		content.WriteString(titleStyle.Render(wrapText(cwd.card.Title, contentWidth)) + "\n")
		content.WriteString(deckStyle.Render("from " + cwd.deckName) + "\n")

		if len(cwd.card.Keywords) > 0 {
			content.WriteString(keywordStyle.Render(wrapText(strings.Join(cwd.card.Keywords, " | "), contentWidth)) + "\n")
		}

		content.WriteString(dividerStyle.Render(strings.Repeat("─", contentWidth)) + "\n")
		content.WriteString("\n")
		content.WriteString(wrapText(cwd.card.Body, contentWidth))

		fmt.Println(borderStyle.Render(content.String()))

		if i < len(indices)-1 {
			fmt.Println()
		}
	}

	return nil
}

func terminalContentWidth() int {
	const (
		defaultTerminalWidth = 100
		minContentWidth      = 40
		maxContentWidth      = 88
		innerOverhead        = 6 // 2 border + 4 horizontal padding
	)

	cols := defaultTerminalWidth
	if raw := strings.TrimSpace(os.Getenv("COLUMNS")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			cols = n
		}
	}

	width := cols - innerOverhead
	if width < minContentWidth {
		return minContentWidth
	}
	if width > maxContentWidth {
		return maxContentWidth
	}
	return width
}

func wrapText(text string, width int) string {
	if width <= 0 || text == "" {
		return text
	}

	lines := strings.Split(text, "\n")
	wrapped := make([]string, 0, len(lines))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			wrapped = append(wrapped, "")
			continue
		}
		wrapped = append(wrapped, wrapLine(line, width)...)
	}

	return strings.Join(wrapped, "\n")
}

func wrapLine(line string, width int) []string {
	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{""}
	}

	var out []string
	current := words[0]
	currentLen := utf8.RuneCountInString(current)

	for _, w := range words[1:] {
		wLen := utf8.RuneCountInString(w)
		if currentLen+1+wLen <= width {
			current += " " + w
			currentLen += 1 + wLen
			continue
		}
		out = append(out, current)
		current = w
		currentLen = wLen
	}

	out = append(out, current)
	return out
}
