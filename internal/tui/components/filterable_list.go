package components

import "strings"

type FilterItem struct {
	Title string
	Desc  string
}

func Filter(items []FilterItem, q string) []FilterItem {
	if strings.TrimSpace(q) == "" {
		return items
	}
	q = strings.ToLower(q)
	out := make([]FilterItem, 0, len(items))
	for _, it := range items {
		if strings.Contains(strings.ToLower(it.Title), q) || strings.Contains(strings.ToLower(it.Desc), q) {
			out = append(out, it)
		}
	}
	return out
}
