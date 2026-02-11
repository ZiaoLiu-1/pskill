package skill

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseFile(path string, sourceCLI string) (Skill, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Skill{}, err
	}
	return Parse(raw, path, sourceCLI)
}

func Parse(raw []byte, path string, sourceCLI string) (Skill, error) {
	txt := string(raw)
	fm := Frontmatter{}
	body := txt

	if strings.HasPrefix(txt, "---\n") {
		parts := strings.SplitN(txt, "\n---\n", 2)
		if len(parts) == 2 {
			yamlPart := strings.TrimPrefix(parts[0], "---\n")
			_ = yaml.Unmarshal([]byte(yamlPart), &fm)
			body = parts[1]
		}
	}

	name := fm.Name
	if name == "" {
		name = fallbackName(path)
	}
	return Skill{
		Name:        name,
		Description: fm.Description,
		Body:        strings.TrimSpace(body),
		Path:        path,
		SourceCLI:   sourceCLI,
		Tags:        inferTags(name, fm.Description, body),
	}, nil
}

func fallbackName(path string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(dir)
	if base == "." || base == "/" {
		return "unknown-skill"
	}
	return base
}

func inferTags(values ...string) []string {
	set := map[string]struct{}{}
	for _, v := range values {
		low := strings.ToLower(v)
		for _, tok := range strings.FieldsFunc(low, func(r rune) bool {
			return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-')
		}) {
			if len(tok) < 4 {
				continue
			}
			set[tok] = struct{}{}
		}
	}
	tags := make([]string, 0, len(set))
	for k := range set {
		tags = append(tags, k)
	}
	return tags
}
