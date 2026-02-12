package registry

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCache_StoreAndLoad(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(dir)

	data := map[string]string{"hello": "world"}
	c.Store("test-key", data)

	raw, ok := c.Load("test-key", 5*time.Minute)
	if !ok {
		t.Fatal("expected cache hit")
	}

	var got map[string]string
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got["hello"] != "world" {
		t.Errorf("expected 'world', got %q", got["hello"])
	}
}

func TestCache_Expired(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(dir)

	c.Store("expired-key", "value")

	// Load with a TTL of 0 (already expired)
	_, ok := c.Load("expired-key", 0)
	if ok {
		t.Error("expected cache miss for expired entry")
	}
}

func TestCache_Miss(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(dir)

	_, ok := c.Load("nonexistent", 5*time.Minute)
	if ok {
		t.Error("expected cache miss")
	}
}

func TestSanitizeFileName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"simple", "simple"},
		{"with spaces", "with_spaces"},
		{"search_react_10", "search_react_10"},
		{"a/b\\c:d", "a_b_c_d"},
		{"UPPER-case_123", "UPPER-case_123"},
		{"日本語", "___"},
	}
	for _, tt := range tests {
		got := sanitizeFileName(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeFileName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
