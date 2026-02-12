package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/registry"
	"github.com/ZiaoLiu-1/pskill/internal/search"
	"github.com/ZiaoLiu-1/pskill/internal/tui/components"
)

type searchResultsMsg struct {
	Local  []search.Result
	Remote []registry.SkillResult
	err    error
}

type DiscoverTab struct {
	cfg        config.Config
	query      string
	typing     bool
	cursor     int
	searching  bool
	searchMode int // 0=keyword, 1=AI semantic
	local      []search.Result
	remote     []registry.SkillResult
	errMsg     string
}

func NewDiscoverTab(cfg config.Config) Tab {
	return &DiscoverTab{cfg: cfg, typing: false}
}

func (t *DiscoverTab) Init() tea.Cmd { return nil }

func (t *DiscoverTab) Update(msg tea.Msg) (Tab, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		if t.typing {
			switch m.String() {
			case "esc":
				t.typing = false
			case "backspace":
				if len(t.query) > 0 {
					t.query = t.query[:len(t.query)-1]
				}
			case "enter":
				t.typing = false
				if t.query != "" {
					return t, t.searchCmd()
				}
			case "ctrl+a":
				t.searchMode = (t.searchMode + 1) % 2
			default:
				if len(m.String()) == 1 {
					t.query += m.String()
				}
			}
		} else {
			switch m.String() {
			case "/":
				t.typing = true
			case "down":
				if t.cursor < len(t.allResults())-1 {
					t.cursor++
				}
			case "up":
				if t.cursor > 0 {
					t.cursor--
				}
			case "m":
				t.searchMode = (t.searchMode + 1) % 2
				if t.query != "" {
					return t, t.searchCmd()
				}
			}
		}

	case searchResultsMsg:
		t.searching = false
		if m.err != nil {
			t.errMsg = m.err.Error()
		} else {
			t.local = m.Local
			t.remote = m.Remote
			t.errMsg = ""
		}
		t.cursor = 0
	}
	return t, nil
}

func (t *DiscoverTab) View(width, height int) string {
	l := ComputeLayout(width, height, true)
	if l.IsTooSmall {
		return RenderTooSmall(width, height)
	}

	modeLabels := []string{"Keyword", "AI Semantic"}

	var list strings.Builder

	// Search bar
	if t.typing {
		list.WriteString(selectedStyle.Render("Search: ") + brightStyle.Render(t.query+"_"))
	} else {
		list.WriteString(dimStyle.Render("Search: ") + brightStyle.Render(t.query))
		if t.query == "" {
			list.WriteString(dimStyle.Render("  / to type"))
		}
	}
	list.WriteString("\n")

	// Mode indicator
	modeStyle := successStyle
	if t.searchMode == 1 {
		modeStyle = warningStyle
	}
	list.WriteString(dimStyle.Render("Mode: ") + modeStyle.Render(modeLabels[t.searchMode]))
	list.WriteString(dimStyle.Render("  (m to toggle)"))
	list.WriteString("\n")

	if t.searching {
		list.WriteString(warningStyle.Render("Searching...") + "\n")
	}
	if t.errMsg != "" {
		list.WriteString(dangerStyle.Render("Error: "+t.errMsg) + "\n")
	}
	list.WriteString("\n")

	results := t.allResults()
	if len(results) == 0 && t.query != "" && !t.searching {
		list.WriteString(dimStyle.Render("  No results. Try a different query.") + "\n")
	}
	if len(results) == 0 && t.query == "" {
		list.WriteString(dimStyle.Render("  Type a query to search skills.") + "\n")
		list.WriteString(dimStyle.Render("  Use AI mode for semantic search:") + "\n")
		list.WriteString(dimStyle.Render("  e.g. \"help me write better code reviews\"") + "\n")
	}

	for i, r := range results {
		prefix := "  "
		if i == t.cursor {
			prefix = selectedStyle.Render("> ")
		}

		badge := dimStyle.Render("R")
		if r.isLocal {
			badge = successStyle.Render("L")
		}

		name := r.name
		if i == t.cursor {
			name = selectedStyle.Render(name)
		} else {
			name = brightStyle.Render(name)
		}

		score := ""
		if r.score > 0 {
			score = dimStyle.Render(fmt.Sprintf("%.0f%%", r.score*100))
		}
		stars := ""
		if r.stars > 0 {
			stars = dimStyle.Render(fmt.Sprintf("★%d", r.stars))
		}

		list.WriteString(fmt.Sprintf("%s%s %-26s %6s %6s  %s\n", prefix, badge, name, score, stars, dimStyle.Render(r.author)))
	}

	leftPane := activePaneStyle.Width(l.LeftW).Height(l.ContentH).Render(list.String())

	// Preview pane
	var preview strings.Builder
	preview.WriteString(titleStyle.Render("Preview") + "\n\n")
	if len(results) > 0 && t.cursor < len(results) {
		sel := results[t.cursor]
		preview.WriteString(brightStyle.Render(sel.name) + "\n")
		if sel.author != "" {
			preview.WriteString(dimStyle.Render("by ") + brightStyle.Render(sel.author) + "\n")
		}
		preview.WriteString("\n")
		preview.WriteString(sel.desc + "\n\n")
		if sel.isLocal {
			preview.WriteString(successStyle.Render("Installed locally") + "\n")
		} else {
			preview.WriteString(dimStyle.Render("Available on registry") + "\n")
		}
		if sel.githubURL != "" {
			preview.WriteString("\n" + dimStyle.Render("GitHub: "+sel.githubURL))
		}
	} else {
		md := "Select a result to preview."
		preview.WriteString(components.RenderMarkdown(md, l.RightW-4))
	}

	rightPane := paneStyle.Width(l.RightW).Height(l.ContentH).Render(preview.String())

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
}

func (t *DiscoverTab) Title() string { return "Discover" }
func (t *DiscoverTab) ShortHelp() []string {
	if t.typing {
		return []string{
			helpEntry("type", "search"),
			helpEntry("ctrl+a", "mode"),
			helpEntry("esc", "stop"),
			helpEntry("enter", "go"),
		}
	}
	return []string{
		helpEntry("/", "search"),
		helpEntry("m", "mode"),
		helpEntry("↑/↓", "nav"),
	}
}

func (t *DiscoverTab) AcceptsTextInput() bool { return t.typing }

type discoveryResult struct {
	name      string
	desc      string
	author    string
	score     float64
	stars     int64
	githubURL string
	isLocal   bool
}

func (t *DiscoverTab) allResults() []discoveryResult {
	out := make([]discoveryResult, 0, len(t.local)+len(t.remote))
	for _, l := range t.local {
		out = append(out, discoveryResult{name: l.Name, desc: l.Description, score: l.Score, isLocal: true})
	}
	for _, r := range t.remote {
		out = append(out, discoveryResult{
			name:      r.Name,
			desc:      r.Description,
			author:    r.Author,
			score:     r.Score,
			stars:     r.Stars,
			githubURL: r.GithubURL,
			isLocal:   false,
		})
	}
	return out
}

func (t *DiscoverTab) searchCmd() tea.Cmd {
	t.searching = true
	query := t.query
	mode := t.searchMode
	return func() tea.Msg {
		// Always search local index
		engine := search.NewEngine(t.cfg.IndexDir)
		local, _ := engine.Search(query, 5)

		client := registry.NewClient(t.cfg.RegistryURL, t.cfg.CacheDir, t.cfg.RegistryAPIKey)
		var remote []registry.SkillResult
		var err error

		if mode == 1 {
			// AI semantic search
			remote, err = client.AISearch(query)
		} else {
			// Keyword search
			remote, _, err = client.Search(query, 15, 1, "stars")
		}

		return searchResultsMsg{Local: local, Remote: remote, err: err}
	}
}
