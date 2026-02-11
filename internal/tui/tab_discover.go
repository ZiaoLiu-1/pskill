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
}

type DiscoverTab struct {
	cfg       config.Config
	query     string
	typing    bool
	cursor    int
	searching bool
	sortMode  int
	local     []search.Result
	remote    []registry.SkillResult
	status    string
}

func NewDiscoverTab(cfg config.Config) Tab {
	return &DiscoverTab{cfg: cfg, typing: true}
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
					return t, t.searchCmd()
				}
			case "enter":
				t.typing = false
				return t, t.searchCmd()
			default:
				if len(m.String()) == 1 {
					t.query += m.String()
					return t, t.searchCmd()
				}
			}
		} else {
			switch m.String() {
			case "/":
				t.typing = true
			case "j", "down":
				if t.cursor < len(t.allResults())-1 {
					t.cursor++
				}
			case "k", "up":
				if t.cursor > 0 {
					t.cursor--
				}
			case "s":
				t.sortMode = (t.sortMode + 1) % 3
			case "enter":
				// Install action placeholder
				t.status = "Install triggered (placeholder)"
			}
		}

	case searchResultsMsg:
		t.local = m.Local
		t.remote = m.Remote
		t.searching = false
		t.cursor = 0
	}
	return t, nil
}

func (t *DiscoverTab) View(width, height int) string {
	l := ComputeLayout(width, height, true)
	if l.IsTooSmall {
		return RenderTooSmall(width, height)
	}

	sortNames := []string{"relevance", "trending", "newest"}

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

	if t.searching {
		list.WriteString(warningStyle.Render("Searching...") + "\n")
	}

	list.WriteString(dimStyle.Render(fmt.Sprintf("Sort: %s", sortNames[t.sortMode])))
	list.WriteString("\n\n")

	results := t.allResults()
	if len(results) == 0 && t.query != "" && !t.searching {
		list.WriteString(dimStyle.Render("  No results. Try a different query.") + "\n")
	}
	if len(results) == 0 && t.query == "" {
		list.WriteString(dimStyle.Render("  Type a query to search skills by meaning.") + "\n")
		list.WriteString(dimStyle.Render("  e.g. \"help me write better code reviews\"") + "\n")
	}

	for i, r := range results {
		prefix := "  "
		if i == t.cursor {
			prefix = selectedStyle.Render("> ")
		}

		badge := dimStyle.Render("(Remote)")
		if r.isLocal {
			badge = successStyle.Render("(Local)")
		}

		name := r.name
		if i == t.cursor {
			name = selectedStyle.Render(name)
		} else {
			name = brightStyle.Render(name)
		}

		score := dimStyle.Render(fmt.Sprintf("%.0f%%", r.score*100))
		list.WriteString(fmt.Sprintf("%s%s %s  %s\n", prefix, badge, name, score))
	}

	leftPane := activePaneStyle.Width(l.LeftW).Height(l.ContentH).Render(list.String())

	// Preview pane
	var preview strings.Builder
	preview.WriteString(titleStyle.Render("Preview") + "\n\n")
	if len(results) > 0 && t.cursor < len(results) {
		sel := results[t.cursor]
		preview.WriteString(brightStyle.Render(sel.name) + "\n")
		preview.WriteString(dimStyle.Render(sel.desc) + "\n\n")
		if sel.isLocal {
			preview.WriteString(successStyle.Render("Installed locally") + "\n")
		} else {
			preview.WriteString(dimStyle.Render("Available on registry") + "\n")
		}
		preview.WriteString("\n" + dimStyle.Render("Press enter to install/view"))
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
			helpEntry("esc", "stop typing"),
			helpEntry("enter", "search"),
		}
	}
	return []string{
		helpEntry("/", "search"),
		helpEntry("j/k", "nav"),
		helpEntry("s", "sort"),
		helpEntry("enter", "install"),
	}
}

func (t *DiscoverTab) AcceptsTextInput() bool { return t.typing }

type discoveryResult struct {
	name    string
	desc    string
	score   float64
	isLocal bool
}

func (t *DiscoverTab) allResults() []discoveryResult {
	out := make([]discoveryResult, 0, len(t.local)+len(t.remote))
	for _, l := range t.local {
		out = append(out, discoveryResult{name: l.Name, desc: l.Description, score: l.Score, isLocal: true})
	}
	for _, r := range t.remote {
		out = append(out, discoveryResult{name: r.Name, desc: r.Description, score: r.Score, isLocal: false})
	}
	return out
}

func (t *DiscoverTab) searchCmd() tea.Cmd {
	t.searching = true
	query := t.query
	return func() tea.Msg {
		engine := search.NewEngine(t.cfg.IndexDir)
		local, _ := engine.Search(query, 8)
		client := registry.NewClient(t.cfg.RegistryURL, t.cfg.CacheDir)
		remote, _ := client.Search(query, 8)
		return searchResultsMsg{Local: local, Remote: remote}
	}
}
