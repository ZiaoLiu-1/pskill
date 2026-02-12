package store

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/ZiaoLiu-1/pskill/internal/skill"
)

func TestEnsureSkillDir_New(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(filepath.Join(dir, "store"))

	path, exists, err := m.EnsureSkillDir("my-skill")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("expected exists=false for new skill")
	}
	if filepath.Base(path) != "my-skill" {
		t.Errorf("expected path ending in 'my-skill', got %q", path)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal("skill directory was not created")
	}
	if !info.IsDir() {
		t.Error("expected a directory")
	}
}

func TestEnsureSkillDir_Existing(t *testing.T) {
	dir := t.TempDir()
	storeDir := filepath.Join(dir, "store")
	m := NewManager(storeDir)

	// Create it first
	_, _, _ = m.EnsureSkillDir("existing-skill")

	// Second call should report exists=true
	_, exists, err := m.EnsureSkillDir("existing-skill")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("expected exists=true for existing skill")
	}
}

func TestImportSkill(t *testing.T) {
	dir := t.TempDir()
	storeDir := filepath.Join(dir, "store")
	srcDir := filepath.Join(dir, "src")
	_ = os.MkdirAll(srcDir, 0o755)

	content := "---\nname: imported\n---\n\n# Imported Skill\n"
	srcPath := filepath.Join(srcDir, "SKILL.md")
	if err := os.WriteFile(srcPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	m := NewManager(storeDir)
	sk := skill.Skill{Name: "imported", Path: srcPath}
	if err := m.ImportSkill(sk); err != nil {
		t.Fatal(err)
	}

	// Verify the file was copied
	destPath := filepath.Join(storeDir, "imported", "SKILL.md")
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatal("imported SKILL.md not found in store")
	}
	if string(data) != content {
		t.Errorf("content mismatch: got %q", string(data))
	}
}

func TestRemoveSkill(t *testing.T) {
	dir := t.TempDir()
	storeDir := filepath.Join(dir, "store")
	m := NewManager(storeDir)

	_, _, _ = m.EnsureSkillDir("to-remove")
	if err := m.RemoveSkill("to-remove"); err != nil {
		t.Fatal(err)
	}

	skillPath := filepath.Join(storeDir, "to-remove")
	if _, err := os.Stat(skillPath); !os.IsNotExist(err) {
		t.Error("expected skill directory to be removed")
	}
}

func TestListSkills(t *testing.T) {
	dir := t.TempDir()
	storeDir := filepath.Join(dir, "store")
	m := NewManager(storeDir)

	_, _, _ = m.EnsureSkillDir("alpha")
	_, _, _ = m.EnsureSkillDir("beta")
	_, _, _ = m.EnsureSkillDir("gamma")

	// Also create a regular file (should be ignored)
	_ = os.WriteFile(filepath.Join(storeDir, "not-a-dir.txt"), []byte("hi"), 0o644)

	skills, err := m.ListSkills()
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(skills)
	if len(skills) != 3 {
		t.Fatalf("expected 3 skills, got %d: %v", len(skills), skills)
	}
	if skills[0] != "alpha" || skills[1] != "beta" || skills[2] != "gamma" {
		t.Errorf("unexpected skills: %v", skills)
	}
}

func TestListSkills_EmptyStore(t *testing.T) {
	skills, err := NewManager("/nonexistent/path").ListSkills()
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 0 {
		t.Errorf("expected empty list, got %v", skills)
	}
}

func TestLinkSkillToCLI(t *testing.T) {
	dir := t.TempDir()
	storeDir := filepath.Join(dir, "store")
	cliDir := filepath.Join(dir, "cli-skills")
	m := NewManager(storeDir)

	// Create a skill in the store
	skillPath, _, _ := m.EnsureSkillDir("link-test")
	_ = os.WriteFile(filepath.Join(skillPath, "SKILL.md"), []byte("test"), 0o644)

	// Link it
	if err := m.LinkSkillToCLI("link-test", cliDir); err != nil {
		t.Fatal(err)
	}

	// Verify symlink exists
	linkPath := filepath.Join(cliDir, "link-test")
	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatal("symlink not created")
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("expected a symlink")
	}

	// Verify it points to the right place
	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatal(err)
	}
	if target != skillPath {
		t.Errorf("symlink target = %q, want %q", target, skillPath)
	}
}

func TestLinkSkillToCLI_EmptyDir(t *testing.T) {
	m := NewManager(t.TempDir())
	// Empty CLI dir should be a no-op
	if err := m.LinkSkillToCLI("something", ""); err != nil {
		t.Errorf("expected nil for empty cliDir, got %v", err)
	}
}
