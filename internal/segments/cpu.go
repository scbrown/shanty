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
//
// The READING is in sysread.go, per platform; this type only decides how to say
// it. Splitting them is what let the same colour rule and the same "n/a" apply to
// a Linux /proc sample and a Darwin `top` sample without either knowing about the
// other.
type CPU struct{}

func (c CPU) Name() string {
	return "cpu"
}

func (c CPU) Render() string {
	usage, ok := cpuPercent()
	if !ok {
		// Still "n/a", and still the honest answer — but now it means every
		// source this platform has was asked and none of them answered, rather
		// than "we only know how to read /proc".
		return "cpu n/a"
	}
	color := colorForPercent(usage)
	return fmt.Sprintf("#[fg=%s]cpu %d%%#[default]", color, int(usage))
}

// sleepBetweenCPUSamples is the gap between the two /proc/stat reads. A single
// read of a monotonically increasing counter is a total, not a rate; the delta is
// the measurement. Named so a test can see the cost rather than guess at it.
func sleepBetweenCPUSamples() { time.Sleep(200 * time.Millisecond) }

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
