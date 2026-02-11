package components

import "github.com/charmbracelet/lipgloss"

func ConfirmDialog(title, body string) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#F38BA8")).
		Padding(1, 2).
		Render(title + "\n\n" + body + "\n\n[y] yes   [n] no")
}
