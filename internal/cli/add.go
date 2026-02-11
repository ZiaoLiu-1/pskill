package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ZiaoLiu-1/pskill/internal/adapter"
	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/project"
	"github.com/ZiaoLiu-1/pskill/internal/registry"
	"github.com/ZiaoLiu-1/pskill/internal/search"
	"github.com/ZiaoLiu-1/pskill/internal/store"
)

func newAddCmd() *cobra.Command {
	var cliTargets string
	var projectScope bool

	cmd := &cobra.Command{
		Use:   "add <skill-name>",
		Short: "Install a skill to store and selected CLIs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillName := args[0]
			cfg, err := config.LoadGlobal()
			if err != nil {
				return err
			}

			st := store.NewManager(cfg.StoreDir)
			destPath, exists, err := st.EnsureSkillDir(skillName)
			if err != nil {
				return err
			}

			if !exists {
				client := registry.NewClient(cfg.RegistryURL, cfg.CacheDir)
				if err := client.DownloadSkill(skillName, destPath); err != nil {
					return err
				}
			}

			targets := cfg.TargetCLIs
			if cliTargets != "" {
				targets = strings.Split(cliTargets, ",")
			}
			adapters := adapter.All()
			for _, t := range targets {
				ad, ok := adapters[strings.TrimSpace(t)]
				if !ok {
					continue
				}
				if err := st.LinkSkillToCLI(skillName, ad.SkillDir()); err != nil {
					fmt.Fprintf(os.Stderr, "warn: unable to link %s to %s: %v\n", skillName, ad.Name(), err)
				}
			}

			engine := search.NewEngine(cfg.IndexDir)
			_ = engine.IndexSkillByPath(skillName, destPath)

			scope := "global"
			if projectScope {
				wd, _ := os.Getwd()
				manifest, err := project.Load(wd)
				if err != nil {
					name := filepath.Base(wd)
					if name == "" || name == "." || name == "/" {
						name = "project"
					}
					manifest = project.Manifest{Name: name, TargetCLIs: targets}
				}
				manifest.Installed = appendIfMissing(manifest.Installed, skillName)
				manifest.TargetCLIs = targets
				_ = project.Save(wd, manifest)
				scope = "project"
			}
			fmt.Fprintf(os.Stdout, "Installed %s (%s)\n", skillName, scope)
			return nil
		},
	}

	cmd.Flags().StringVar(&cliTargets, "cli", "", "comma-separated target CLIs")
	cmd.Flags().BoolVar(&projectScope, "project", false, "mark skill for current project")
	return cmd
}

func appendIfMissing(items []string, item string) []string {
	for _, it := range items {
		if it == item {
			return items
		}
	}
	return append(items, item)
}
