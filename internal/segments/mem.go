package segments

import "fmt"

// Mem renders memory usage percentage with color coding.
//
// The READING is in sysread.go, per platform. See CPU for why the two are apart.
type Mem struct{}

func (m Mem) Name() string {
	return "mem"
}

func (m Mem) Render() string {
	pct, ok := memPercent()
	if !ok {
		return "mem n/a"
	}
	color := colorForPercent(pct)
	return fmt.Sprintf("#[fg=%s]mem %d%%#[default]", color, int(pct))
}
