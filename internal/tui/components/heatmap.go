package components

import (
	"fmt"
	"strings"
)

func WeeklyHeatmap(values map[string]int64) string {
	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	lines := make([]string, 0, len(days))
	for _, d := range days {
		count := values[d]
		blocks := strings.Repeat("▓", int(count))
		lines = append(lines, fmt.Sprintf("%s %s", d, blocks))
	}
	return strings.Join(lines, "\n")
}
