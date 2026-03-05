package internal

var sparks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// Sparkline 将一组数值渲染为 Unicode 迷你折线（如 "▁▂▃▅▇▆▃▂"）。
// 适合内嵌在表格单元格中展示趋势。空切片返回空字符串。
func Sparkline(values []float64) string {
	if len(values) == 0 {
		return ""
	}
	min, max := values[0], values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	rng := max - min
	if rng == 0 {
		rng = 1
	}
	out := make([]rune, len(values))
	for i, v := range values {
		idx := int((v - min) / rng * float64(len(sparks)-1))
		if idx >= len(sparks) {
			idx = len(sparks) - 1
		}
		out[i] = sparks[idx]
	}
	return string(out)
}
