package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/tui"
)

func newMonitorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "monitor",
		Short: "Open monitor dashboard tab",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadGlobal()
			if err != nil {
				return err
			}
			app := tui.NewAppWithTab(cfg, debug, tui.TabMonitor)
			prog := tea.NewProgram(app, tea.WithAltScreen())
			_, err = prog.Run()
			return err
		},
	}
}
