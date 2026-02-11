package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/registry"
	"github.com/ZiaoLiu-1/pskill/internal/tui/components"
)

type trendingMsg struct {
	items []registry.SkillResult
}

type TrendingTab struct {
	cfg        config.Config
	categories []string
	catIdx     int
	rangeMode  int
	items      []registry.SkillResult
	cursor     int
	loading    bool
}

func NewTrendingTab(cfg config.Config) Tab {
	return &TrendingTab{
		cfg:        cfg,
		categories: []string{"All", "Coding", "Writing", "DevOps", "Data", "Design"},
	}
}

func (t *TrendingTab) Init() tea.Cmd { return t.loadCmd() }

func (t *TrendingTab) Update(msg tea.Msg) (Tab, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "left", "h":
			if t.catIdx > 0 {
				t.catIdx--
				return t, t.loadCmd()
			}
		case "right", "l":
			if t.catIdx < len(t.categories)-1 {
				t.catIdx++
				return t, t.loadCmd()
			}
		case "r":
			t.rangeMode = (t.rangeMode + 1) % 3
			return t, t.loadCmd()
		case "j", "down":
			if t.cursor < len(t.items)-1 {
				t.cursor++
			}
		case "k", "up":
			if t.cursor > 0 {
				t.cursor--
			}
		case "i", "enter":
			if len(t.items) > 0 {
				// Install placeholder
				t.loading = true // trigger spinner or toast
			}
		}
	case trendingMsg:
		t.items = m.items
		t.loading = false
	}
	return t, nil
}

func (t *TrendingTab) View(width, height int) string {
	var b strings.Builder

	b.WriteString(brightStyle.Render("  Trending Feed") + "\n")
	b.WriteString("  CATEGORIES: " + components.CategoryPills(t.categories, t.catIdx) + "\n")
	b.WriteString("  RANGE: " + t.renderRange() + "\n")

	if t.loading {
		b.WriteString(warningStyle.Render("  Loading...") + "\n")
	}
	b.WriteString("\n")

	for i, it := range t.items {
		prefix := "  "
		if i == t.cursor {
			prefix = selectedStyle.Render("> ")
		}

		spark := components.Sparkline([]int64{2, 3, 5, 3, 7, 8, 9})
		rank := warningStyle.Render(fmt.Sprintf("#%d", i+1))
		name := brightStyle.Render(it.Name)
		if i == t.cursor {
			name = selectedStyle.Render(it.Name)
		}
		stats := dimStyle.Render(fmt.Sprintf("★%d ↓%d", it.Stars, it.Downloads))

		b.WriteString(fmt.Sprintf("%s%s %s  %s  %s\n", prefix, rank, name, stats, spark))
		b.WriteString(fmt.Sprintf("     %s\n", dimStyle.Render(it.Description)))
	}

	if len(t.items) == 0 && !t.loading {
		b.WriteString(dimStyle.Render("  No trending data available.") + "\n")
	}

	return activePaneStyle.Width(width - 4).Height(height - 2).Render(b.String())
}

func (t *TrendingTab) Title() string { return "Trending" }
func (t *TrendingTab) ShortHelp() []string {
	return []string{
		helpEntry("←/→", "category"),
		helpEntry("r", "range"),
		helpEntry("j/k", "nav"),
		helpEntry("i", "install"),
	}
}
func (t *TrendingTab) AcceptsTextInput() bool { return false }

func (t *TrendingTab) loadCmd() tea.Cmd {
	t.loading = true
	cat := strings.ToLower(t.categories[t.catIdx])
	rng := t.rangeSlug()
	return func() tea.Msg {
		client := registry.NewClient(t.cfg.RegistryURL, t.cfg.CacheDir)
		items, _ := client.Trending(15, rng, cat)
		return trendingMsg{items: items}
	}
}

func (t *TrendingTab) rangeSlug() string {
	switch t.rangeMode {
	case 1:
		return "month"
	case 2:
		return "all"
	default:
		return "week"
	}
}

func (t *TrendingTab) renderRange() string {
	labels := []string{"This Week", "This Month", "All Time"}
	parts := make([]string, 0, len(labels))
	for i, l := range labels {
		if i == t.rangeMode {
			parts = append(parts, selectedStyle.Render("["+l+"]"))
		} else {
			parts = append(parts, dimStyle.Render(l))
		}
	}
	return strings.Join(parts, "  ")
}
