package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/monitor"
	"github.com/ZiaoLiu-1/pskill/internal/tui/components"
)

type monitorMsg struct {
	stats monitor.Aggregates
}

type MonitorTab struct {
	cfg   config.Config
	stats monitor.Aggregates
}

func NewMonitorTab(cfg config.Config) Tab {
	return &MonitorTab{
		cfg: cfg,
		stats: monitor.Aggregates{
			TopSkills: map[string]int64{},
			ByCLI:     map[string]int64{},
			Recent:    []monitor.Event{},
			Stale:     []string{},
		},
	}
}

func (t *MonitorTab) Init() tea.Cmd { return t.loadCmd() }

func (t *MonitorTab) Update(msg tea.Msg) (Tab, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "r":
			return t, t.loadCmd()
		case "d":
			// Delete stale placeholder
			if len(t.stats.Stale) > 0 {
				t.stats.Stale = []string{} // clear for now
			}
		}
	case monitorMsg:
		t.stats = m.stats
	}
	return t, nil
}

func (t *MonitorTab) View(width, height int) string {
	halfW := (width - 6) / 2
	halfH := (height - 4) / 2
	if halfW < 10 {
		halfW = 10
	}
	if halfH < 3 {
		halfH = 3
	}

	topLeft := t.renderTopSkills(halfW, halfH)
	topRight := t.renderCLIBreakdown(halfW, halfH)
	bottomLeft := t.renderRecent(halfW, halfH)
	bottomRight := t.renderStale(halfW, halfH)

	top := lipgloss.JoinHorizontal(lipgloss.Top, topLeft, " ", topRight)
	bottom := lipgloss.JoinHorizontal(lipgloss.Top, bottomLeft, " ", bottomRight)

	return lipgloss.JoinVertical(lipgloss.Left, top, bottom)
}

func (t *MonitorTab) Title() string { return "Monitor" }
func (t *MonitorTab) ShortHelp() []string {
	return []string{
		helpEntry("r", "refresh"),
		helpEntry("d", "delete stale"),
	}
}
func (t *MonitorTab) AcceptsTextInput() bool { return false }

func (t *MonitorTab) renderTopSkills(w, h int) string {
	title := titleStyle.Render("TOP SKILLS") + "\n\n"
	type pair struct {
		name string
		n    int64
	}
	pairs := make([]pair, 0, len(t.stats.TopSkills))
	for k, v := range t.stats.TopSkills {
		pairs = append(pairs, pair{name: k, n: v})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].n > pairs[j].n })

	var b strings.Builder
	for _, p := range pairs {
		spark := components.Sparkline([]int64{1, 3, p.n / 2, p.n})
		b.WriteString(fmt.Sprintf("%-18s [%3d] %s\n", p.name, p.n, spark))
	}
	if len(pairs) == 0 {
		b.WriteString(dimStyle.Render("No usage data yet.\nUse skills to see stats here."))
	}

	return paneStyle.Width(w).Height(h).Render(title + b.String())
}

func (t *MonitorTab) renderCLIBreakdown(w, h int) string {
	title := titleStyle.Render("CLI BREAKDOWN") + "\n\n"
	total := int64(0)
	for _, n := range t.stats.ByCLI {
		total += n
	}
	var b strings.Builder
	if total == 0 {
		b.WriteString(dimStyle.Render("No CLI usage yet."))
	} else {
		for cli, n := range t.stats.ByCLI {
			pct := (100 * n) / total
			bar := strings.Repeat("█", int(pct/5)) + strings.Repeat("░", 20-int(pct/5))
			b.WriteString(fmt.Sprintf("%-8s %3d%% %s\n", cli, pct, bar))
		}
	}

	return paneStyle.Width(w).Height(h).Render(title + b.String())
}

func (t *MonitorTab) renderRecent(w, h int) string {
	title := titleStyle.Render("RECENT ACTIVITY") + "\n\n"
	var b strings.Builder
	if len(t.stats.Recent) == 0 {
		b.WriteString(dimStyle.Render("No recent activity."))
	} else {
		for i, ev := range t.stats.Recent {
			if i > 6 {
				break
			}
			b.WriteString(fmt.Sprintf("%s  %s (%s)\n",
				dimStyle.Render(ev.Timestamp.Format("15:04")),
				brightStyle.Render(ev.SkillName),
				dimStyle.Render(ev.CLI),
			))
		}
	}

	return paneStyle.Width(w).Height(h).Render(title + b.String())
}

func (t *MonitorTab) renderStale(w, h int) string {
	title := titleStyle.Render("STALE SKILLS") + "\n\n"
	var b strings.Builder
	if len(t.stats.Stale) == 0 {
		b.WriteString(dimStyle.Render("No stale skills. All good!"))
	} else {
		for _, s := range t.stats.Stale {
			b.WriteString(warningStyle.Render("  "+s) + "\n")
		}
	}

	return paneStyle.Width(w).Height(h).Render(title + b.String())
}

func (t *MonitorTab) loadCmd() tea.Cmd {
	return func() tea.Msg {
		tr, err := monitor.NewTracker(t.cfg.StatsDB)
		if err != nil {
			return monitorMsg{stats: monitor.Aggregates{TopSkills: map[string]int64{}, ByCLI: map[string]int64{}}}
		}
		defer tr.Close()
		st, _ := tr.Stats()
		return monitorMsg{stats: st}
	}
}
