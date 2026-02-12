package installer

import (
	"os"

	"github.com/ZiaoLiu-1/pskill/internal/project"
	"github.com/ZiaoLiu-1/pskill/internal/store"
)

// SkillStatus describes where a skill is installed.
type SkillStatus int

const (
	StatusNotInstalled    SkillStatus = iota // not in central store
	StatusInStore                            // in central store but not in current project
	StatusInProject                          // in central store AND in current project
)

// CheckStatus returns the install status of a skill by name.
func CheckStatus(storeDir, skillName string) SkillStatus {
	st := store.NewManager(storeDir)
	skills, err := st.ListSkills()
	if err != nil {
		return StatusNotInstalled
	}
	inStore := false
	for _, s := range skills {
		if s == skillName {
			inStore = true
			break
		}
	}
	if !inStore {
		return StatusNotInstalled
	}

	wd, _ := os.Getwd()
	if wd != "" {
		m, err := project.Load(wd)
		if err == nil {
			for _, s := range m.Installed {
				if s == skillName {
					return StatusInProject
				}
			}
		}
	}

	return StatusInStore
}

// BuildStatusMap builds a lookup for a batch of skill names.
// Much more efficient than calling CheckStatus per-skill.
func BuildStatusMap(storeDir string) (projectSet map[string]bool, storeSet map[string]bool) {
	projectSet = map[string]bool{}
	storeSet = map[string]bool{}

	st := store.NewManager(storeDir)
	skills, err := st.ListSkills()
	if err == nil {
		for _, s := range skills {
			storeSet[s] = true
		}
	}

	wd, _ := os.Getwd()
	if wd != "" {
		m, err := project.Load(wd)
		if err == nil {
			for _, s := range m.Installed {
				projectSet[s] = true
			}
		}
	}
	return
}
