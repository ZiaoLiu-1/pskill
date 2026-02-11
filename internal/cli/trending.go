package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/registry"
)

func newTrendingCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "trending",
		Short: "Show trending skills from skillsmp.com",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadGlobal()
			if err != nil {
				return err
			}
			client := registry.NewClient(cfg.RegistryURL, cfg.CacheDir)
			items, err := client.Trending(limit, "week", "all")
			if err != nil {
				return err
			}
			for i, it := range items {
				fmt.Printf("%2d. %-30s ★%d ↓%d\n", i+1, it.Name, it.Stars, it.Downloads)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 10, "number of results")
	return cmd
}
