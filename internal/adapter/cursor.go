package adapter

import (
	"os"
	"path/filepath"
)

type CursorAdapter struct{}

func (CursorAdapter) Name() string { return "cursor" }
func (CursorAdapter) SupportsSkills() bool {
	return true
}
func (CursorAdapter) SkillDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cursor", "skills")
}
