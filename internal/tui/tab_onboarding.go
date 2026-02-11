package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/detector"
	"github.com/ZiaoLiu-1/pskill/internal/scanner"
	"github.com/ZiaoLiu-1/pskill/internal/skill"
	"github.com/ZiaoLiu-1/pskill/internal/store"
)

// --- onboarding step state machine ---

type onboardingStep int

const (
	stepWelcome onboardingStep = iota
	stepDetect
	stepSelectCLIs
	stepScanImport
	stepDefaultPack
	stepComplete
)

// --- messages ---

type cliDetectedMsg struct {
	clis []detector.CLIInfo
}

type skillsImportedMsg struct {
	skills []skill.Skill
	names  []string
}

// --- tab ---

type OnboardingTab struct {
	cfg    config.Config
	step   onboardingStep
	width  int
	height int

	// step 2: detected CLIs
	clis     []detector.CLIInfo
	detected bool

	// step 3: CLI selection
	cliCursor int

	// step 4: scan results
	scannedSkills []skill.Skill
	scannedNames  []string
	scanning      bool
	imported      bool

	// step 5: default pack selection
	packCursor int
	packToggle []bool
}

func NewOnboardingTab(cfg config.Config) Tab {
	return &OnboardingTab{
		cfg:  cfg,
		step: stepWelcome,
	}
}

func (t *OnboardingTab) Init() tea.Cmd { return nil }

func (t *OnboardingTab) Update(msg tea.Msg) (Tab, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = m.Width
		t.height = m.Height
		return t, nil

	case tea.KeyMsg:
		return t.handleKey(m)

	case cliDetectedMsg:
		t.clis = m.clis
		t.detected = true
		// Auto-advance to select CLIs
		t.step = stepSelectCLIs
		return t, nil

	case skillsImportedMsg:
		t.scannedSkills = m.skills
		t.scannedNames = m.names
		t.scanning = false
		t.imported = true
		// Auto-advance to default pack
		t.step = stepDefaultPack
		t.packToggle = make([]bool, len(t.scannedNames))
		return t, nil
	}
	return t, nil
}

func (t *OnboardingTab) handleKey(m tea.KeyMsg) (Tab, tea.Cmd) {
	key := m.String()

	switch t.step {

	case stepWelcome:
		if key == "enter" {
			t.step = stepDetect
			return t, t.detectCmd()
		}

	case stepDetect:
		// Waiting for async detection, no keys

	case stepSelectCLIs:
		compatCLIs := t.compatibleCLIs()
		switch key {
		case "j", "down":
			if t.cliCursor < len(compatCLIs)-1 {
				t.cliCursor++
			}
		case "k", "up":
			if t.cliCursor > 0 {
				t.cliCursor--
			}
		case " ":
			if len(compatCLIs) > 0 {
				idx := compatCLIs[t.cliCursor].origIdx
				t.clis[idx].SupportsSkills = !t.clis[idx].SupportsSkills
			}
		case "enter":
			// Save selected targets
			targets := make([]string, 0)
			for _, c := range t.clis {
				if c.Installed && c.SupportsSkills {
					targets = append(targets, c.Name)
				}
			}
			t.cfg.TargetCLIs = targets
			// Move to scan step
			t.step = stepScanImport
			t.scanning = true
			return t, t.scanCmd()
		}

	case stepScanImport:
		// Waiting for async scan, no keys

	case stepDefaultPack:
		switch key {
		case "j", "down":
			if t.packCursor < len(t.scannedNames)-1 {
				t.packCursor++
			}
		case "k", "up":
			if t.packCursor > 0 {
				t.packCursor--
			}
		case " ":
			if len(t.packToggle) > 0 {
				t.packToggle[t.packCursor] = !t.packToggle[t.packCursor]
			}
		case "a":
			for i := range t.packToggle {
				t.packToggle[i] = true
			}
		case "n":
			for i := range t.packToggle {
				t.packToggle[i] = false
			}
		case "enter", "esc":
			defaults := make([]string, 0)
			for i, name := range t.scannedNames {
				if i < len(t.packToggle) && t.packToggle[i] {
					defaults = append(defaults, name)
				}
			}
			t.cfg.DefaultSkills = defaults
			t.step = stepComplete
		}

	case stepComplete:
		if key == "enter" {
			return t, t.finishCmd()
		}
	}

	return t, nil
}

func (t *OnboardingTab) View(width, height int) string {
	if width > 0 {
		t.width = width
	}
	if height > 0 {
		t.height = height
	}

	switch t.step {
	case stepWelcome:
		return t.viewWelcome()
	case stepDetect:
		return t.viewDetect()
	case stepSelectCLIs:
		return t.viewSelectCLIs()
	case stepScanImport:
		return t.viewScanImport()
	case stepDefaultPack:
		return t.viewDefaultPack()
	case stepComplete:
		return t.viewComplete()
	}
	return ""
}

func (t *OnboardingTab) Title() string         { return "Setup" }
func (t *OnboardingTab) AcceptsTextInput() bool { return false }
func (t *OnboardingTab) ShortHelp() []string {
	switch t.step {
	case stepWelcome:
		return []string{helpEntry("enter", "begin setup")}
	case stepDetect:
		return []string{helpEntry("", "scanning...")}
	case stepSelectCLIs:
		return []string{helpEntry("j/k", "nav"), helpEntry("space", "toggle"), helpEntry("enter", "confirm")}
	case stepScanImport:
		return []string{helpEntry("", "importing...")}
	case stepDefaultPack:
		return []string{helpEntry("j/k", "nav"), helpEntry("space", "toggle"), helpEntry("a", "all"), helpEntry("n", "none"), helpEntry("enter", "confirm")}
	case stepComplete:
		return []string{helpEntry("enter", "open dashboard")}
	}
	return nil
}

// --- views ---

func (t *OnboardingTab) viewWelcome() string {
	var b strings.Builder

	logo := titleStyle.Render(`
    ██████╗ ███████╗██╗  ██╗██╗██╗     ██╗
    ██╔══██╗██╔════╝██║ ██╔╝██║██║     ██║
    ██████╔╝███████╗█████╔╝ ██║██║     ██║
    ██╔═══╝ ╚════██║██╔═██╗ ██║██║     ██║
    ██║     ███████║██║  ██╗██║███████╗███████╗
    ╚═╝     ╚══════╝╚═╝  ╚═╝╚═╝╚══════╝╚══════╝`)

	b.WriteString(logo + "\n")
	b.WriteString(dimStyle.Render("    Universal LLM Skill Manager") + "\n\n")
	b.WriteString(brightStyle.Render("    Welcome! Let's set up pskill for your system.") + "\n\n")
	b.WriteString(dimStyle.Render("    This wizard will:") + "\n")
	b.WriteString(dimStyle.Render("      1. Detect installed LLM CLIs") + "\n")
	b.WriteString(dimStyle.Render("      2. Choose which CLIs to manage") + "\n")
	b.WriteString(dimStyle.Render("      3. Scan & import existing skills") + "\n")
	b.WriteString(dimStyle.Render("      4. Set your default skill pack") + "\n\n")
	b.WriteString(selectedStyle.Render("    Press enter to continue...") + "\n")

	return t.pad(b.String())
}

func (t *OnboardingTab) viewDetect() string {
	var b strings.Builder
	b.WriteString(t.stepHeader(1, "Detect CLIs"))
	b.WriteString(warningStyle.Render("    Scanning system for LLM CLIs...") + "\n\n")

	for _, c := range t.clis {
		status := dangerStyle.Render("not found")
		if c.Installed {
			if c.SupportsSkills {
				status = successStyle.Render("ready")
			} else {
				status = warningStyle.Render("no skill support")
			}
		}
		loc := ""
		if c.Installed {
			loc = dimStyle.Render(shorten(c.BaseDir))
		}
		b.WriteString(fmt.Sprintf("    %-10s %-30s %s\n", c.Name, loc, status))
	}

	if !t.detected {
		b.WriteString("\n" + dimStyle.Render("    Please wait...") + "\n")
	}

	return t.pad(b.String())
}

func (t *OnboardingTab) viewSelectCLIs() string {
	var b strings.Builder
	b.WriteString(t.stepHeader(2, "Select Target CLIs"))
	b.WriteString(dimStyle.Render("    Which CLIs should pskill manage skills for?") + "\n")
	b.WriteString(dimStyle.Render("    (Skills will be symlinked to selected CLIs)") + "\n\n")

	compatCLIs := t.compatibleCLIs()
	for i, cc := range compatCLIs {
		c := t.clis[cc.origIdx]
		check := dimStyle.Render("[ ]")
		if c.SupportsSkills {
			check = successStyle.Render("[x]")
		}
		prefix := "    "
		if i == t.cliCursor {
			prefix = selectedStyle.Render("  > ")
		}
		dir := dimStyle.Render(shorten(c.SkillDir))
		b.WriteString(fmt.Sprintf("%s%s %-10s %s\n", prefix, check, c.Name, dir))
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("    j/k navigate   space toggle   enter confirm") + "\n")

	return t.pad(b.String())
}

func (t *OnboardingTab) viewScanImport() string {
	var b strings.Builder
	b.WriteString(t.stepHeader(3, "Scan & Import"))

	if t.scanning {
		b.WriteString(warningStyle.Render("    Scanning for existing skills...") + "\n")
	}

	if t.imported {
		b.WriteString(successStyle.Render(fmt.Sprintf("    Found %d skills:", len(t.scannedNames))) + "\n\n")
		for _, sk := range t.scannedSkills {
			b.WriteString(fmt.Sprintf("      %-24s %s  %s\n",
				brightStyle.Render(sk.Name),
				dimStyle.Render("("+sk.SourceCLI+")"),
				successStyle.Render("imported"),
			))
		}
		home, _ := os.UserHomeDir()
		b.WriteString(fmt.Sprintf("\n    All %d skills imported to %s\n",
			len(t.scannedNames),
			dimStyle.Render(strings.Replace(t.cfg.StoreDir, home, "~", 1)),
		))
	}

	return t.pad(b.String())
}

func (t *OnboardingTab) viewDefaultPack() string {
	var b strings.Builder
	b.WriteString(t.stepHeader(4, "Default Skill Pack"))
	b.WriteString(dimStyle.Render("    Choose skills to auto-install in new projects") + "\n")
	b.WriteString(dimStyle.Render("    (pskill init will install these by default)") + "\n\n")

	visible := t.height - 12
	if visible < 3 {
		visible = 3
	}

	start := 0
	if t.packCursor >= visible {
		start = t.packCursor - visible + 1
	}

	for i := start; i < len(t.scannedNames) && i < start+visible; i++ {
		name := t.scannedNames[i]
		check := dimStyle.Render("[ ]")
		if i < len(t.packToggle) && t.packToggle[i] {
			check = successStyle.Render("[x]")
		}
		prefix := "    "
		if i == t.packCursor {
			prefix = selectedStyle.Render("  > ")
		}
		b.WriteString(fmt.Sprintf("%s%s %s\n", prefix, check, brightStyle.Render(name)))
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("    j/k navigate  space toggle  a all  n none  enter confirm  esc skip") + "\n")

	return t.pad(b.String())
}

func (t *OnboardingTab) viewComplete() string {
	var b strings.Builder
	b.WriteString("\n\n")
	b.WriteString(successStyle.Render("    Setup complete!") + "\n\n")

	b.WriteString(fmt.Sprintf("    Target CLIs:     %s\n", brightStyle.Render(strings.Join(t.cfg.TargetCLIs, ", "))))
	b.WriteString(fmt.Sprintf("    Skills imported: %s\n", brightStyle.Render(fmt.Sprintf("%d", len(t.scannedNames)))))

	defCount := 0
	for _, v := range t.packToggle {
		if v {
			defCount++
		}
	}
	b.WriteString(fmt.Sprintf("    Default pack:    %s\n", brightStyle.Render(fmt.Sprintf("%d skills", defCount))))

	home, _ := os.UserHomeDir()
	b.WriteString(fmt.Sprintf("    Config saved:    %s\n", dimStyle.Render(strings.Replace(t.cfg.HomeDir, home, "~", 1)+"/config.yaml")))

	b.WriteString("\n")
	b.WriteString(selectedStyle.Render("    Press enter to open the dashboard...") + "\n")

	return t.pad(b.String())
}

// --- commands ---

func (t *OnboardingTab) detectCmd() tea.Cmd {
	return func() tea.Msg {
		clis, _ := detector.DetectInstalledCLIs()
		return cliDetectedMsg{clis: clis}
	}
}

func (t *OnboardingTab) scanCmd() tea.Cmd {
	cfg := t.cfg
	return func() tea.Msg {
		inv, _ := scanner.ScanSystemSkills()
		st := store.NewManager(cfg.StoreDir)
		names := make([]string, 0, len(inv.Skills))
		for _, sk := range inv.Skills {
			_ = st.ImportSkill(sk)
			names = append(names, sk.Name)
		}
		return skillsImportedMsg{skills: inv.Skills, names: names}
	}
}

func (t *OnboardingTab) finishCmd() tea.Cmd {
	cfg := t.cfg
	names := t.scannedNames
	count := len(names)
	return func() tea.Msg {
		_ = config.SaveGlobal(cfg)
		return onboardingDoneMsg{
			cfg:          cfg,
			scannedNames: names,
			scannedCount: count,
		}
	}
}

// --- helpers ---

type compatCLI struct {
	origIdx int
}

func (t *OnboardingTab) compatibleCLIs() []compatCLI {
	out := make([]compatCLI, 0)
	for i, c := range t.clis {
		if c.Installed {
			out = append(out, compatCLI{origIdx: i})
		}
	}
	return out
}

func (t *OnboardingTab) stepHeader(num int, title string) string {
	return fmt.Sprintf("\n    %s  %s\n    %s\n\n",
		titleStyle.Render(fmt.Sprintf("Step %d", num)),
		brightStyle.Render(title),
		dimStyle.Render(strings.Repeat("─", 45)),
	)
}

func (t *OnboardingTab) pad(content string) string {
	return content
}

func shorten(path string) string {
	home, _ := os.UserHomeDir()
	return strings.Replace(path, home, "~", 1)
}
