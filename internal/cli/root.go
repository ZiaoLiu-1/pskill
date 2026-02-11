package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/tui"
)

var (
	debug bool
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pskill",
		Short: "Universal LLM skill manager",
		Long:  "pskill is a TUI-first package manager for LLM skills across Cursor, Claude, Codex, and more.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadGlobal()
			if err != nil {
				return err
			}

			app := tui.NewApp(cfg, debug)
			prog := tea.NewProgram(app, tea.WithAltScreen())
			_, err = prog.Run()
			return err
		},
	}

	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")
	cmd.AddCommand(
		newInitCmd(),
		newAddCmd(),
		newRemoveCmd(),
		newListCmd(),
		newDetectCmd(),
		newScanCmd(),
		newSearchCmd(),
		newTrendingCmd(),
		newMonitorCmd(),
	)
	return cmd
}

func Execute() error {
	return newRootCmd().Execute()
}

func withTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

func printJSON(data string) {
	fmt.Fprintln(os.Stdout, data)
}
