package segments

import (
	"fmt"
	"syscall"
)

// Disk renders root partition usage percentage with color coding.
type Disk struct{}

func (d Disk) Name() string {
	return "disk"
}

func (d Disk) Render() string {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		return "n/a"
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	if total == 0 {
		return "n/a"
	}

	used := total - free
	pct := float64(used) / float64(total) * 100.0
	color := colorForPercent(pct)
	return fmt.Sprintf("#[fg=%s]%d%%#[default]", color, int(pct))
}
