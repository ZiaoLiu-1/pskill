package detector

import (
	"os"
	"path/filepath"
)

type CLIInfo struct {
	Name           string `json:"name"`
	Installed      bool   `json:"installed"`
	SupportsSkills bool   `json:"supportsSkills"`
	BaseDir        string `json:"baseDir"`
	SkillDir       string `json:"skillDir"`
}

func DetectInstalledCLIs() ([]CLIInfo, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	candidates := []CLIInfo{
		{Name: "cursor", BaseDir: filepath.Join(home, ".cursor"), SkillDir: filepath.Join(home, ".cursor", "skills"), SupportsSkills: true},
		{Name: "claude", BaseDir: filepath.Join(home, ".claude"), SkillDir: filepath.Join(home, ".claude", "skills"), SupportsSkills: true},
		{Name: "codex", BaseDir: filepath.Join(home, ".codex"), SkillDir: filepath.Join(home, ".codex", "skills"), SupportsSkills: true},
		{Name: "gemini", BaseDir: filepath.Join(home, ".gemini"), SkillDir: "", SupportsSkills: false},
	}

	out := make([]CLIInfo, 0, len(candidates))
	for _, c := range candidates {
		_, err := os.Stat(c.BaseDir)
		c.Installed = err == nil
		out = append(out, c)
	}
	return out, nil
}
