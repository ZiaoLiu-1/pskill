package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/project"
)

// projectsDiscoveredMsg is sent when background project discovery completes.
type projectsDiscoveredMsg struct {
	projects []project.Info
}

type ProjectsTab struct {
	cfg      config.Config
	projects []project.Info
	cursor   int
	loading  bool
	ready    bool
}

func NewProjectsTab(cfg config.Config) Tab {
	return &ProjectsTab{cfg: cfg}
}

func (t *ProjectsTab) Init() tea.Cmd {
	t.loading = true
	return discoverProjectsCmd
}

func discoverProjectsCmd() tea.Msg {
	home, _ := os.UserHomeDir()
	// Scan common code directories with limited depth
	roots := []string{
		"~/Desktop",
		"~/Documents",
		"~/Projects",
		"~/projects",
		"~/Code",
		"~/code",
		"~/dev",
		"~/Dev",
		"~/src",
		"~/workspace",
		"~/Workspace",
		"~/repos",
		"~/Repos",
		"~/go/src",
		home, // top-level home (depth 1 only catches direct children)
	}
	results := project.Discover(roots, 3)
	return projectsDiscoveredMsg{projects: results}
}

func (t *ProjectsTab) Update(msg tea.Msg) (Tab, tea.Cmd) {
	switch m := msg.(type) {
	case projectsDiscoveredMsg:
		t.projects = m.projects
		t.loading = false
		t.ready = true
		t.cursor = 0
		// Auto-select current project if found
		for i, p := range t.projects {
			if p.IsCurrent {
				t.cursor = i
				break
			}
		}
	case tea.KeyMsg:
		switch m.String() {
		case "down":
			if t.cursor < len(t.projects)-1 {
				t.cursor++
			}
		case "up":
			if t.cursor > 0 {
				t.cursor--
			}
		case "r":
			t.loading = true
			return t, discoverProjectsCmd
		}
	}
	return t, nil
}

func (t *ProjectsTab) View(width, height int) string {
	l := ComputeLayout(width, height, true)
	if l.IsTooSmall {
		return RenderTooSmall(width, height)
	}

	if t.loading {
		return paneStyle.Width(width - 2).Height(l.ContentH).Render(
			"\n" + dimStyle.Render("  Scanning for projects..."),
		)
	}

	// === Left pane: project list ===
	var list strings.Builder

	cwd, _ := os.Getwd()
	home, _ := os.UserHomeDir()
	cwdDisplay := shortenPath(cwd, home)
	list.WriteString(dimStyle.Render("cwd: ") + brightStyle.Render(cwdDisplay))
	list.WriteString(dimStyle.Render(fmt.Sprintf("  %d projects", len(t.projects))))
	list.WriteString("\n\n")

	if len(t.projects) == 0 {
		list.WriteString(dimStyle.Render("  No projects with pskill.yaml found.\n"))
		list.WriteString(dimStyle.Render("  Run 'pskill add --project <skill>' in a project to create one.\n"))
	}

	visibleH := l.ContentH - 3
	if visibleH < 1 {
		visibleH = 1
	}
	if t.cursor >= len(t.projects) && len(t.projects) > 0 {
		t.cursor = len(t.projects) - 1
	}

	start := 0
	if t.cursor >= visibleH {
		start = t.cursor - visibleH + 1
	}
	end := start + visibleH
	if end > len(t.projects) {
		end = len(t.projects)
	}

	for i := start; i < end; i++ {
		p := t.projects[i]
		prefix := "  "
		if i == t.cursor {
			prefix = selectedStyle.Render("> ")
		}

		name := p.Name
		tag := ""
		if p.IsCurrent {
			tag = successStyle.Render(" (current)")
		}

		skillCount := dimStyle.Render(fmt.Sprintf("%d skills", len(p.Skills)))

		if i == t.cursor {
			name = selectedStyle.Render(name)
		} else {
			name = brightStyle.Render(name)
		}

		list.WriteString(fmt.Sprintf("%s%-28s %s%s\n", prefix, name, skillCount, tag))
	}

	leftPane := activePaneStyle.Width(l.LeftW).Height(l.ContentH).Render(list.String())

	// === Right pane: project detail ===
	var detail strings.Builder

	if len(t.projects) > 0 && t.cursor < len(t.projects) {
		p := t.projects[t.cursor]
		pathDisplay := shortenPath(p.Path, home)

		detail.WriteString(titleStyle.Render(p.Name))
		if p.IsCurrent {
			detail.WriteString(successStyle.Render("  current"))
		}
		detail.WriteString("\n\n")

		detail.WriteString(dimStyle.Render("Path: ") + brightStyle.Render(pathDisplay) + "\n\n")

		// Target CLIs
		if len(p.CLIs) > 0 {
			detail.WriteString(brightStyle.Render("Target CLIs") + "\n")
			detail.WriteString(dimStyle.Render(strings.Repeat("─", l.RightW-4)) + "\n")
			for _, cli := range p.CLIs {
				detail.WriteString("  " + cliBadgeInline(cli) + "\n")
			}
			detail.WriteString("\n")
		}

		// Installed skills
		detail.WriteString(brightStyle.Render("Installed Skills") + "\n")
		detail.WriteString(dimStyle.Render(strings.Repeat("─", l.RightW-4)) + "\n")

		if len(p.Skills) == 0 {
			detail.WriteString(dimStyle.Render("  No skills installed\n"))
		} else {
			for _, sk := range p.Skills {
				detail.WriteString("  " + brightStyle.Render("• ") + brightStyle.Render(sk) + "\n")
			}
		}

		detail.WriteString("\n")
		detail.WriteString(dimStyle.Render(fmt.Sprintf("Manifest: %s/pskill.yaml", pathDisplay)))
	} else {
		detail.WriteString(dimStyle.Render("No project selected"))
	}

	rightPane := paneStyle.Width(l.RightW).Height(l.ContentH).Render(detail.String())

	if !l.HasDetail {
		return leftPane
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
}

func (t *ProjectsTab) Title() string      { return "Projects" }
func (t *ProjectsTab) AcceptsTextInput() bool { return false }
func (t *ProjectsTab) ShortHelp() []string {
	return []string{
		helpEntry("↑/↓", "nav"),
		helpEntry("r", "refresh"),
	}
}

// shortenPath replaces home dir prefix with ~ for display.
func shortenPath(path, home string) string {
	if home != "" && strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}
