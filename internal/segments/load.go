package segments

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Load renders the 1-minute load average.
type Load struct{}

func (l Load) Name() string {
	return "load"
}

func (l Load) Render() string {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return "n/a"
	}
	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return "n/a"
	}
	load, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return "n/a"
	}
	return fmt.Sprintf("%.1f", load)
}
