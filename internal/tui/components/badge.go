package components

import "github.com/charmbracelet/lipgloss"

func CLIBadge(cli string) string {
	style := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#1E1E2E"))
	switch cli {
	case "cursor":
		return style.Background(lipgloss.Color("#A6E3A1")).Render("C")
	case "claude":
		return style.Background(lipgloss.Color("#F9E2AF")).Render("Cl")
	case "codex":
		return style.Background(lipgloss.Color("#89B4FA")).Render("Co")
	default:
		return style.Background(lipgloss.Color("#7F849C")).Render("?")
	}
}
