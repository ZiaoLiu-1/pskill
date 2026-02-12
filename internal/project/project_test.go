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
