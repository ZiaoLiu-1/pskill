package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/detector"
)

type DashboardTab struct {
	cfg        config.Config
	clis       []detector.CLIInfo
	skillCount int
	ready      bool
}

func NewDashboardTab(cfg config.Config) Tab {
	return &DashboardTab{cfg: cfg}
}

func (t *DashboardTab) Init() tea.Cmd {
	return func() tea.Msg {
		clis, _ := detector.DetectInstalledCLIs()
		return clis
	}
}

func (t *DashboardTab) Update(msg tea.Msg) (Tab, tea.Cmd) {
	switch m := msg.(type) {
	case []detector.CLIInfo:
		t.clis = m
		t.ready = true
	case skillsScannedMsg:
		t.skillCount = m.count
	}
	return t, nil
}

func (t *DashboardTab) View(width, height int) string {
	if !t.ready {
		return "\n  Scanning system..."
	}

	var b strings.Builder

	// Logo
	logo := titleStyle.Render(Logo)

	b.WriteString(logo)
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("    Universal LLM Skill Manager"))
	b.WriteString("\n\n")

	// System status
	b.WriteString(brightStyle.Render("  System Status"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  " + strings.Repeat("─", 40)))
	b.WriteString("\n")

	for _, c := range t.clis {
		status := dangerStyle.Render("not found")
		if c.Installed {
			if c.SupportsSkills {
				status = successStyle.Render("ready")
			} else {
				status = warningStyle.Render("no skill support")
			}
		}
		b.WriteString(fmt.Sprintf("  %-10s %s\n", c.Name, status))
	}

	b.WriteString(fmt.Sprintf("\n  Skills in store: %s\n", titleStyle.Render(fmt.Sprintf("%d", t.skillCount))))
	home, _ := os.UserHomeDir()
	b.WriteString(fmt.Sprintf("  Store path:      %s\n", dimStyle.Render(strings.Replace(t.cfg.StoreDir, home, "~", 1))))
	b.WriteString(fmt.Sprintf("  Registry:        %s\n", dimStyle.Render(t.cfg.RegistryURL)))

	// Quick actions
	b.WriteString("\n")
	b.WriteString(brightStyle.Render("  Quick Actions"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  " + strings.Repeat("─", 40)))
	b.WriteString("\n\n")

	actions := []struct {
		key  string
		desc string
	}{
		{"2", "Browse local skills"},
		{"3", "Search & discover new skills"},
		{"4", "View trending skills"},
		{"5", "Usage monitor dashboard"},
		{"6", "Settings & CLI configuration"},
	}

	for _, a := range actions {
		b.WriteString(fmt.Sprintf("  %s  %s\n",
			helpKeyStyle.Render(fmt.Sprintf("[%s]", a.key)),
			brightStyle.Render(a.desc),
		))
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  Press a number key to navigate, or tab to cycle."))
	b.WriteString("\n")

	content := b.String()

	// Center content
	contentW := lipgloss.Width(content)
	if contentW < width {
		pad := (width - contentW) / 4
		if pad > 0 {
			lines := strings.Split(content, "\n")
			padded := make([]string, 0, len(lines))
			prefix := strings.Repeat(" ", pad)
			for _, l := range lines {
				padded = append(padded, prefix+l)
			}
			content = strings.Join(padded, "\n")
		}
	}

	return content
}

func (t *DashboardTab) Title() string { return "Dashboard" }
func (t *DashboardTab) ShortHelp() []string {
	return []string{
		helpEntry("1-6", "navigate"),
		helpEntry("tab", "next"),
	}
}
func (t *DashboardTab) AcceptsTextInput() bool { return false }
