package adapter

import (
	"os"
	"path/filepath"
)

type CodexAdapter struct{}

func (CodexAdapter) Name() string { return "codex" }
func (CodexAdapter) SupportsSkills() bool {
	return true
}
func (CodexAdapter) SkillDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".codex", "skills")
}
