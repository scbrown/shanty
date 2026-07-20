package segments

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// CPU renders CPU usage percentage with color coding.
// green (<50%), orange (<80%), red (>=80%).
type CPU struct{}

func (c CPU) Name() string {
	return "cpu"
}

func (c CPU) Render() string {
	idle1, total1 := readCPUStat()
	if total1 == 0 {
		return "n/a"
	}
	time.Sleep(200 * time.Millisecond)
	idle2, total2 := readCPUStat()
	if total2 == total1 {
		return "0%"
	}

	idleDelta := float64(idle2 - idle1)
	totalDelta := float64(total2 - total1)
	usage := (1.0 - idleDelta/totalDelta) * 100.0

	color := colorForPercent(usage)
	return fmt.Sprintf("#[fg=%s]%d%%#[default]", color, int(usage))
}

func readCPUStat() (idle, total uint64) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) < 5 {
				return 0, 0
			}
			for i := 1; i < len(fields); i++ {
				val, _ := strconv.ParseUint(fields[i], 10, 64)
				total += val
			}
			idleVal, _ := strconv.ParseUint(fields[4], 10, 64)
			return idleVal, total
		}
	}
	return 0, 0
}
