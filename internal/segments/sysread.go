package segments

import (
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// Reading the machine's own vital signs, on more than one kind of machine.
//
// WHY THIS FILE EXISTS. The cpu and mem segments were written against /proc,
// which is Linux-only, and their failure path returned the string "n/a". On a
// BSD userland — the operator's Mac, which runs a crew of its own — that meant
// the bar permanently displayed "cpu n/a mem n/a". Neither segment was broken in
// any way a reader could act on; they were simply asking a kernel that does not
// answer, and saying "n/a" as though the machine had declined to say.
//
// "n/a" IS STILL THE RIGHT ANSWER WHEN NOTHING CAN BE READ, and it is kept for
// exactly that. What changed is that it now means "I asked this platform and it
// would not tell me", rather than "I only know how to ask Linux".
//
// SHELLING OUT IS DELIBERATE ON DARWIN. The numbers live behind Mach calls
// (host_statistics) that pure Go cannot reach without cgo, and cgo would make
// this tool's build depend on a C toolchain for two status-bar figures. `vm_stat`
// and `top` are in the base system, present on every macOS, and the cost is one
// exec per refresh — 3ms for vm_stat. `top` is the expensive one at ~0.6s, which
// is why it is only reached for cpu, whose Linux path already spends 200ms
// sleeping between samples for the same reason: a single sample of a counter is
// not a rate.

// cpuPercent is the machine's CPU utilisation, and whether it could be read.
//
// The bool is not decoration. Returning 0 for "could not tell" would paint the
// bar green and calm on a machine that might be pinned — the failure direction a
// status bar must never take, since its whole job is to be glanceable.
func cpuPercent() (float64, bool) {
	switch runtime.GOOS {
	case "darwin":
		return cpuPercentDarwin()
	default:
		return cpuPercentProc()
	}
}

// cpuPercentProc: two samples of /proc/stat, 200ms apart. The original
// implementation, unchanged in behaviour.
func cpuPercentProc() (float64, bool) {
	idle1, total1 := readCPUStat()
	if total1 == 0 {
		return 0, false
	}
	sleepBetweenCPUSamples()
	idle2, total2 := readCPUStat()
	if total2 == total1 {
		return 0, true // genuinely idle: the counters did not move
	}
	idleDelta := float64(idle2 - idle1)
	totalDelta := float64(total2 - total1)
	return (1.0 - idleDelta/totalDelta) * 100.0, true
}

var cpuUsageLine = regexp.MustCompile(`([0-9.]+)%\s+idle`)

// cpuPercentDarwin asks `top` for TWO samples and reads the second.
//
// The second, and only the second. `top -l 1` reports usage averaged since boot,
// which on a machine that has been up for days is a flat, meaningless number that
// never moves — the same defect as reading /proc/stat once. `-l 2` makes the
// second sample an interval measurement, which is the figure being asked for.
// `-n 0 -s 0` suppresses the process table and the delay between samples.
func cpuPercentDarwin() (float64, bool) {
	out, err := exec.Command("top", "-l", "2", "-n", "0", "-s", "0").Output()
	if err != nil {
		return 0, false
	}
	var last string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "CPU usage:") {
			last = line
		}
	}
	m := cpuUsageLine.FindStringSubmatch(last)
	if m == nil {
		return 0, false
	}
	idle, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, false
	}
	return 100.0 - idle, true
}

// memPercent is the machine's memory utilisation, and whether it could be read.
func memPercent() (float64, bool) {
	switch runtime.GOOS {
	case "darwin":
		return memPercentDarwin()
	default:
		return memPercentProc()
	}
}

func memPercentProc() (float64, bool) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, false
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
		return 0, false
	}
	return float64(total-available) / float64(total) * 100.0, true
}

var vmStatPage = regexp.MustCompile(`page size of (\d+) bytes`)
var vmStatLine = regexp.MustCompile(`^(.*?):\s+(\d+)\.?$`)

// memPercentDarwin: hw.memsize for the total, vm_stat for what is in use.
//
// USED = active + wired + compressed, which is what Activity Monitor calls
// "Memory Used". Free and inactive are deliberately NOT counted: macOS fills
// otherwise-idle RAM with reclaimable file cache by design, so counting inactive
// as used would show a healthy machine at 95% forever and train the operator to
// ignore the number — the same "fires constantly on a benign condition" failure
// the fleet's own warning texts were shortened to avoid.
func memPercentDarwin() (float64, bool) {
	totalOut, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return 0, false
	}
	total, err := strconv.ParseUint(strings.TrimSpace(string(totalOut)), 10, 64)
	if err != nil || total == 0 {
		return 0, false
	}
	out, err := exec.Command("vm_stat").Output()
	if err != nil {
		return 0, false
	}
	text := string(out)
	pageSize := uint64(4096)
	if m := vmStatPage.FindStringSubmatch(text); m != nil {
		if v, err := strconv.ParseUint(m[1], 10, 64); err == nil && v > 0 {
			pageSize = v
		}
	}
	pages := map[string]uint64{}
	for _, line := range strings.Split(text, "\n") {
		m := vmStatLine.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		if v, err := strconv.ParseUint(m[2], 10, 64); err == nil {
			pages[strings.TrimSpace(m[1])] = v
		}
	}
	active, okA := pages["Pages active"]
	wired, okW := pages["Pages wired down"]
	if !okA || !okW {
		return 0, false
	}
	compressed := pages["Pages occupied by compressor"]
	used := (active + wired + compressed) * pageSize
	if used > total {
		used = total
	}
	return float64(used) / float64(total) * 100.0, true
}
