package search

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/blevesearch/bleve/v2"

	"github.com/ZiaoLiu-1/pskill/internal/skill"
)

type Result struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Score       float64 `json:"score"`
}

type indexedSkill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Body        string `json:"body"`
	Tags        string `json:"tags"`
}

type Engine struct {
	indexDir string
}

func NewEngine(indexDir string) *Engine {
	return &Engine{indexDir: indexDir}
}

func (e *Engine) IndexSkill(sk skill.Skill) error {
	idx, err := e.openOrCreate()
	if err != nil {
		return err
	}
	defer idx.Close()
	doc := indexedSkill{
		Name:        sk.Name,
		Description: sk.Description,
		Body:        sk.Body,
		Tags:        strings.Join(sk.Tags, " "),
	}
	return idx.Index(sk.Name, doc)
}

func (e *Engine) IndexSkillByPath(name, dir string) error {
	path := filepath.Join(dir, "SKILL.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	sk, err := skill.Parse(raw, path, "")
	if err != nil {
		return err
	}
	sk.Name = name
	return e.IndexSkill(sk)
}

func (e *Engine) Search(query string, limit int) ([]Result, error) {
	idx, err := e.openOrCreate()
	if err != nil {
		return nil, err
	}
	defer idx.Close()
	q := bleve.NewQueryStringQuery(query)
	req := bleve.NewSearchRequestOptions(q, limit, 0, false)
	req.Fields = []string{"name", "description"}
	resp, err := idx.Search(req)
	if err != nil {
		return nil, err
	}
	out := make([]Result, 0, len(resp.Hits))
	for _, h := range resp.Hits {
		item := Result{
			Name:        h.ID,
			Description: asString(h.Fields["description"]),
			Score:       h.Score,
		}
		out = append(out, item)
	}
	return out, nil
}

func (e *Engine) openOrCreate() (bleve.Index, error) {
	_ = os.MkdirAll(e.indexDir, 0o755)
	idx, err := bleve.Open(e.indexDir)
	if err == nil {
		return idx, nil
	}
	mapping := bleve.NewIndexMapping()
	return bleve.New(e.indexDir, mapping)
}

func asString(v interface{}) string {
	s, _ := v.(string)
	return s
}
