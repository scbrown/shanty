package segments

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Mem renders memory usage percentage with color coding.
type Mem struct{}

func (m Mem) Name() string {
	return "mem"
}

func (m Mem) Render() string {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return "n/a"
	}

	var total, available uint64
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		val, _ := strconv.ParseUint(fields[1], 10, 64)
		switch fields[0] {
		case "MemTotal:":
			total = val
		case "MemAvailable:":
			available = val
		}
	}

	if total == 0 {
		return "n/a"
	}

	used := total - available
	pct := float64(used) / float64(total) * 100.0
	color := colorForPercent(pct)
	return fmt.Sprintf("#[fg=%s]%d%%#[default]", color, int(pct))
}
