package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	m := Manifest{
		Name:          "test-project",
		TargetCLIs:    []string{"cursor", "claude"},
		DefaultSkills: []string{"frontend-design"},
		Installed:     []string{"frontend-design", "resume-tailoring"},
	}

	if err := Save(dir, m); err != nil {
		t.Fatal(err)
	}

	// Verify file exists
	if _, err := os.Stat(filepath.Join(dir, "pskill.yaml")); err != nil {
		t.Fatal("pskill.yaml not created")
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	if loaded.Name != "test-project" {
		t.Errorf("expected name 'test-project', got %q", loaded.Name)
	}
	if len(loaded.TargetCLIs) != 2 {
		t.Errorf("expected 2 target CLIs, got %d", len(loaded.TargetCLIs))
	}
	if len(loaded.Installed) != 2 {
		t.Errorf("expected 2 installed skills, got %d", len(loaded.Installed))
	}
}

func TestLoad_NotFound(t *testing.T) {
	_, err := Load(t.TempDir())
	if err == nil {
		t.Error("expected error for missing manifest")
	}
}

func TestPathFor(t *testing.T) {
	got := pathFor("/some/dir")
	want := filepath.Join("/some/dir", "pskill.yaml")
	if got != want {
		t.Errorf("pathFor = %q, want %q", got, want)
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	if Exists(dir) {
		t.Error("expected Exists=false for empty dir")
	}
	if err := Save(dir, Manifest{Name: "test"}); err != nil {
		t.Fatal(err)
	}
	if !Exists(dir) {
		t.Error("expected Exists=true after saving manifest")
	}
}

func TestDiscover(t *testing.T) {
	root := t.TempDir()

	// Create two project dirs with pskill.yaml
	projA := filepath.Join(root, "proj-a")
	projB := filepath.Join(root, "proj-b")
	noProj := filepath.Join(root, "no-proj")
	_ = os.MkdirAll(projA, 0o755)
	_ = os.MkdirAll(projB, 0o755)
	_ = os.MkdirAll(noProj, 0o755)

	_ = Save(projA, Manifest{Name: "alpha", Installed: []string{"skill-a"}})
	_ = Save(projB, Manifest{Name: "beta", Installed: []string{"skill-b", "skill-c"}})

	results := Discover([]string{root}, 2)
	if len(results) < 2 {
		t.Fatalf("expected at least 2 projects, got %d", len(results))
	}

	names := map[string]bool{}
	for _, r := range results {
		names[r.Name] = true
	}
	if !names["alpha"] {
		t.Error("expected to find project 'alpha'")
	}
	if !names["beta"] {
		t.Error("expected to find project 'beta'")
	}
}

func TestDiscover_Empty(t *testing.T) {
	root := t.TempDir()
	results := Discover([]string{root}, 2)
	if len(results) != 0 {
		t.Errorf("expected 0 projects, got %d", len(results))
	}
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := expandHome("~/test")
	want := filepath.Join(home, "test")
	if got != want {
		t.Errorf("expandHome(~/test) = %q, want %q", got, want)
	}

	got2 := expandHome("/absolute/path")
	if got2 != "/absolute/path" {
		t.Errorf("expandHome should not change absolute paths, got %q", got2)
	}
}
