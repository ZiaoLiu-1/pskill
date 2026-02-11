package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ZiaoLiu-1/pskill/internal/detector"
)

func newDetectCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "detect",
		Short: "Detect installed LLM CLIs and skill directories",
		RunE: func(cmd *cobra.Command, args []string) error {
			items, err := detector.DetectInstalledCLIs()
			if err != nil {
				return err
			}
			if asJSON {
				out, _ := json.MarshalIndent(items, "", "  ")
				fmt.Println(string(out))
				return nil
			}
			for _, it := range items {
				fmt.Printf("%-8s found=%t skills=%t dir=%s\n", it.Name, it.Installed, it.SupportsSkills, it.SkillDir)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "output as JSON")
	return cmd
}
