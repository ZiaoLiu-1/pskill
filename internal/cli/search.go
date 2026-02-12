package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/registry"
	"github.com/ZiaoLiu-1/pskill/internal/search"
)

func newSearchCmd() *cobra.Command {
	var online bool
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search skills by meaning",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.Join(args, " ")
			cfg, err := config.LoadGlobal()
			if err != nil {
				return err
			}
			engine := search.NewEngine(cfg.IndexDir)
			local, _ := engine.Search(query, 10)
			for i, item := range local {
				fmt.Printf("L%02d %-28s %.2f\n", i+1, item.Name, item.Score)
			}
			if online {
				client := registry.NewClient(cfg.RegistryURL, cfg.CacheDir, cfg.RegistryAPIKey)
				remote, _ := client.AISearch(query)
				for i, item := range remote {
					fmt.Printf("R%02d %-28s %.2f  by %s\n", i+1, item.Name, item.Score, item.Author)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&online, "online", false, "also search skillsmp.com")
	return cmd
}
