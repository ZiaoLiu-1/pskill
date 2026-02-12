package skill

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestParse_WithFrontmatter(t *testing.T) {
	raw := []byte("---\nname: my-skill\ndescription: A test skill\n---\n\n# My Skill\n\nBody content here.\n")
	sk, err := Parse(raw, "/tmp/my-skill/SKILL.md", "cursor")
	if err != nil {
		t.Fatal(err)
	}
	if sk.Name != "my-skill" {
		t.Errorf("expected name 'my-skill', got %q", sk.Name)
	}
	if sk.Description != "A test skill" {
		t.Errorf("expected description 'A test skill', got %q", sk.Description)
	}
	if sk.SourceCLI != "cursor" {
		t.Errorf("expected sourceCLI 'cursor', got %q", sk.SourceCLI)
	}
	if sk.Body != "# My Skill\n\nBody content here." {
		t.Errorf("unexpected body: %q", sk.Body)
	}
}

func TestParse_WithoutFrontmatter(t *testing.T) {
	raw := []byte("# Just a plain skill\n\nNo frontmatter here.\n")
	sk, err := Parse(raw, "/tmp/plain-skill/SKILL.md", "claude")
	if err != nil {
		t.Fatal(err)
	}
	// Should fall back to directory name
	if sk.Name != "plain-skill" {
		t.Errorf("expected fallback name 'plain-skill', got %q", sk.Name)
	}
	if sk.Description != "" {
		t.Errorf("expected empty description, got %q", sk.Description)
	}
}

func TestParse_EmptyFrontmatterName(t *testing.T) {
	raw := []byte("---\ndescription: No name field\n---\n\nBody.\n")
	sk, err := Parse(raw, "/home/user/.cursor/skills/auto-named/SKILL.md", "")
	if err != nil {
		t.Fatal(err)
	}
	if sk.Name != "auto-named" {
		t.Errorf("expected fallback name 'auto-named', got %q", sk.Name)
	}
}

func TestFallbackName(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/home/user/.cursor/skills/frontend-design/SKILL.md", "frontend-design"},
		{filepath.Join("skills", "my-skill", "SKILL.md"), "my-skill"},
		{"SKILL.md", "unknown-skill"},
	}
	for _, tt := range tests {
		got := fallbackName(tt.path)
		if got != tt.want {
			t.Errorf("fallbackName(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestInferTags(t *testing.T) {
	tags := inferTags("frontend-design", "Create polished UI components", "")
	if len(tags) == 0 {
		t.Fatal("expected at least one tag")
	}
	// Tags shorter than 4 chars should be excluded
	for _, tag := range tags {
		if len(tag) < 4 {
			t.Errorf("tag %q is shorter than 4 chars", tag)
		}
	}
	// Should contain recognizable tokens
	sort.Strings(tags)
	found := false
	for _, tag := range tags {
		if tag == "frontend-design" || tag == "polished" || tag == "create" || tag == "components" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected meaningful tags, got %v", tags)
	}
}

func TestParseFile(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "test-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: test-skill\ndescription: A test\n---\n\n# Test\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	sk, err := ParseFile(filepath.Join(skillDir, "SKILL.md"), "codex")
	if err != nil {
		t.Fatal(err)
	}
	if sk.Name != "test-skill" {
		t.Errorf("expected 'test-skill', got %q", sk.Name)
	}
}

func TestParseFile_NotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/SKILL.md", "")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
