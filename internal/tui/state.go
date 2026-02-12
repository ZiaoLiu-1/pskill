package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ZiaoLiu-1/pskill/internal/config"
)

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
	TabProjects
	TabSettings
	TabOnboarding // special full-screen tab, not in the tab bar
)

const TabCount = 7 // only the 7 main tabs cycle with tab/shift+tab

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

// onboardingDoneMsg is emitted when the onboarding wizard finishes.
// app.go catches it to save config, switch to dashboard, and trigger a scan.
type onboardingDoneMsg struct {
	cfg            config.Config
	scannedNames   []string
	scannedCount   int
}
