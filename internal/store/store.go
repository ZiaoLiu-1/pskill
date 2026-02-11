package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZiaoLiu-1/pskill/internal/detector"
	"github.com/ZiaoLiu-1/pskill/internal/skill"
)

type Manager struct {
	storeDir string
}

func NewManager(storeDir string) *Manager {
	return &Manager{storeDir: storeDir}
}

func (m *Manager) EnsureSkillDir(name string) (string, bool, error) {
	if err := os.MkdirAll(m.storeDir, 0o755); err != nil {
		return "", false, err
	}
	path := filepath.Join(m.storeDir, name)
	_, err := os.Stat(path)
	if err == nil {
		return path, true, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", false, err
	}
	return path, false, os.MkdirAll(path, 0o755)
}

func (m *Manager) ImportSkill(sk skill.Skill) error {
	dest, _, err := m.EnsureSkillDir(sk.Name)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(sk.Path)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dest, "SKILL.md"), raw, 0o644)
}

func (m *Manager) RemoveSkill(name string) error {
	return os.RemoveAll(filepath.Join(m.storeDir, name))
}

func (m *Manager) ListSkills() ([]string, error) {
	entries, err := os.ReadDir(m.storeDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

func (m *Manager) LinkSkillToCLI(skillName, cliDir string) error {
	if cliDir == "" {
		return nil
	}
	src := filepath.Join(m.storeDir, skillName)
	dst := filepath.Join(cliDir, skillName)
	return EnsureSymlink(src, dst)
}

func (m *Manager) UnlinkSkillEverywhere(skillName string) error {
	clis, err := detector.DetectInstalledCLIs()
	if err != nil {
		return err
	}
	for _, cli := range clis {
		if !cli.SupportsSkills || !cli.Installed || strings.TrimSpace(cli.SkillDir) == "" {
			continue
		}
		_ = os.Remove(filepath.Join(cli.SkillDir, skillName))
	}
	return nil
}
