package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/scanner"
	"github.com/ZiaoLiu-1/pskill/internal/store"
)

type toastMsg struct {
	text     string
	duration time.Duration
}

type toastTimeoutMsg struct{}

type App struct {
	cfg        config.Config
	debug      bool
	width      int
	height     int
	activeTab  AppTabID
	tabs       map[AppTabID]Tab
	status     string
	project    string
	onboarding bool
	toast      string
}

func NewApp(cfg config.Config, debug bool) *App {
	// Check if first run -- start onboarding wizard
	if config.IsFirstRun() {
		return newAppOnboarding(cfg, debug)
	}
	return NewAppWithTab(cfg, debug, TabDashboard)
}

// NewAppOnboarding creates the app in onboarding mode (for pskill init --force too).
func NewAppOnboarding(cfg config.Config, debug bool) *App {
	return newAppOnboarding(cfg, debug)
}

func newAppOnboarding(cfg config.Config, debug bool) *App {
	tabs := map[AppTabID]Tab{
		TabOnboarding: NewOnboardingTab(cfg),
	}
	return &App{
		cfg:        cfg,
		debug:      debug,
		activeTab:  TabOnboarding,
		tabs:       tabs,
		status:     "Setup",
		project:    cwdBase(),
		onboarding: true,
	}
}

func NewAppWithTab(cfg config.Config, debug bool, startTab AppTabID) *App {
	tabs := map[AppTabID]Tab{
		TabDashboard: NewDashboardTab(cfg),
		TabMySkills:  NewSkillsTab(cfg),
		TabDiscover:  NewDiscoverTab(cfg),
		TabTrending:  NewTrendingTab(cfg),
		TabMonitor:   NewMonitorTab(cfg),
		TabSettings:  NewSettingsTab(cfg),
	}
	return &App{
		cfg:       cfg,
		debug:     debug,
		activeTab: startTab,
		tabs:      tabs,
		status:    "Ready",
		project:   cwdBase(),
	}
}

func (a *App) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(a.tabs))
	for _, t := range a.tabs {
		if c := t.Init(); c != nil {
			cmds = append(cmds, c)
		}
	}
	// Only auto-scan when NOT in onboarding (onboarding does its own scan)
	if !a.onboarding {
		cmds = append(cmds, a.scanSystemCmd())
	}
	return tea.Batch(cmds...)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = m.Width
		a.height = m.Height
		// Forward to active tab (onboarding needs it)
		if t, ok := a.tabs[a.activeTab]; ok {
			nt, cmd := t.Update(msg)
			a.tabs[a.activeTab] = nt
			return a, cmd
		}
		return a, nil

	case onboardingDoneMsg:
		// Transition from onboarding to normal mode
		a.cfg = m.cfg
		a.onboarding = false
		// Build the normal tabs
		a.tabs = map[AppTabID]Tab{
			TabDashboard: NewDashboardTab(a.cfg),
			TabMySkills:  NewSkillsTab(a.cfg),
			TabDiscover:  NewDiscoverTab(a.cfg),
			TabTrending:  NewTrendingTab(a.cfg),
			TabMonitor:   NewMonitorTab(a.cfg),
			TabSettings:  NewSettingsTab(a.cfg),
		}
		a.activeTab = TabDashboard
		a.status = fmt.Sprintf("Setup complete - %d skills", m.scannedCount)

		// Initialize all tabs and broadcast scanned skills
		cmds := make([]tea.Cmd, 0)
		for _, t := range a.tabs {
			if c := t.Init(); c != nil {
				cmds = append(cmds, c)
			}
		}
		// Forward scanned skills to My Skills tab and Dashboard
		scannedMsg := skillsScannedMsg{names: m.scannedNames, count: m.scannedCount}
		if tab, ok := a.tabs[TabMySkills]; ok {
			nt, cmd := tab.Update(scannedMsg)
			a.tabs[TabMySkills] = nt
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		if tab, ok := a.tabs[TabDashboard]; ok {
			nt, cmd := tab.Update(scannedMsg)
			a.tabs[TabDashboard] = nt
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return a, tea.Batch(cmds...)

	case skillsScannedMsg:
		a.status = fmt.Sprintf("Scanned %d skills", m.count)
		// Forward to skills tab and dashboard
		var cmds []tea.Cmd
		if tab, ok := a.tabs[TabMySkills]; ok {
			nt, cmd := tab.Update(m)
			a.tabs[TabMySkills] = nt
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		if tab, ok := a.tabs[TabDashboard]; ok {
			nt, cmd := tab.Update(m)
			a.tabs[TabDashboard] = nt
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return a, tea.Batch(cmds...)

	case statusMsg:
		a.status = m.text
		return a, nil

	case toastMsg:
		a.toast = m.text
		return a, tea.Tick(m.duration, func(t time.Time) tea.Msg {
			return toastTimeoutMsg{}
		})

	case toastTimeoutMsg:
		a.toast = ""
		return a, nil

	case tea.KeyMsg:
		// During onboarding, only intercept ctrl+c, everything else goes to onboarding tab
		if a.onboarding {
			if m.String() == "ctrl+c" {
				return a, tea.Quit
			}
			t := a.tabs[a.activeTab]
			nt, cmd := t.Update(msg)
			a.tabs[a.activeTab] = nt
			return a, cmd
		}

		// Normal mode key handling
		if a.tabs[a.activeTab].AcceptsTextInput() {
			switch m.String() {
			case "ctrl+c":
				return a, tea.Quit
			case "esc":
				// Let tab handle esc
			default:
				// Don't intercept
			}
		} else {
			switch m.String() {
			case "q", "ctrl+c":
				return a, tea.Quit
			case "1":
				a.activeTab = TabDashboard
				return a, nil
			case "2":
				a.activeTab = TabMySkills
				return a, nil
			case "3":
				a.activeTab = TabDiscover
				return a, nil
			case "4":
				a.activeTab = TabTrending
				return a, nil
			case "5":
				a.activeTab = TabMonitor
				return a, nil
			case "6":
				a.activeTab = TabSettings
				return a, nil
			}
		}

		// Tab/shift+tab always work for navigation
		switch m.String() {
		case "tab":
			a.activeTab = (a.activeTab + 1) % TabCount
			return a, nil
		case "shift+tab":
			a.activeTab = (a.activeTab + TabCount - 1) % TabCount
			return a, nil
		}
	}

	// Forward all messages to active tab
	t := a.tabs[a.activeTab]
	nt, cmd := t.Update(msg)
	a.tabs[a.activeTab] = nt
	return a, cmd
}

func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "  Initializing pskill..."
	}

	w := a.width

	// During onboarding: full-screen wizard, no chrome
	if a.onboarding {
		help := a.renderHelpBar(w)
		contentH := a.height - 1
		if contentH < 3 {
			contentH = 3
		}
		content := a.tabs[a.activeTab].View(w, contentH)
		return lipgloss.JoinVertical(lipgloss.Left, content, help)
	}

	// Normal mode: header + tabs + separator + content + help
	header := a.renderHeader(w)
	tabBar := a.renderTabBar(w)
	sep := dimStyle.Render(strings.Repeat("─", w))
	help := a.renderHelpBar(w)

	contentH := a.height - 4
	if contentH < 3 {
		contentH = 3
	}

	content := a.tabs[a.activeTab].View(w, contentH)

	return lipgloss.JoinVertical(lipgloss.Left, header, tabBar, sep, content, help)
}

func (a *App) renderHeader(w int) string {
	left := titleStyle.Render("pskill v0.1.0")
	mid := dimStyle.Render(fmt.Sprintf(" %s ", a.project))

	badges := make([]string, 0, len(a.cfg.TargetCLIs))
	for _, cli := range a.cfg.TargetCLIs {
		badges = append(badges, cliBadgeInline(cli))
	}
	right := strings.Join(badges, " ")

	gap := w - lipgloss.Width(left) - lipgloss.Width(mid) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return left + mid + strings.Repeat(" ", gap) + right
}

func (a *App) renderTabBar(w int) string {
	labels := []struct {
		key  string
		name string
		id   AppTabID
	}{
		{"1", "Dashboard", TabDashboard},
		{"2", "My Skills", TabMySkills},
		{"3", "Discover", TabDiscover},
		{"4", "Trending", TabTrending},
		{"5", "Monitor", TabMonitor},
		{"6", "Settings", TabSettings},
	}

	parts := make([]string, 0, len(labels))
	for _, l := range labels {
		label := fmt.Sprintf("[%s] %s", l.key, l.name)
		if l.id == a.activeTab {
			parts = append(parts, activeTabStyle.Render(label))
		} else {
			parts = append(parts, tabStyle.Render(label))
		}
		parts = append(parts, "  ")
	}
	return strings.Join(parts, "")
}

func (a *App) renderHelpBar(w int) string {
	tabHelp := a.tabs[a.activeTab].ShortHelp()

	parts := make([]string, 0, len(tabHelp)+3)
	for _, h := range tabHelp {
		parts = append(parts, h)
	}
	parts = append(parts, helpEntry("tab", "switch"))
	parts = append(parts, helpEntry("q", "quit"))

	left := strings.Join(parts, "  ")
	right := successStyle.Render(a.status)
	if a.toast != "" {
		right = activeTabStyle.Render(a.toast)
	}

	gap := w - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	return left + strings.Repeat(" ", gap) + right
}

func cliBadgeInline(cli string) string {
	switch cli {
	case "cursor":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#1E1E2E")).Background(ColorSuccess).Padding(0, 1).Render("Cursor")
	case "claude":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#1E1E2E")).Background(ColorWarning).Padding(0, 1).Render("Claude")
	case "codex":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#1E1E2E")).Background(ColorSecondary).Padding(0, 1).Render("Codex")
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#1E1E2E")).Background(ColorMuted).Padding(0, 1).Render(cli)
	}
}

func (a *App) scanSystemCmd() tea.Cmd {
	return func() tea.Msg {
		inv, err := scanner.ScanSystemSkills()
		if err != nil {
			return skillsScannedMsg{names: []string{}, count: 0}
		}
		// Import into store
		st := store.NewManager(a.cfg.StoreDir)
		names := make([]string, 0, len(inv.Skills))
		for _, sk := range inv.Skills {
			_ = st.ImportSkill(sk)
			names = append(names, sk.Name)
		}
		// Also list anything already in store
		existing, _ := st.ListSkills()
		seen := map[string]bool{}
		for _, n := range names {
			seen[n] = true
		}
		for _, n := range existing {
			if !seen[n] {
				names = append(names, n)
				seen[n] = true
			}
		}
		return skillsScannedMsg{names: names, count: len(names)}
	}
}

func cwdBase() string {
	wd, err := os.Getwd()
	if err != nil {
		return "~"
	}
	return filepath.Base(wd)
}

func logDebug(enabled bool, msg string) {
	if !enabled {
		return
	}
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".pskill", "debug.log")
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	line := fmt.Sprintf("%s %s\n", time.Now().Format(time.RFC3339), strings.TrimSpace(msg))
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(line)
}
