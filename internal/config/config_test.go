package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	d := defaults()
	if d.RegistryURL != "https://skillsmp.com" {
		t.Errorf("unexpected default registry: %s", d.RegistryURL)
	}
	if len(d.TargetCLIs) != 3 {
		t.Errorf("expected 3 default target CLIs, got %d", len(d.TargetCLIs))
	}
	if d.AutoUpdateTrending != true {
		t.Error("expected AutoUpdateTrending to default to true")
	}
	if d.HomeDir == "" || d.StoreDir == "" || d.CacheDir == "" || d.IndexDir == "" {
		t.Error("expected all directory paths to be non-empty")
	}
}

func TestSaveAndLoadGlobal(t *testing.T) {
	// Override HOME so we don't affect the real config
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	cfg := defaults()
	// Override to use temp dir
	cfg.HomeDir = filepath.Join(tmpHome, ".pskill")
	cfg.StoreDir = filepath.Join(cfg.HomeDir, "store")
	cfg.CacheDir = filepath.Join(cfg.HomeDir, "cache")
	cfg.IndexDir = filepath.Join(cfg.HomeDir, "index")
	cfg.StatsDB = filepath.Join(cfg.HomeDir, "stats.db")
	cfg.TargetCLIs = []string{"cursor", "claude"}
	cfg.DefaultSkills = []string{"frontend-design"}

	if err := SaveGlobal(cfg); err != nil {
		t.Fatal(err)
	}

	// Verify file was written
	if _, err := os.Stat(configPath()); err != nil {
		t.Fatal("config file not created")
	}
}

func TestIsFirstRun(t *testing.T) {
	// With a nonexistent HOME, config should not exist
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	if !IsFirstRun() {
		t.Error("expected IsFirstRun=true when no config exists")
	}
}

func TestConfigPath(t *testing.T) {
	path := configPath()
	if filepath.Base(path) != "config.yaml" {
		t.Errorf("expected config.yaml, got %s", filepath.Base(path))
	}
}
