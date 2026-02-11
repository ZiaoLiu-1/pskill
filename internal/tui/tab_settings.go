package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/detector"
)

type SettingsTab struct {
	cfg    config.Config
	clis   []detector.CLIInfo
	cursor int
}

func NewSettingsTab(cfg config.Config) Tab {
	return &SettingsTab{cfg: cfg, clis: []detector.CLIInfo{}}
}

func (t *SettingsTab) Init() tea.Cmd {
	return func() tea.Msg {
		clis, _ := detector.DetectInstalledCLIs()
		return clis
	}
}

func (t *SettingsTab) Update(msg tea.Msg) (Tab, tea.Cmd) {
	switch m := msg.(type) {
	case []detector.CLIInfo:
		t.clis = m
	case tea.KeyMsg:
		switch m.String() {
		case "j", "down":
			if t.cursor < len(t.clis)-1 {
				t.cursor++
			}
		case "k", "up":
			if t.cursor > 0 {
				t.cursor--
			}
		case " ":
			if len(t.clis) > 0 && t.cursor < len(t.clis) {
				t.clis[t.cursor].SupportsSkills = !t.clis[t.cursor].SupportsSkills
			}
		case "a":
			t.cfg.AutoUpdateTrending = !t.cfg.AutoUpdateTrending
		}
	}
	return t, nil
}

func (t *SettingsTab) View(width, height int) string {
	home, _ := os.UserHomeDir()
	shorten := func(p string) string { return strings.Replace(p, home, "~", 1) }

	var b strings.Builder

	// Section: CLI Integrations
	b.WriteString(titleStyle.Render("  CLI Integrations") + "\n")
	b.WriteString(dimStyle.Render("  " + strings.Repeat("─", 45)) + "\n")
	for i, c := range t.clis {
		check := dimStyle.Render("[ ]")
		if c.SupportsSkills && c.Installed {
			check = successStyle.Render("[x]")
		}
		prefix := "  "
		if i == t.cursor {
			prefix = selectedStyle.Render("> ")
		}
		status := ""
		if !c.Installed {
			status = dangerStyle.Render(" (not found)")
		}
		b.WriteString(fmt.Sprintf("%s%s %-10s %s%s\n", prefix, check, c.Name, dimStyle.Render(shorten(c.BaseDir)), status))
	}

	// Section: Paths
	b.WriteString("\n" + titleStyle.Render("  Paths & Storage") + "\n")
	b.WriteString(dimStyle.Render("  " + strings.Repeat("─", 45)) + "\n")
	b.WriteString(fmt.Sprintf("  Central Store:  %s\n", dimStyle.Render(shorten(t.cfg.StoreDir))))
	b.WriteString(fmt.Sprintf("  Search Index:   %s\n", dimStyle.Render(shorten(t.cfg.IndexDir))))
	b.WriteString(fmt.Sprintf("  Stats DB:       %s\n", dimStyle.Render(shorten(t.cfg.StatsDB))))

	// Section: Registry
	b.WriteString("\n" + titleStyle.Render("  Registry") + "\n")
	b.WriteString(dimStyle.Render("  " + strings.Repeat("─", 45)) + "\n")
	b.WriteString(fmt.Sprintf("  URL:            %s\n", dimStyle.Render(t.cfg.RegistryURL)))
	auto := dimStyle.Render("[ ] No")
	if t.cfg.AutoUpdateTrending {
		auto = successStyle.Render("[x] Yes")
	}
	b.WriteString(fmt.Sprintf("  Auto-update:    %s\n", auto))

	// Section: Defaults
	b.WriteString("\n" + titleStyle.Render("  Default Skills") + "\n")
	b.WriteString(dimStyle.Render("  " + strings.Repeat("─", 45)) + "\n")
	if len(t.cfg.DefaultSkills) == 0 {
		b.WriteString(dimStyle.Render("  (none configured)") + "\n")
	} else {
		for _, s := range t.cfg.DefaultSkills {
			b.WriteString("  - " + brightStyle.Render(s) + "\n")
		}
	}

	return activePaneStyle.Width(width - 4).Height(height - 2).Render(b.String())
}

func (t *SettingsTab) Title() string { return "Settings" }
func (t *SettingsTab) ShortHelp() []string {
	return []string{
		helpEntry("j/k", "nav"),
		helpEntry("space", "toggle"),
		helpEntry("a", "auto-update"),
	}
}
func (t *SettingsTab) AcceptsTextInput() bool { return false }
