package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func CategoryPills(categories []string, selected int) string {
	out := make([]string, 0, len(categories))
	for i, c := range categories {
		if i == selected {
			out = append(out, lipgloss.NewStyle().Foreground(lipgloss.Color("#1E1E2E")).Background(lipgloss.Color("#00D1FF")).Padding(0, 1).Render(c))
		} else {
			out = append(out, lipgloss.NewStyle().Foreground(lipgloss.Color("#7F849C")).Padding(0, 1).Render(c))
		}
	}
	return strings.Join(out, " ")
}
