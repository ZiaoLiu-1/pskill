package tui

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary   = lipgloss.Color("#00D1FF")
	ColorSecondary = lipgloss.Color("#74C7EC")
	ColorSuccess   = lipgloss.Color("#A6E3A1")
	ColorWarning   = lipgloss.Color("#F9E2AF")
	ColorDanger    = lipgloss.Color("#F38BA8")
	ColorMuted     = lipgloss.Color("#6B7280")
	ColorSurface   = lipgloss.Color("#1E1E2E")
	ColorBright    = lipgloss.Color("#CDD6F4")
	ColorDim       = lipgloss.Color("#45475A")
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	brightStyle = lipgloss.NewStyle().
			Foreground(ColorBright)

	successStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	warningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	dangerStyle = lipgloss.NewStyle().
			Foreground(ColorDanger)

	selectedStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	activeTabStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Underline(true)

	tabStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	paneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDim).
			Padding(0, 1)

	activePaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)
)

func helpEntry(key, desc string) string {
	return helpKeyStyle.Render(key) + " " + helpDescStyle.Render(desc)
}
