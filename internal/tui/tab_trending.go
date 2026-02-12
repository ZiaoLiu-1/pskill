package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/installer"
	"github.com/ZiaoLiu-1/pskill/internal/registry"
)

type trendingMsg struct {
	items []registry.SkillResult
	total int
	err   error
}

type trendingInstallDoneMsg struct {
	result *installer.Result
	err    error
}

type TrendingTab struct {
	cfg        config.Config
	items      []registry.SkillResult
	total      int
	cursor     int
	page       int
	pageSize   int
	loading    bool
	installing bool
	confirming bool // true when confirmation modal is shown
	errMsg     string
}

func NewTrendingTab(cfg config.Config) Tab {
	return &TrendingTab{cfg: cfg, page: 1, pageSize: 20}
}

func (t *TrendingTab) Init() tea.Cmd { return t.loadCmd() }

func (t *TrendingTab) Update(msg tea.Msg) (Tab, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		// Confirmation modal keys
		if t.confirming {
			switch m.String() {
			case "y", "Y", "enter":
				t.confirming = false
				t.installing = true
				return t, t.installCmd(t.items[t.cursor])
			case "n", "N", "esc":
				t.confirming = false
			}
			return t, nil
		}

		switch m.String() {
		case "j", "down":
			if t.cursor < len(t.items)-1 {
				t.cursor++
			}
		case "k", "up":
			if t.cursor > 0 {
				t.cursor--
			}
		case "r":
			return t, t.loadCmd()
		case ">", ".":
			if t.page*t.pageSize < t.total {
				t.page++
				t.cursor = 0
				return t, t.loadCmd()
			}
		case "<", ",":
			if t.page > 1 {
				t.page--
				t.cursor = 0
				return t, t.loadCmd()
			}
		case "enter":
			if !t.installing && len(t.items) > 0 && t.cursor < len(t.items) {
				t.confirming = true
			}
		}

	case trendingMsg:
		t.loading = false
		if m.err != nil {
			t.errMsg = m.err.Error()
		} else {
			t.items = m.items
			t.total = m.total
			t.errMsg = ""
		}

	case trendingInstallDoneMsg:
		t.installing = false
		if m.err != nil {
			t.errMsg = m.err.Error()
			return t, func() tea.Msg {
				return toastMsg{text: "Install failed: " + m.err.Error(), duration: 3 * time.Second}
			}
		}
		summary := "Installed " + m.result.SkillName
		if len(m.result.LinkedCLIs) > 0 {
			summary += " → " + strings.Join(m.result.LinkedCLIs, ", ")
		}
		return t, tea.Batch(
			func() tea.Msg { return statusMsg{text: summary} },
			func() tea.Msg { return toastMsg{text: "Installed " + m.result.SkillName, duration: 3 * time.Second} },
		)
	}
	return t, nil
}

func (t *TrendingTab) View(width, height int) string {
	l := ComputeLayout(width, height, true)
	if l.IsTooSmall {
		return RenderTooSmall(width, height)
	}

	// === Left pane: list ===
	var list strings.Builder
	list.WriteString(brightStyle.Render("Trending Skills"))
	totalPages := (t.total + t.pageSize - 1) / t.pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	list.WriteString(dimStyle.Render(fmt.Sprintf("  sorted by ★ stars  (%d total)  page %d/%d", t.total, t.page, totalPages)))
	list.WriteString("\n\n")

	if t.loading {
		list.WriteString(warningStyle.Render("  Loading from skillsmp.com...") + "\n")
	}
	if t.installing {
		list.WriteString(warningStyle.Render("  Installing...") + "\n")
	}
	if t.errMsg != "" {
		list.WriteString(dangerStyle.Render("  Error: "+t.errMsg) + "\n")
	}

	visibleH := l.ContentH - 3
	if visibleH < 1 {
		visibleH = 1
	}
	if t.cursor >= len(t.items) && len(t.items) > 0 {
		t.cursor = len(t.items) - 1
	}

	start := 0
	if t.cursor >= visibleH {
		start = t.cursor - visibleH + 1
	}
	end := start + visibleH
	if end > len(t.items) {
		end = len(t.items)
	}

	for i := start; i < end; i++ {
		it := t.items[i]
		prefix := "  "
		if i == t.cursor {
			prefix = selectedStyle.Render("> ")
		}

		globalRank := (t.page-1)*t.pageSize + i + 1
		rank := warningStyle.Render(fmt.Sprintf("#%-3d", globalRank))
		name := it.Name
		if i == t.cursor {
			name = selectedStyle.Render(name)
		} else {
			name = brightStyle.Render(name)
		}

		stars := dimStyle.Render(fmt.Sprintf("★%d", it.Stars))
		author := dimStyle.Render(it.Author)

		list.WriteString(fmt.Sprintf("%s%s %-28s %8s  %s\n", prefix, rank, name, stars, author))
	}

	if len(t.items) == 0 && !t.loading && t.errMsg == "" {
		list.WriteString(dimStyle.Render("  No trending data available.") + "\n")
	}

	leftPane := activePaneStyle.Width(l.LeftW).Height(l.ContentH).Render(list.String())

	// === Right pane: detail or confirmation modal ===
	var detail strings.Builder

	if t.confirming && len(t.items) > 0 && t.cursor < len(t.items) {
		// Confirmation modal
		it := t.items[t.cursor]
		detail.WriteString(warningStyle.Render("  Install Skill?") + "\n\n")

		detail.WriteString(brightStyle.Render("  "+it.Name) + "\n")
		detail.WriteString(dimStyle.Render("  by "+it.Author) + "\n\n")

		detail.WriteString(dimStyle.Render("  This will:") + "\n")
		detail.WriteString(brightStyle.Render("  1.") + " Download to central store\n")
		detail.WriteString(dimStyle.Render("     ~/.pskill/store/"+it.Name+"/") + "\n")
		detail.WriteString(brightStyle.Render("  2.") + " Symlink to project CLI dirs\n")
		for _, cli := range t.cfg.TargetCLIs {
			detail.WriteString(dimStyle.Render("     ."+cli+"/skills/"+it.Name) + "\n")
		}
		detail.WriteString(brightStyle.Render("  3.") + " Symlink to global CLI dirs\n")
		for _, cli := range t.cfg.TargetCLIs {
			detail.WriteString(dimStyle.Render("     ~/."+cli+"/skills/"+it.Name) + "\n")
		}
		detail.WriteString(brightStyle.Render("  4.") + " Update pskill.yaml\n")

		detail.WriteString("\n")

		confirmStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1E1E2E")).
			Background(ColorSuccess).
			Padding(0, 2).
			Bold(true)
		cancelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1E1E2E")).
			Background(ColorDanger).
			Padding(0, 2).
			Bold(true)

		detail.WriteString("  " + confirmStyle.Render("y/enter  Install") + "  " + cancelStyle.Render("n/esc  Cancel") + "\n")
	} else if len(t.items) > 0 && t.cursor < len(t.items) {
		// Normal detail view
		it := t.items[t.cursor]
		detail.WriteString(titleStyle.Render(it.Name) + "\n")
		detail.WriteString(dimStyle.Render("by ") + brightStyle.Render(it.Author) + "\n\n")

		detail.WriteString(it.Description + "\n\n")

		detail.WriteString(dimStyle.Render("Stars: ") + warningStyle.Render(fmt.Sprintf("★ %d", it.Stars)) + "\n")
		if it.GithubURL != "" {
			detail.WriteString(dimStyle.Render("GitHub: ") + dimStyle.Render(it.GithubURL) + "\n")
		}
		if it.SkillURL != "" {
			detail.WriteString(dimStyle.Render("URL: ") + dimStyle.Render(it.SkillURL) + "\n")
		}

		detail.WriteString("\n" + dimStyle.Render("Press enter to install"))
	} else {
		detail.WriteString(dimStyle.Render("No skill selected"))
	}

	rightPane := paneStyle.Width(l.RightW).Height(l.ContentH).Render(detail.String())

	if !l.HasDetail {
		return leftPane
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
}

func (t *TrendingTab) Title() string { return "Trending" }
func (t *TrendingTab) ShortHelp() []string {
	if t.confirming {
		return []string{
			helpEntry("y/enter", "install"),
			helpEntry("n/esc", "cancel"),
		}
	}
	return []string{
		helpEntry("j/k", "nav"),
		helpEntry("enter", "install"),
		helpEntry("</>", "page"),
		helpEntry("r", "refresh"),
	}
}
func (t *TrendingTab) AcceptsTextInput() bool { return false }

func (t *TrendingTab) loadCmd() tea.Cmd {
	t.loading = true
	page := t.page
	limit := t.pageSize
	return func() tea.Msg {
		client := registry.NewClient(t.cfg.RegistryURL, t.cfg.CacheDir, t.cfg.RegistryAPIKey)
		items, total, err := client.Trending(limit, page)
		return trendingMsg{items: items, total: total, err: err}
	}
}

func (t *TrendingTab) installCmd(item registry.SkillResult) tea.Cmd {
	cfg := t.cfg
	return func() tea.Msg {
		result, err := installer.InstallFromRegistryResult(cfg, item, true)
		return trendingInstallDoneMsg{result: result, err: err}
	}
}
