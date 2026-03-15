package cmd

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alan-botts/divine/internal/deck"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	drawDeck     string
	drawCount    int
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

	decks, err := deck.LoadAll()
	if err != nil {
		return err
	}

	if len(decks) == 0 {
		return fmt.Errorf("no decks available in embedded data")
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
		card       deck.Card
		deckName   string
		deckAuthor string
	}
	var pool []cardWithDeck
	for _, d := range decks {
		for _, c := range d.Cards {
			pool = append(pool, cardWithDeck{
				card:       c,
				deckName:   d.Meta.Name,
				deckAuthor: d.Meta.Author,
			})
		}
	}

	if drawCount > len(pool) {
		drawCount = len(pool)
	}

	// Shuffle and pick
	indices := rand.Perm(len(pool))[:drawCount]

	// Styles
	titleBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("111")).
		Padding(0, 1).
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
		titleWidth := contentWidth - 4
		if titleWidth < 20 {
			titleWidth = 20
		}
		content.WriteString(titleBoxStyle.Render(wrapText(cwd.card.Title, titleWidth)) + "\n")
		content.WriteString(deckStyle.Render(fmt.Sprintf("from %s - by %s", cwd.deckName, cwd.deckAuthor)) + "\n")

		if len(cwd.card.Keywords) > 0 {
			content.WriteString(keywordStyle.Render(wrapText(strings.Join(cwd.card.Keywords, " | "), contentWidth)) + "\n")
		}

		details := cardDetails(cwd.card)
		if len(details) > 0 {
			content.WriteString(wrapText(strings.Join(details, "\n"), contentWidth) + "\n")
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

func cardDetails(card deck.Card) []string {
	if len(card.Fields) == 0 {
		return nil
	}

	keys := make([]string, 0, len(card.Fields))
	for k := range card.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	details := make([]string, 0, len(keys))
	for _, k := range keys {
		v := formatDetailValue(card.Fields[k])
		if v == "" {
			continue
		}
		details = append(details, fmt.Sprintf("%s: %s", k, v))
	}
	return details
}

func formatDetailValue(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(t)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", t)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", t)
	case float32, float64, bool:
		return fmt.Sprintf("%v", t)
	case []string:
		return strings.Join(t, ", ")
	case []interface{}:
		parts := make([]string, 0, len(t))
		for _, item := range t {
			s := formatDetailValue(item)
			if s != "" {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, ", ")
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", t))
	}
}
