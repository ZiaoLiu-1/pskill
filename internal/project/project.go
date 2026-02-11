package project

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Manifest struct {
	Name          string   `yaml:"name"`
	TargetCLIs    []string `yaml:"targetClis"`
	DefaultSkills []string `yaml:"defaultSkills"`
	Installed     []string `yaml:"installed"`
}

func pathFor(dir string) string {
	return filepath.Join(dir, "pskill.yaml")
}

func Load(dir string) (Manifest, error) {
	p := pathFor(dir)
	raw, err := os.ReadFile(p)
	if err != nil {
		return Manifest{}, err
	}
	var m Manifest
	if err := yaml.Unmarshal(raw, &m); err != nil {
		return Manifest{}, err
	}
	return m, nil
}

func Save(dir string, m Manifest) error {
	raw, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(pathFor(dir), raw, 0o644)
}
