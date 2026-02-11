package components

func Sparkline(points []int64) string {
	if len(points) == 0 {
		return ""
	}
	blocks := []rune("▁▂▃▄▅▆▇█")
	var max int64 = 1
	for _, p := range points {
		if p > max {
			max = p
		}
	}
	out := make([]rune, 0, len(points))
	for _, p := range points {
		idx := int((p * int64(len(blocks)-1)) / max)
		out = append(out, blocks[idx])
	}
	return string(out)
}
