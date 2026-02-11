package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ZiaoLiu-1/pskill/internal/adapter"
	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/store"
)

func newListCmd() *cobra.Command {
	var asJSON bool
	var cliName string
	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List installed skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadGlobal()
			if err != nil {
				return err
			}
			st := store.NewManager(cfg.StoreDir)
			skills, err := st.ListSkills()
			if err != nil {
				return err
			}

			if cliName != "" {
				key := strings.ToLower(strings.TrimSpace(cliName))
				ad, ok := adapter.All()[key]
				if !ok || !ad.SupportsSkills() {
					fmt.Fprintf(os.Stderr, "pskill: CLI %q not found or does not support skills\n", cliName)
					return nil
				}
				skillDir := ad.SkillDir()
				filtered := make([]string, 0, len(skills))
				for _, s := range skills {
					linkPath := filepath.Join(skillDir, s)
					if _, err := os.Stat(linkPath); err == nil {
						filtered = append(filtered, s)
					}
				}
				skills = filtered
			}

			if asJSON {
				out, _ := json.MarshalIndent(skills, "", "  ")
				fmt.Println(string(out))
				return nil
			}
			for _, s := range skills {
				fmt.Println(s)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "output as JSON")
	cmd.Flags().StringVar(&cliName, "cli", "", "filter by cli name")
	return cmd
}
