package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/registry"
)

type trendingMsg struct {
	items []registry.SkillResult
	total int
	err   error
}

type TrendingTab struct {
	cfg     config.Config
	items   []registry.SkillResult
	total   int
	cursor  int
	loading bool
	errMsg  string
}

func NewTrendingTab(cfg config.Config) Tab {
	return &TrendingTab{cfg: cfg}
}

func (t *TrendingTab) Init() tea.Cmd { return t.loadCmd() }

func (t *TrendingTab) Update(msg tea.Msg) (Tab, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
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
	list.WriteString(dimStyle.Render(fmt.Sprintf("  sorted by ★ stars  (%d total)", t.total)))
	list.WriteString("\n\n")

	if t.loading {
		list.WriteString(warningStyle.Render("  Loading from skillsmp.com...") + "\n")
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

		rank := warningStyle.Render(fmt.Sprintf("#%-3d", i+1))
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

	// === Right pane: detail ===
	var detail strings.Builder
	if len(t.items) > 0 && t.cursor < len(t.items) {
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
	return []string{
		helpEntry("j/k", "nav"),
		helpEntry("r", "refresh"),
	}
}
func (t *TrendingTab) AcceptsTextInput() bool { return false }

func (t *TrendingTab) loadCmd() tea.Cmd {
	t.loading = true
	return func() tea.Msg {
		client := registry.NewClient(t.cfg.RegistryURL, t.cfg.CacheDir, t.cfg.RegistryAPIKey)
		items, total, err := client.Trending(20, 1)
		return trendingMsg{items: items, total: total, err: err}
	}
}
