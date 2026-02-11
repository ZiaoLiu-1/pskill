package cli

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/detector"
	"github.com/ZiaoLiu-1/pskill/internal/scanner"
	"github.com/ZiaoLiu-1/pskill/internal/store"
	"github.com/ZiaoLiu-1/pskill/internal/tui"
)

func newInitCmd() *cobra.Command {
	var importExisting bool
	var nonInteractive bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize pskill (interactive wizard or quick setup)",
		Long:  "Run the interactive onboarding wizard to configure pskill, detect CLIs, scan skills, and set defaults. Use --no-tui for non-interactive mode.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadGlobal()
			if err != nil {
				return err
			}

			// Use interactive TUI wizard if stdout is a terminal and --no-tui not set
			if !nonInteractive && isTerminal() {
				app := tui.NewAppOnboarding(cfg, debug)
				prog := tea.NewProgram(app, tea.WithAltScreen())
				_, err := prog.Run()
				return err
			}

			// Non-interactive fallback
			return initNonInteractive(cfg, importExisting)
		},
	}

	cmd.Flags().BoolVar(&importExisting, "import-existing", true, "scan existing skills and import to central store")
	cmd.Flags().BoolVar(&nonInteractive, "no-tui", false, "run non-interactively (no wizard)")
	return cmd
}

func initNonInteractive(cfg config.Config, importExisting bool) error {
	detected, err := detector.DetectInstalledCLIs()
	if err != nil {
		return err
	}

	targets := make([]string, 0, len(detected))
	for _, cli := range detected {
		if cli.SupportsSkills {
			targets = append(targets, cli.Name)
		}
	}
	cfg.TargetCLIs = targets
	if err := config.SaveGlobal(cfg); err != nil {
		return err
	}

	if importExisting {
		inv, err := scanner.ScanSystemSkills()
		if err != nil {
			return err
		}
		st := store.NewManager(cfg.StoreDir)
		for _, sk := range inv.Skills {
			if err := st.ImportSkill(sk); err != nil {
				fmt.Fprintf(os.Stderr, "warn: unable to import %s: %v\n", sk.Name, err)
			}
		}
	}

	fmt.Fprintf(os.Stdout, "Initialized pskill with targets: %s\n", strings.Join(cfg.TargetCLIs, ", "))
	return nil
}

func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
