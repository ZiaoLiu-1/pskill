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

// --- messages ---

type trendingMsg struct {
	items []registry.SkillResult
	total int
	err   error
}

type trendingInstallDoneMsg struct {
	result *installer.Result
	err    error
}

type trendingUninstallDoneMsg struct {
	name string
	err  error
}

// --- tab state ---

type trendingState int

const (
	trendingBrowse   trendingState = iota // normal list browsing
	trendingConfirm                       // confirmation dialog
	trendingWorking                       // install/uninstall in progress
)

type TrendingTab struct {
	cfg        config.Config
	items      []registry.SkillResult
	total      int
	cursor     int
	page       int
	pageSize   int
	loading    bool
	state      trendingState
	errMsg     string
	projectSet map[string]bool // skills installed in current project
	storeSet   map[string]bool // skills in central store
}

func NewTrendingTab(cfg config.Config) Tab {
	return &TrendingTab{cfg: cfg, page: 1, pageSize: 20}
}

func (t *TrendingTab) Init() tea.Cmd {
	t.refreshStatus()
	return t.loadCmd()
}

func (t *TrendingTab) refreshStatus() {
	t.projectSet, t.storeSet = installer.BuildStatusMap(t.cfg.StoreDir)
}

func (t *TrendingTab) Update(msg tea.Msg) (Tab, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch t.state {

		case trendingConfirm:
			switch m.String() {
			case "y", "Y", "enter":
				t.state = trendingWorking
				return t, t.installCmd(t.items[t.cursor])
			case "n", "N", "esc", "q":
				t.state = trendingBrowse
			}
			return t, nil

		case trendingWorking:
			// Block all keys except quit while working
			return t, nil

		default: // trendingBrowse
			switch m.String() {
			case "down":
				if t.cursor < len(t.items)-1 {
					t.cursor++
				}
			case "up":
				if t.cursor > 0 {
					t.cursor--
				}
			case "r":
				return t, t.loadCmd()
			case "right":
				if t.page*t.pageSize < t.total {
					t.page++
					t.cursor = 0
					return t, t.loadCmd()
				}
			case "left":
				if t.page > 1 {
					t.page--
					t.cursor = 0
					return t, t.loadCmd()
				}
			case "enter":
				if len(t.items) > 0 && t.cursor < len(t.items) {
					t.state = trendingConfirm
				}
			case "x":
				if len(t.items) > 0 && t.cursor < len(t.items) {
					t.state = trendingWorking
					return t, t.uninstallCmd(t.items[t.cursor].Name)
				}
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
		t.state = trendingBrowse
		t.refreshStatus()
		if m.err != nil {
			t.errMsg = m.err.Error()
			return t, func() tea.Msg {
				return toastMsg{text: "Failed: " + m.err.Error(), duration: 3 * time.Second}
			}
		}
		t.errMsg = ""
		linked := strings.Join(m.result.LinkedCLIs, ", ")
		return t, tea.Batch(
			func() tea.Msg { return statusMsg{text: "Installed " + m.result.SkillName} },
			func() tea.Msg {
				return toastMsg{text: "Installed " + m.result.SkillName + " → " + linked, duration: 3 * time.Second}
			},
		)

	case trendingUninstallDoneMsg:
		t.state = trendingBrowse
		t.refreshStatus()
		if m.err != nil {
			t.errMsg = m.err.Error()
			return t, func() tea.Msg {
				return toastMsg{text: "Uninstall failed: " + m.err.Error(), duration: 3 * time.Second}
			}
		}
		t.errMsg = ""
		return t, tea.Batch(
			func() tea.Msg { return statusMsg{text: "Uninstalled " + m.name + " from project"} },
			func() tea.Msg {
				return toastMsg{text: "Removed " + m.name + " from project", duration: 3 * time.Second}
			},
		)
	}
	return t, nil
}

func (t *TrendingTab) View(width, height int) string {
	l := ComputeLayout(width, height, true)
	if l.IsTooSmall {
		return RenderTooSmall(width, height)
	}

	leftPane := t.renderList(l)
	rightPane := t.renderDetail(l)

	if !l.HasDetail {
		return leftPane
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
}

// --- left pane: skill list ---

func (t *TrendingTab) renderList(l Layout) string {
	var b strings.Builder
	totalPages := (t.total + t.pageSize - 1) / t.pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	b.WriteString(brightStyle.Render("Trending Skills"))
	b.WriteString(dimStyle.Render(fmt.Sprintf("  ★ stars  %d total  page %d/%d", t.total, t.page, totalPages)))
	b.WriteString("\n\n")

	if t.loading {
		b.WriteString(warningStyle.Render("  Loading...") + "\n")
	}
	if t.state == trendingWorking {
		b.WriteString(warningStyle.Render("  Working...") + "\n")
	}
	if t.errMsg != "" {
		b.WriteString(dangerStyle.Render("  "+t.errMsg) + "\n")
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

		// Determine install status
		inProject := t.projectSet[it.Name]
		inStore := t.storeSet[it.Name]

		// Status indicator
		indicator := "  "
		if inProject {
			indicator = successStyle.Render("✓ ")
		} else if inStore {
			indicator = warningStyle.Render("● ")
		}

		prefix := " "
		if i == t.cursor {
			prefix = selectedStyle.Render(">")
		}

		globalRank := (t.page-1)*t.pageSize + i + 1
		rankStr := fmt.Sprintf("#%-3d", globalRank)
		nameStr := it.Name
		starsStr := fmt.Sprintf("★%d", it.Stars)
		authorStr := it.Author

		// Apply color based on status
		if inProject {
			// Green for project-installed
			gs := lipgloss.NewStyle().Foreground(ColorSuccess)
			if i == t.cursor {
				gs = gs.Bold(true)
			}
			b.WriteString(fmt.Sprintf("%s%s%s %-28s %8s  %s\n",
				prefix, indicator,
				gs.Render(rankStr), gs.Render(nameStr), gs.Render(starsStr), gs.Render(authorStr)))
		} else if inStore {
			// Yellow for globally installed but not in project
			ys := lipgloss.NewStyle().Foreground(ColorWarning)
			if i == t.cursor {
				ys = ys.Bold(true)
			}
			b.WriteString(fmt.Sprintf("%s%s%s %-28s %8s  %s\n",
				prefix, indicator,
				ys.Render(rankStr), ys.Render(nameStr), ys.Render(starsStr), ys.Render(authorStr)))
		} else {
			// Normal styling
			rank := warningStyle.Render(rankStr)
			name := brightStyle.Render(nameStr)
			if i == t.cursor {
				name = selectedStyle.Render(nameStr)
			}
			stars := dimStyle.Render(starsStr)
			author := dimStyle.Render(authorStr)
			b.WriteString(fmt.Sprintf("%s%s%s %-28s %8s  %s\n", prefix, indicator, rank, name, stars, author))
		}
	}

	if len(t.items) == 0 && !t.loading && t.errMsg == "" {
		b.WriteString(dimStyle.Render("  No trending data.") + "\n")
	}

	return activePaneStyle.Width(l.LeftW).Height(l.ContentH).Render(b.String())
}

// --- right pane: detail or confirmation ---

func (t *TrendingTab) renderDetail(l Layout) string {
	if len(t.items) == 0 || t.cursor >= len(t.items) {
		return paneStyle.Width(l.RightW).Height(l.ContentH).Render(
			dimStyle.Render("No skill selected"),
		)
	}

	it := t.items[t.cursor]

	if t.state == trendingConfirm {
		return t.renderConfirm(l, it)
	}

	return t.renderSkillDetail(l, it)
}

func (t *TrendingTab) renderSkillDetail(l Layout, it registry.SkillResult) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(it.Name) + "\n")
	b.WriteString(dimStyle.Render("by ") + brightStyle.Render(it.Author) + "\n\n")

	// Wrap description to fit pane width
	desc := it.Description
	maxW := l.RightW - 6
	if maxW > 10 && len(desc) > maxW {
		b.WriteString(wordWrap(desc, maxW))
	} else {
		b.WriteString(desc)
	}
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render("Stars: ") + warningStyle.Render(fmt.Sprintf("★ %d", it.Stars)) + "\n\n")

	b.WriteString(successStyle.Render("enter") + dimStyle.Render(" install to project") + "\n")
	b.WriteString(dangerStyle.Render("x") + dimStyle.Render("     uninstall from project") + "\n")

	return paneStyle.Width(l.RightW).Height(l.ContentH).Render(b.String())
}

func (t *TrendingTab) renderConfirm(l Layout, it registry.SkillResult) string {
	w := l.RightW - 6
	if w < 20 {
		w = 20
	}

	sep := dimStyle.Render(strings.Repeat("─", w))

	var b strings.Builder

	// Title
	b.WriteString("\n")
	b.WriteString(warningStyle.Render("  INSTALL SKILL") + "\n")
	b.WriteString("  " + sep + "\n\n")

	// Skill info
	b.WriteString("  " + brightStyle.Render(it.Name))
	b.WriteString(dimStyle.Render("  by " + it.Author) + "\n\n")

	// What will happen
	b.WriteString("  " + dimStyle.Render("Store") + "\n")
	b.WriteString("    " + dimStyle.Render("~/.pskill/store/"+it.Name+"/") + "\n\n")

	b.WriteString("  " + dimStyle.Render("Project symlinks") + "\n")
	for _, cli := range t.cfg.TargetCLIs {
		b.WriteString("    " + successStyle.Render("→") + " " + dimStyle.Render("."+cli+"/skills/"+it.Name) + "\n")
	}
	b.WriteString("\n")

	b.WriteString("  " + dimStyle.Render("Global symlinks") + "\n")
	for _, cli := range t.cfg.TargetCLIs {
		b.WriteString("    " + successStyle.Render("→") + " " + dimStyle.Render("~/."+cli+"/skills/"+it.Name) + "\n")
	}

	b.WriteString("\n  " + sep + "\n\n")

	// Buttons
	yBtn := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1E1E2E")).
		Background(ColorSuccess).
		Padding(0, 3).
		Bold(true).
		Render("y  Install")

	nBtn := lipgloss.NewStyle().
		Foreground(ColorBright).
		Background(ColorDim).
		Padding(0, 3).
		Render("esc  Cancel")

	b.WriteString("  " + yBtn + "  " + nBtn + "\n")

	return paneStyle.Width(l.RightW).Height(l.ContentH).Render(b.String())
}

func (t *TrendingTab) Title() string { return "Trending" }

func (t *TrendingTab) ShortHelp() []string {
	if t.state == trendingConfirm {
		return []string{
			helpEntry("y/enter", "install"),
			helpEntry("esc", "cancel"),
		}
	}
	return []string{
		helpEntry("↑/↓", "nav"),
		helpEntry("enter", "install"),
		helpEntry("x", "uninstall"),
		helpEntry("←/→", "page"),
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

func (t *TrendingTab) uninstallCmd(name string) tea.Cmd {
	cfg := t.cfg
	return func() tea.Msg {
		err := installer.UninstallFromProject(cfg, name)
		return trendingUninstallDoneMsg{name: name, err: err}
	}
}

// wordWrap breaks text into lines of maxWidth characters at word boundaries.
func wordWrap(text string, maxWidth int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	var lines []string
	line := words[0]
	for _, w := range words[1:] {
		if len(line)+1+len(w) > maxWidth {
			lines = append(lines, line)
			line = w
		} else {
			line += " " + w
		}
	}
	lines = append(lines, line)
	return strings.Join(lines, "\n")
}
