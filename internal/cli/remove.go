package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/monitor"
	"github.com/ZiaoLiu-1/pskill/internal/store"
)

func newRemoveCmd() *cobra.Command {
	var prune bool
	cmd := &cobra.Command{
		Use:   "remove <skill-name>",
		Short: "Unlink a skill and optionally prune store",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadGlobal()
			if err != nil {
				return err
			}
			st := store.NewManager(cfg.StoreDir)
			if err := st.UnlinkSkillEverywhere(args[0]); err != nil {
				return err
			}
			if prune {
				if err := st.RemoveSkill(args[0]); err != nil {
					return err
				}
			}
			// Record usage event
			if tr, err := monitor.NewTracker(cfg.StatsDB); err == nil {
				wd, _ := os.Getwd()
				_ = tr.Record(monitor.Event{
					SkillName: args[0],
					CLI:       "global",
					Project:   filepath.Base(wd),
					EventType: "remove",
				})
				tr.Close()
			}

			fmt.Printf("Removed %s\n", args[0])
			return nil
		},
	}
	cmd.Flags().BoolVar(&prune, "prune", false, "remove from central store too")
	return cmd
}
