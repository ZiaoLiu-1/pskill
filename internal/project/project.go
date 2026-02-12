package project

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Manifest struct {
	Name          string   `yaml:"name"`
	TargetCLIs    []string `yaml:"targetClis"`
	DefaultSkills []string `yaml:"defaultSkills"`
	Installed     []string `yaml:"installed"`
}

// Info represents a discovered project on disk.
type Info struct {
	Path      string   // absolute directory path
	Name      string   // project name (from manifest or dirname)
	Skills    []string // installed skill names
	CLIs      []string // target CLIs
	IsCurrent bool     // true if this is the cwd
}

func pathFor(dir string) string {
	return filepath.Join(dir, "pskill.yaml")
}

// Exists returns true if the directory contains a pskill.yaml.
func Exists(dir string) bool {
	_, err := os.Stat(pathFor(dir))
	return err == nil
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

// Discover scans common directories for projects containing pskill.yaml.
// It walks up to maxDepth levels under each search root.
// Returns results sorted with the current directory first.
func Discover(searchRoots []string, maxDepth int) []Info {
	cwd, _ := os.Getwd()

	seen := map[string]bool{}
	var results []Info

	for _, root := range searchRoots {
		root = expandHome(root)
		walkForProjects(root, 0, maxDepth, cwd, seen, &results)
	}

	// Check cwd explicitly in case it's not under a search root
	if cwd != "" && !seen[cwd] {
		if Exists(cwd) {
			info := loadInfo(cwd, cwd)
			results = append([]Info{info}, results...)
			seen[cwd] = true
		}
	}

	return results
}

func walkForProjects(dir string, depth, maxDepth int, cwd string, seen map[string]bool, results *[]Info) {
	if depth > maxDepth {
		return
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return
	}

	if seen[abs] {
		return
	}

	// Check if this directory has a pskill.yaml
	if Exists(abs) {
		seen[abs] = true
		info := loadInfo(abs, cwd)
		*results = append(*results, info)
		// Don't recurse into projects (they won't contain nested projects)
		return
	}

	// Don't recurse into hidden directories or node_modules etc
	base := filepath.Base(abs)
	if strings.HasPrefix(base, ".") || base == "node_modules" || base == "vendor" || base == "__pycache__" {
		return
	}

	entries, err := os.ReadDir(abs)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		walkForProjects(filepath.Join(abs, e.Name()), depth+1, maxDepth, cwd, seen, results)
	}
}

func loadInfo(dir, cwd string) Info {
	info := Info{
		Path:      dir,
		Name:      filepath.Base(dir),
		IsCurrent: dir == cwd,
	}
	m, err := Load(dir)
	if err == nil {
		if m.Name != "" {
			info.Name = m.Name
		}
		info.Skills = m.Installed
		info.CLIs = m.TargetCLIs
	}
	return info
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}
