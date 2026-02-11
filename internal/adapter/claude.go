package adapter

import (
	"os"
	"path/filepath"
)

type ClaudeAdapter struct{}

func (ClaudeAdapter) Name() string { return "claude" }
func (ClaudeAdapter) SupportsSkills() bool {
	return true
}
func (ClaudeAdapter) SkillDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "skills")
}
