package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ZiaoLiu-1/pskill/internal/detector"
	"github.com/ZiaoLiu-1/pskill/internal/skill"
)

type Inventory struct {
	Skills []skill.Skill `json:"skills"`
}

func ScanSystemSkills() (Inventory, error) {
	installed, err := detector.DetectInstalledCLIs()
	if err != nil {
		return Inventory{}, err
	}

	out := Inventory{Skills: []skill.Skill{}}
	seen := map[string]bool{}

	for _, cli := range installed {
		if !cli.Installed {
			continue
		}

		// Scan multiple skill directories per CLI
		dirs := skillDirsFor(cli)
		for _, dir := range dirs {
			scanDir(dir, cli.Name, &out, seen)
		}
	}
	return out, nil
}

func skillDirsFor(cli detector.CLIInfo) []string {
	dirs := []string{}

	if cli.SupportsSkills && cli.SkillDir != "" {
		dirs = append(dirs, cli.SkillDir)
	}

	// Also check built-in skill directories
	switch cli.Name {
	case "cursor":
		dirs = append(dirs, filepath.Join(cli.BaseDir, "skills-cursor"))
	case "codex":
		dirs = append(dirs, filepath.Join(cli.BaseDir, "skills", ".system"))
	case "claude":
		// Check plugins for skills
		pluginCache := filepath.Join(cli.BaseDir, "plugins", "cache")
		if entries, err := os.ReadDir(pluginCache); err == nil {
			for _, src := range entries {
				if !src.IsDir() {
					continue
				}
				srcPath := filepath.Join(pluginCache, src.Name())
				scanPluginSource(srcPath, &dirs)
			}
		}
	}

	return dirs
}

func scanPluginSource(srcPath string, dirs *[]string) {
	plugins, err := os.ReadDir(srcPath)
	if err != nil {
		return
	}
	for _, plugin := range plugins {
		if !plugin.IsDir() {
			continue
		}
		pluginPath := filepath.Join(srcPath, plugin.Name())
		versions, err := os.ReadDir(pluginPath)
		if err != nil {
			continue
		}
		for _, ver := range versions {
			if !ver.IsDir() {
				continue
			}
			skillsDir := filepath.Join(pluginPath, ver.Name(), "skills")
			if _, err := os.Stat(skillsDir); err == nil {
				*dirs = append(*dirs, skillsDir)
			}
		}
	}
}

func scanDir(dir, cliName string, out *Inventory, seen map[string]bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		path := filepath.Join(dir, name, "SKILL.md")
		if _, err := os.Stat(path); err != nil {
			continue
		}

		// Use directory name as the canonical skill name
		if seen[name] {
			continue
		}
		seen[name] = true

		sk, err := skill.ParseFile(path, cliName)
		if err != nil {
			continue
		}
		// Override with directory name - this is the canonical identifier
		sk.Name = name
		if sk.Description == "" {
			sk.Description = strings.ReplaceAll(name, "-", " ")
		}

		out.Skills = append(out.Skills, sk)
	}
}
