package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ZiaoLiu-1/pskill/internal/adapter"
	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/scanner"
	"github.com/ZiaoLiu-1/pskill/internal/store"
)

func newScanCmd() *cobra.Command {
	var asJSON bool
	var importToStore bool
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan local system for existing skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			inv, err := scanner.ScanSystemSkills()
			if err != nil {
				return err
			}
			if asJSON {
				out, _ := json.MarshalIndent(inv, "", "  ")
				fmt.Println(string(out))
				return nil
			}
			if importToStore {
				cfg, err := config.LoadGlobal()
				if err == nil {
					st := store.NewManager(cfg.StoreDir)
					adapters := adapter.All()
					for _, sk := range inv.Skills {
						_ = st.ImportSkill(sk)
						if ad, ok := adapters[sk.SourceCLI]; ok {
							_ = st.LinkSkillToCLI(sk.Name, ad.SkillDir())
						}
					}
				}
			}
			fmt.Printf("Detected %d skills\n", len(inv.Skills))
			for _, sk := range inv.Skills {
				fmt.Printf("- %s [%s]\n", sk.Name, sk.SourceCLI)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "output as JSON")
	cmd.Flags().BoolVar(&importToStore, "import", true, "import scanned skills into central store and create symlinks")
	return cmd
}
