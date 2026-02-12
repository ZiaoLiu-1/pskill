package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZiaoLiu-1/pskill/internal/adapter"
	"github.com/ZiaoLiu-1/pskill/internal/config"
	"github.com/ZiaoLiu-1/pskill/internal/monitor"
	"github.com/ZiaoLiu-1/pskill/internal/project"
	"github.com/ZiaoLiu-1/pskill/internal/registry"
	"github.com/ZiaoLiu-1/pskill/internal/search"
	"github.com/ZiaoLiu-1/pskill/internal/store"
)

// Result reports what happened during an install.
type Result struct {
	SkillName   string
	StorePath   string
	LinkedCLIs  []string // e.g. "cursor (project)", "claude (global)"
	ProjectPath string   // cwd if project manifest was updated
}

// InstallFromRegistryResult downloads a skill into the central store,
// symlinks it into project-local AND global CLI skill directories,
// indexes for local search, records to monitor, and updates pskill.yaml.
func InstallFromRegistryResult(cfg config.Config, result registry.SkillResult, markProject bool) (*Result, error) {
	skillName := strings.TrimSpace(result.Name)
	if skillName == "" {
		return nil, fmt.Errorf("skill name is empty")
	}

	res := &Result{SkillName: skillName}

	// 1. Download into central store
	st := store.NewManager(cfg.StoreDir)
	destPath, exists, err := st.EnsureSkillDir(skillName)
	if err != nil {
		return nil, fmt.Errorf("create store dir: %w", err)
	}
	res.StorePath = destPath

	if !exists {
		client := registry.NewClient(cfg.RegistryURL, cfg.CacheDir, cfg.RegistryAPIKey)
		if err := client.DownloadSkill(skillName, result.GithubURL, destPath); err != nil {
			return nil, fmt.Errorf("download: %w", err)
		}
	}

	// 2. Symlink into global CLI skill directories (~/.cursor/skills/, etc.)
	adapters := adapter.All()
	for _, target := range cfg.TargetCLIs {
		ad, ok := adapters[strings.TrimSpace(target)]
		if !ok || !ad.SupportsSkills() {
			continue
		}
		globalDir := ad.SkillDir()
		if err := st.LinkSkillToCLI(skillName, globalDir); err == nil {
			res.LinkedCLIs = append(res.LinkedCLIs, target+" (global)")
		}
	}

	// 3. Symlink into project-local CLI skill directories (<cwd>/.cursor/skills/, etc.)
	wd, _ := os.Getwd()
	if wd != "" {
		for _, target := range cfg.TargetCLIs {
			localDir := projectCLISkillDir(wd, target)
			if localDir == "" {
				continue
			}
			if err := st.LinkSkillToCLI(skillName, localDir); err == nil {
				res.LinkedCLIs = append(res.LinkedCLIs, target+" (project)")
			}
		}
	}

	// 4. Index for local search
	engine := search.NewEngine(cfg.IndexDir)
	_ = engine.IndexSkillByPath(skillName, destPath)

	// 5. Update project manifest (pskill.yaml)
	if markProject && wd != "" {
		manifest, err := project.Load(wd)
		if err != nil {
			name := filepath.Base(wd)
			if name == "" || name == "." || name == "/" {
				name = "project"
			}
			manifest = project.Manifest{Name: name, TargetCLIs: cfg.TargetCLIs}
		}
		manifest.Installed = appendIfMissing(manifest.Installed, skillName)
		manifest.TargetCLIs = cfg.TargetCLIs
		_ = project.Save(wd, manifest)
		res.ProjectPath = wd
	}

	// 6. Record usage event
	if tr, err := monitor.NewTracker(cfg.StatsDB); err == nil {
		cliName := "global"
		if len(cfg.TargetCLIs) > 0 {
			cliName = cfg.TargetCLIs[0]
		}
		_ = tr.Record(monitor.Event{
			SkillName: skillName,
			CLI:       cliName,
			Project:   filepath.Base(wd),
			EventType: "install",
		})
		_ = tr.Close()
	}

	return res, nil
}

// projectCLISkillDir returns the project-local skills directory for a CLI.
// e.g. for "cursor" in /Users/me/myproject → /Users/me/myproject/.cursor/skills
func projectCLISkillDir(projectDir, cliName string) string {
	switch cliName {
	case "cursor":
		return filepath.Join(projectDir, ".cursor", "skills")
	case "claude":
		return filepath.Join(projectDir, ".claude", "skills")
	case "codex":
		return filepath.Join(projectDir, ".codex", "skills")
	default:
		return ""
	}
}

// UninstallFromProject removes symlinks from both project-local AND global
// CLI skill directories, removes from pskill.yaml, and removes from central store.
func UninstallFromProject(cfg config.Config, skillName string) error {
	wd, _ := os.Getwd()
	if wd == "" {
		return fmt.Errorf("cannot determine working directory")
	}

	adapters := adapter.All()

	// Remove project-local symlinks (<cwd>/.cursor/skills/<name>, etc.)
	for _, target := range cfg.TargetCLIs {
		localDir := projectCLISkillDir(wd, target)
		if localDir == "" {
			continue
		}
		_ = os.Remove(filepath.Join(localDir, skillName))
	}

	// Remove global CLI symlinks (~/.cursor/skills/<name>, etc.)
	for _, target := range cfg.TargetCLIs {
		ad, ok := adapters[strings.TrimSpace(target)]
		if !ok || !ad.SupportsSkills() {
			continue
		}
		_ = os.Remove(filepath.Join(ad.SkillDir(), skillName))
	}

	// Remove from central store
	st := store.NewManager(cfg.StoreDir)
	_ = st.RemoveSkill(skillName)

	// Update project manifest
	manifest, err := project.Load(wd)
	if err == nil {
		manifest.Installed = removeItem(manifest.Installed, skillName)
		_ = project.Save(wd, manifest)
	}

	// Record event
	if tr, err := monitor.NewTracker(cfg.StatsDB); err == nil {
		_ = tr.Record(monitor.Event{
			SkillName: skillName,
			CLI:       "global",
			Project:   filepath.Base(wd),
			EventType: "uninstall",
		})
		_ = tr.Close()
	}

	return nil
}

func appendIfMissing(items []string, item string) []string {
	for _, it := range items {
		if it == item {
			return items
		}
	}
	return append(items, item)
}

func removeItem(items []string, item string) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		if it != item {
			out = append(out, it)
		}
	}
	return out
}
