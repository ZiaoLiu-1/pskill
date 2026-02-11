package tui

import tea "github.com/charmbracelet/bubbletea"

type ViewState int

const (
	StateList ViewState = iota
	StateDetail
	StateConfirm
	StateOverlay
	StateSearch
)

type AppTabID int

const (
	TabDashboard AppTabID = iota
	TabMySkills
	TabDiscover
	TabTrending
	TabMonitor
	TabSettings
)

const TabCount = 6

type Tab interface {
	Init() tea.Cmd
	Update(tea.Msg) (Tab, tea.Cmd)
	View(width, height int) string
	Title() string
	ShortHelp() []string
	AcceptsTextInput() bool
}

type statusMsg struct {
	text string
}

type skillsScannedMsg struct {
	names []string
	count int
}
