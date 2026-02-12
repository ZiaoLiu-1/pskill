package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// SkillResult is the unified result type used across the app.
type SkillResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Author      string `json:"author"`
	Description string `json:"description"`
	GithubURL   string `json:"githubUrl"`
	SkillURL    string `json:"skillUrl"`
	Stars       int64  `json:"stars"`
	UpdatedAt   int64  `json:"updatedAt"`
	Score       float64 `json:"score,omitempty"` // AI search score
}

// --- API response structures ---

type searchResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Skills     []SkillResult `json:"skills"`
		Pagination struct {
			Page       int  `json:"page"`
			Limit      int  `json:"limit"`
			Total      int  `json:"total"`
			TotalPages int  `json:"totalPages"`
			HasNext    bool `json:"hasNext"`
		} `json:"pagination"`
	} `json:"data"`
	Error *apiError `json:"error,omitempty"`
}

type aiSearchResult struct {
	FileID   string      `json:"file_id"`
	Filename string      `json:"filename"`
	Score    float64     `json:"score"`
	Skill    *SkillResult `json:"skill,omitempty"`
}

type aiSearchResponse struct {
	Success bool `json:"success"`
	Data    struct {
		SearchQuery string           `json:"search_query"`
		Data        []aiSearchResult `json:"data"`
	} `json:"data"`
	Error *apiError `json:"error,omitempty"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Client talks to the skillsmp.com API.
type Client struct {
	baseURL string
	apiKey  string
	cache   *Cache
	http    *http.Client
}

func NewClient(baseURL, cacheDir, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		cache:   NewCache(cacheDir),
		http: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

// Search performs a keyword search via /api/v1/skills/search.
func (c *Client) Search(query string, limit int, sortBy string) ([]SkillResult, int, error) {
	if sortBy == "" {
		sortBy = "recent"
	}
	cacheKey := fmt.Sprintf("search_%s_%d_%s", query, limit, sortBy)
	if data, ok := c.cache.Load(cacheKey, 5*time.Minute); ok {
		var resp searchResponse
		if json.Unmarshal(data, &resp) == nil && resp.Success {
			return resp.Data.Skills, resp.Data.Pagination.Total, nil
		}
	}

	reqURL := fmt.Sprintf("%s/api/v1/skills/search?q=%s&limit=%d&sortBy=%s",
		c.baseURL, url.QueryEscape(query), limit, url.QueryEscape(sortBy))

	body, err := c.doGet(reqURL)
	if err != nil {
		return nil, 0, err
	}

	var resp searchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, 0, fmt.Errorf("parse response: %w", err)
	}
	if !resp.Success {
		msg := "unknown error"
		if resp.Error != nil {
			msg = resp.Error.Message
		}
		return nil, 0, fmt.Errorf("API error: %s", msg)
	}

	c.cache.Store(cacheKey, resp)
	return resp.Data.Skills, resp.Data.Pagination.Total, nil
}

// AISearch performs semantic search via /api/v1/skills/ai-search.
func (c *Client) AISearch(query string) ([]SkillResult, error) {
	cacheKey := fmt.Sprintf("ai_search_%s", query)
	if data, ok := c.cache.Load(cacheKey, 10*time.Minute); ok {
		var results []SkillResult
		if json.Unmarshal(data, &results) == nil {
			return results, nil
		}
	}

	reqURL := fmt.Sprintf("%s/api/v1/skills/ai-search?q=%s",
		c.baseURL, url.QueryEscape(query))

	body, err := c.doGet(reqURL)
	if err != nil {
		return nil, err
	}

	var resp aiSearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if !resp.Success {
		msg := "unknown error"
		if resp.Error != nil {
			msg = resp.Error.Message
		}
		return nil, fmt.Errorf("API error: %s", msg)
	}

	// Extract skills from AI search results
	results := make([]SkillResult, 0, len(resp.Data.Data))
	for _, item := range resp.Data.Data {
		if item.Skill == nil {
			continue
		}
		sk := *item.Skill
		sk.Score = item.Score
		results = append(results, sk)
	}

	c.cache.Store(cacheKey, results)
	return results, nil
}

// Trending returns skills sorted by stars (most popular).
func (c *Client) Trending(limit int, page int) ([]SkillResult, int, error) {
	return c.Search("*", limit, "stars")
}

// DownloadSkill downloads a skill's SKILL.md from its GitHub URL.
func (c *Client) DownloadSkill(name, githubURL, destination string) error {
	_ = os.MkdirAll(destination, 0o755)

	// Try to fetch raw SKILL.md from GitHub
	if githubURL != "" {
		rawURL := githubToRaw(githubURL)
		if rawURL != "" {
			req, _ := http.NewRequest(http.MethodGet, rawURL, nil)
			resp, err := c.http.Do(req)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == 200 {
					raw, err := io.ReadAll(resp.Body)
					if err == nil && len(raw) > 0 {
						return os.WriteFile(filepath.Join(destination, "SKILL.md"), raw, 0o644)
					}
				}
			}
		}
	}

	// Fallback: synthesize minimal SKILL.md
	content := fmt.Sprintf("---\nname: %s\ndescription: Installed via pskill\n---\n\n# %s\n\nInstalled from skillsmp.com registry.\n", name, name)
	return os.WriteFile(filepath.Join(destination, "SKILL.md"), []byte(content), 0o644)
}

// doGet performs an authenticated GET request.
func (c *Client) doGet(reqURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr struct {
			Error *apiError `json:"error"`
		}
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Error != nil {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, apiErr.Error.Message)
		}
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return body, nil
}

// githubToRaw converts a GitHub tree URL to a raw content URL for SKILL.md.
func githubToRaw(ghURL string) string {
	// https://github.com/user/repo/tree/branch/path
	// → https://raw.githubusercontent.com/user/repo/branch/path/SKILL.md
	u, err := url.Parse(ghURL)
	if err != nil || u.Host != "github.com" {
		return ""
	}
	// path: /user/repo/tree/branch/rest...
	parts := splitPath(u.Path)
	if len(parts) < 5 || parts[2] != "tree" {
		return ""
	}
	user := parts[0]
	repo := parts[1]
	branch := parts[3]
	rest := parts[4:]
	rawPath := fmt.Sprintf("/%s/%s/%s/%s/SKILL.md", user, repo, branch, joinPath(rest))
	return "https://raw.githubusercontent.com" + rawPath
}

func splitPath(p string) []string {
	var parts []string
	for _, s := range filepath.SplitList(p) {
		for _, seg := range split(s) {
			if seg != "" {
				parts = append(parts, seg)
			}
		}
	}
	return parts
}

func split(s string) []string {
	var parts []string
	current := ""
	for _, r := range s {
		if r == '/' {
			if current != "" {
				parts = append(parts, current)
			}
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func joinPath(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += "/"
		}
		result += p
	}
	return result
}
