package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Cache struct {
	cacheDir string
}

type cacheEntry struct {
	StoredAt time.Time       `json:"storedAt"`
	Data     json.RawMessage `json:"data"`
}

func NewCache(cacheDir string) *Cache {
	_ = os.MkdirAll(cacheDir, 0o755)
	return &Cache{cacheDir: cacheDir}
}

func (c *Cache) Store(key string, value interface{}) {
	raw, _ := json.Marshal(value)
	entry := cacheEntry{StoredAt: time.Now(), Data: raw}
	payload, _ := json.Marshal(entry)
	_ = os.WriteFile(filepath.Join(c.cacheDir, sanitizeFileName(key)+".json"), payload, 0o644)
}

func (c *Cache) Load(key string, ttl time.Duration) ([]byte, bool) {
	path := filepath.Join(c.cacheDir, sanitizeFileName(key)+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var entry cacheEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		return nil, false
	}
	if time.Since(entry.StoredAt) > ttl {
		return nil, false
	}
	return entry.Data, true
}

func sanitizeFileName(in string) string {
	out := []rune(in)
	for i, r := range out {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			out[i] = '_'
		}
	}
	return string(out)
}
