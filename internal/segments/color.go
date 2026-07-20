package segments

// colorForPercent returns a Dracula palette color based on percentage thresholds.
// green (<50%), orange (<80%), red (>=80%).
func colorForPercent(pct float64) string {
	switch {
	case pct >= 80:
		return "#ff5555" // Dracula red
	case pct >= 50:
		return "#ffb86c" // Dracula orange
	default:
		return "#50fa7b" // Dracula green
	}
}
