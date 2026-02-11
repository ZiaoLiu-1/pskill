package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/skill"
	"github.com/ZiaoLiu-1/pskill/internal/tui/components"
)

type SkillsTab struct {
	cfg       config.Config
	state     ViewState
	filter    string
	cursor    int
	groupMode int
	items     []skillEntry
}

type skillEntry struct {
	Name string
	Desc string
	CLI  string
	Path string
}

func NewSkillsTab(cfg config.Config) Tab {
	return &SkillsTab{cfg: cfg, state: StateList, items: []skillEntry{}}
}

func (t *SkillsTab) Init() tea.Cmd { return nil }

func (t *SkillsTab) Update(msg tea.Msg) (Tab, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		if t.state == StateSearch {
			switch m.String() {
			case "esc":
				t.state = StateList
				t.filter = ""
			case "backspace":
				if len(t.filter) > 0 {
					t.filter = t.filter[:len(t.filter)-1]
				}
			case "enter":
				t.state = StateList
			default:
				if len(m.String()) == 1 {
					t.filter += m.String()
				}
			}
			t.cursor = 0
			return t, nil
		}

		switch m.String() {
		case "j", "down":
			if t.cursor < len(t.filtered())-1 {
				t.cursor++
			}
		case "k", "up":
			if t.cursor > 0 {
				t.cursor--
			}
		case "/":
			t.state = StateSearch
			t.filter = ""
		case "enter":
			if t.state == StateDetail {
				t.state = StateList
			} else {
				t.state = StateDetail
			}
		case "g":
			t.groupMode = (t.groupMode + 1) % 4
		case "esc":
			t.state = StateList
			t.filter = ""
		}

	case skillsScannedMsg:
		t.items = t.loadSkillEntries(m.names)
	}
	return t, nil
}

func (t *SkillsTab) View(width, height int) string {
	items := t.filtered()
	if t.cursor >= len(items) {
		if len(items) > 0 {
			t.cursor = len(items) - 1
		} else {
			t.cursor = 0
		}
	}

	hasDetail := t.state == StateDetail && len(items) > 0
	leftW := width - 4
	rightW := 0
	if hasDetail {
		leftW = width/2 - 2
		rightW = width - leftW - 4
	}

	// Build list pane
	var list strings.Builder

	// Filter bar
	if t.state == StateSearch {
		list.WriteString(selectedStyle.Render("Filter: ") + brightStyle.Render(t.filter+"_"))
	} else if t.filter != "" {
		list.WriteString(dimStyle.Render("Filter: "+t.filter))
	} else {
		list.WriteString(dimStyle.Render("/ to filter"))
	}
	list.WriteString(dimStyle.Render(fmt.Sprintf("  %s  %d/%d", t.groupLabel(), len(items), len(t.items))))
	list.WriteString("\n\n")

	if len(items) == 0 {
		list.WriteString(dimStyle.Render("  No skills found.\n"))
		list.WriteString(dimStyle.Render("  Run 'pskill scan' or press tab to go to Discover.\n"))
	}

	visible := height - 4
	if visible < 1 {
		visible = 1
	}
	start := 0
	if t.cursor >= visible {
		start = t.cursor - visible + 1
	}

	for i := start; i < len(items) && i < start+visible; i++ {
		entry := items[i]
		prefix := "  "
		if i == t.cursor {
			prefix = selectedStyle.Render("> ")
		}

		badge := components.CLIBadge(entry.CLI)
		name := entry.Name
		desc := entry.Desc
		if len(desc) > 30 {
			desc = desc[:27] + "..."
		}

		if i == t.cursor {
			name = selectedStyle.Render(name)
		} else {
			name = brightStyle.Render(name)
		}

		list.WriteString(fmt.Sprintf("%s%s %-30s %s %s\n", prefix, badge, name, dimStyle.Render(desc), ""))
	}

	leftPane := activePaneStyle.Width(leftW).Height(height - 2).Render(list.String())

	if !hasDetail {
		return leftPane
	}

	// Detail pane
	selected := items[t.cursor]
	var detail strings.Builder
	detail.WriteString(titleStyle.Render("# "+selected.Name) + "\n\n")
	detail.WriteString(dimStyle.Render("CLI: ") + brightStyle.Render(selected.CLI) + "\n")
	detail.WriteString(dimStyle.Render("Path: ") + dimStyle.Render(selected.Path) + "\n\n")

	// Try reading actual SKILL.md content
	mdPath := filepath.Join(t.cfg.StoreDir, selected.Name, "SKILL.md")
	if raw, err := os.ReadFile(mdPath); err == nil {
		body := string(raw)
		if len(body) > 500 {
			body = body[:500] + "\n..."
		}
		detail.WriteString(components.RenderMarkdown(body, rightW-4))
	} else {
		detail.WriteString(dimStyle.Render("No SKILL.md preview available."))
	}

	rightPane := paneStyle.Width(rightW).Height(height - 2).Render(detail.String())

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
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
		helpEntry("j/k", "nav"),
		helpEntry("g", "group"),
		helpEntry("enter", "detail"),
	}
}

func (t *SkillsTab) AcceptsTextInput() bool {
	return t.state == StateSearch
}

func (t *SkillsTab) filtered() []skillEntry {
	if strings.TrimSpace(t.filter) == "" {
		return t.items
	}
	q := strings.ToLower(t.filter)
	out := make([]skillEntry, 0, len(t.items))
	for _, s := range t.items {
		if strings.Contains(strings.ToLower(s.Name), q) || strings.Contains(strings.ToLower(s.Desc), q) {
			out = append(out, s)
		}
	}
	return out
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
