package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/detector"
	"github.com/ZiaoLiu-1/pskill/internal/scanner"
	"github.com/ZiaoLiu-1/pskill/internal/store"
)

func newInitCmd() *cobra.Command {
	var importExisting bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize pskill defaults for this machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadGlobal()
			if err != nil {
				return err
			}

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
		},
	}

	cmd.Flags().BoolVar(&importExisting, "import-existing", true, "scan existing skills and import to central store")
	return cmd
}
