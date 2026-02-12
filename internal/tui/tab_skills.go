package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/installer"
	"github.com/ZiaoLiu-1/pskill/internal/skill"
	"github.com/ZiaoLiu-1/pskill/internal/tui/components"
)

type SkillsTab struct {
	cfg        config.Config
	state      ViewState
	filter     string
	cursor     int
	groupMode  int
	items      []skillEntry
	filtered   []skillEntry
	viewport   viewport.Model
	ready      bool
	projectSet map[string]bool // skills installed in current project
}

type skillEntry struct {
	Name string
	Desc string
	CLI  string
	Path string
}

func NewSkillsTab(cfg config.Config) Tab {
	return &SkillsTab{
		cfg:   cfg,
		state: StateList,
		items: []skillEntry{},
	}
}

func (t *SkillsTab) Init() tea.Cmd { return nil }

func (t *SkillsTab) Update(msg tea.Msg) (Tab, tea.Cmd) {
	var cmd tea.Cmd

	switch m := msg.(type) {
	case tea.KeyMsg:
		if t.state == StateSearch {
			switch m.String() {
			case "esc":
				t.state = StateList
				t.filter = ""
				t.updateFiltered()
			case "backspace":
				if len(t.filter) > 0 {
					t.filter = t.filter[:len(t.filter)-1]
					t.updateFiltered()
				}
			case "enter":
				t.state = StateList
			default:
				if len(m.String()) == 1 {
					t.filter += m.String()
					t.updateFiltered()
				}
			}
			t.cursor = 0
			return t, nil
		}

		switch m.String() {
		case "down":
			if t.cursor < len(t.filtered)-1 {
				t.cursor++
			} else if len(t.filtered) > 0 {
				t.cursor = len(t.filtered) - 1
			}
		case "up":
			if t.cursor > 0 {
				t.cursor--
			}
		case "/":
			t.state = StateSearch
			t.filter = ""
			t.updateFiltered()
		case "enter":
			if t.state == StateDetail {
				t.state = StateList
			} else {
				t.state = StateDetail
				t.updateViewport()
			}
		case "g":
			t.groupMode = (t.groupMode + 1) % 4
			t.updateFiltered() // re-sort
		case "esc":
			t.state = StateList
			t.filter = ""
			t.updateFiltered()
		}

		if t.state == StateDetail {
			var vpCmd tea.Cmd
			t.viewport, vpCmd = t.viewport.Update(msg)
			return t, vpCmd
		}

	case skillsScannedMsg:
		t.items = t.loadSkillEntries(m.names)
		t.projectSet, _ = installer.BuildStatusMap(t.cfg.StoreDir)
		t.updateFiltered()
	}
	return t, cmd
}

func (t *SkillsTab) View(width, height int) string {
	l := ComputeLayout(width, height, t.state == StateDetail)
	if l.IsTooSmall {
		return RenderTooSmall(width, height)
	}

	if !t.ready {
		t.viewport = viewport.New(l.RightW, l.ContentH)
		t.ready = true
	} else if t.state == StateDetail {
		t.viewport.Width = l.RightW
		t.viewport.Height = l.ContentH
	}

	var list strings.Builder

	if t.state == StateSearch {
		list.WriteString(selectedStyle.Render("Filter: ") + brightStyle.Render(t.filter+"_"))
	} else if t.filter != "" {
		list.WriteString(dimStyle.Render("Filter: "+t.filter))
	} else {
		list.WriteString(dimStyle.Render("/ to filter"))
	}
	list.WriteString(dimStyle.Render(fmt.Sprintf("  %s  %d/%d", t.groupLabel(), len(t.filtered), len(t.items))))
	list.WriteString("\n\n")

	if len(t.filtered) == 0 {
		list.WriteString(dimStyle.Render("  No skills found.\n"))
		list.WriteString(dimStyle.Render("  Run 'pskill scan' or press tab to go to Discover.\n"))
	}

	visibleHeight := l.ContentH - 2
	if visibleHeight < 1 {
		visibleHeight = 1
	}
	
	if t.cursor >= len(t.filtered) && len(t.filtered) > 0 {
		t.cursor = len(t.filtered) - 1
	}

	start := 0
	if t.cursor >= visibleHeight {
		start = t.cursor - visibleHeight + 1
	}
	end := start + visibleHeight
	if end > len(t.filtered) {
		end = len(t.filtered)
	}

	for i := start; i < end; i++ {
		entry := t.filtered[i]
		inProject := t.projectSet[entry.Name]

		prefix := "  "
		if i == t.cursor {
			prefix = selectedStyle.Render("> ")
		}

		indicator := "  "
		if inProject {
			indicator = successStyle.Render("✓ ")
		}

		badge := components.CLIBadge(entry.CLI)
		desc := entry.Desc
		if len(desc) > 30 {
			desc = desc[:27] + "..."
		}

		if inProject {
			// Green for project-installed skills
			gs := lipgloss.NewStyle().Foreground(ColorSuccess)
			if i == t.cursor {
				gs = gs.Bold(true)
			}
			list.WriteString(fmt.Sprintf("%s%s%s %-30s %s\n", prefix, indicator, badge, gs.Render(entry.Name), gs.Render(desc)))
		} else {
			name := entry.Name
			if i == t.cursor {
				name = selectedStyle.Render(name)
			} else {
				name = brightStyle.Render(name)
			}
			list.WriteString(fmt.Sprintf("%s%s%s %-30s %s\n", prefix, indicator, badge, name, dimStyle.Render(desc)))
		}
	}

	leftPane := activePaneStyle.Width(l.LeftW).Height(l.ContentH).Render(list.String())

	if !l.HasDetail {
		return leftPane
	}

	if t.state == StateDetail {
		rightPane := paneStyle.Width(l.RightW).Height(l.ContentH).Render(t.viewport.View())
		return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
	}

	var detail strings.Builder
	if len(t.filtered) > 0 {
		selected := t.filtered[t.cursor]
		detail.WriteString(titleStyle.Render("# "+selected.Name) + "\n\n")
		detail.WriteString(dimStyle.Render("CLI: ") + brightStyle.Render(selected.CLI) + "\n")
		detail.WriteString(dimStyle.Render("Path: ") + dimStyle.Render(selected.Path) + "\n\n")
		detail.WriteString(dimStyle.Render("Press Enter to view full detail"))
	} else {
		detail.WriteString(dimStyle.Render("No skill selected"))
	}
	
	rightPane := paneStyle.Width(l.RightW).Height(l.ContentH).Render(detail.String())
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
}

func (t *SkillsTab) updateFiltered() {
	q := strings.ToLower(t.filter)
	var filtered []skillEntry

	for _, s := range t.items {
		if strings.TrimSpace(t.filter) == "" || strings.Contains(strings.ToLower(s.Name), q) || strings.Contains(strings.ToLower(s.Desc), q) {
			filtered = append(filtered, s)
		}
	}

	switch t.groupMode {
	case 1:
		sort.Slice(filtered, func(i, j int) bool {
			if filtered[i].CLI != filtered[j].CLI {
				return filtered[i].CLI < filtered[j].CLI
			}
			return filtered[i].Name < filtered[j].Name
		})
	default:
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Name < filtered[j].Name
		})
	}

	t.filtered = filtered
	t.cursor = 0
}

func (t *SkillsTab) updateViewport() {
	if len(t.filtered) == 0 {
		t.viewport.SetContent("No skill selected")
		return
	}
	selected := t.filtered[t.cursor]
	
	var content strings.Builder
	content.WriteString(titleStyle.Render("# "+selected.Name) + "\n\n")
	content.WriteString(dimStyle.Render("CLI: ") + brightStyle.Render(selected.CLI) + "\n")
	content.WriteString(dimStyle.Render("Path: ") + dimStyle.Render(selected.Path) + "\n\n")

	mdPath := filepath.Join(t.cfg.StoreDir, selected.Name, "SKILL.md")
	if raw, err := os.ReadFile(mdPath); err == nil {
		body := string(raw)
		content.WriteString(components.RenderMarkdown(body, t.viewport.Width-4))
	} else {
		content.WriteString(dimStyle.Render("No SKILL.md preview available."))
	}
	
	t.viewport.SetContent(content.String())
	t.viewport.GotoTop()
}

func (t *SkillsTab) Title() string { return "My Skills" }
func (t *SkillsTab) ShortHelp() []string {
	if t.state == StateSearch {
		return []string{
			helpEntry("type", "filter"),
			helpEntry("esc", "cancel"),
			helpEntry("enter", "apply"),
		}
	}
	return []string{
		helpEntry("/", "filter"),
		helpEntry("↑/↓", "nav"),
		helpEntry("g", "group"),
		helpEntry("enter", "detail"),
	}
}

func (t *SkillsTab) AcceptsTextInput() bool {
	return t.state == StateSearch
}

func (t *SkillsTab) groupLabel() string {
	switch t.groupMode {
	case 1:
		return "by-cli"
	case 2:
		return "by-activity"
	case 3:
		return "by-recency"
	default:
		return "flat"
	}
}

func (t *SkillsTab) loadSkillEntries(names []string) []skillEntry {
	entries := make([]skillEntry, 0, len(names))
	for _, name := range names {
		entry := skillEntry{Name: name, CLI: "store"}
		mdPath := filepath.Join(t.cfg.StoreDir, name, "SKILL.md")
		if sk, err := skill.ParseFile(mdPath, ""); err == nil {
			entry.Desc = sk.Description
			entry.CLI = sk.SourceCLI
			entry.Path = sk.Path
			if entry.CLI == "" {
				entry.CLI = "store"
			}
		}
		entries = append(entries, entry)
	}
	return entries
}
