package segments

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/scbrown/shanty/internal/stread"
)

func TestClockRender(t *testing.T) {
	c := Clock{}
	if c.Name() != "clock" {
		t.Errorf("expected name 'clock', got %q", c.Name())
	}
	result := c.Render()
	// Should contain HH:MM format wrapped in Dracula cyan tmux color
	if !strings.Contains(result, "#[fg=#8be9fd]") {
		t.Errorf("expected Dracula cyan color code, got %q", result)
	}
	if !strings.Contains(result, ":") {
		t.Errorf("expected time with colon, got %q", result)
	}
}

func TestHostRender(t *testing.T) {
	h := Host{}
	if h.Name() != "host" {
		t.Errorf("expected name 'host', got %q", h.Name())
	}
	result := h.Render()
	if result == "" {
		t.Error("expected non-empty hostname")
	}
	// Should be wrapped in Dracula green tmux color
	if !strings.Contains(result, "#[fg=#50fa7b]") {
		t.Errorf("expected Dracula green color code, got %q", result)
	}
}

func TestLoadRender(t *testing.T) {
	l := Load{}
	if l.Name() != "load" {
		t.Errorf("expected name 'load', got %q", l.Name())
	}
	result := l.Render()
	// Should be a number like "0.5" or "n/a"
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestMemRender(t *testing.T) {
	m := Mem{}
	if m.Name() != "mem" {
		t.Errorf("expected name 'mem', got %q", m.Name())
	}
	result := m.Render()
	// Should contain a percentage or "n/a"
	if result == "" {
		t.Error("expected non-empty result")
	}
	if result != "n/a" && !strings.Contains(result, "%") {
		t.Errorf("expected percentage in result, got %q", result)
	}
}

func TestDiskRender(t *testing.T) {
	d := Disk{}
	if d.Name() != "disk" {
		t.Errorf("expected name 'disk', got %q", d.Name())
	}
	result := d.Render()
	if result == "" {
		t.Error("expected non-empty result")
	}
	if result != "n/a" && !strings.Contains(result, "%") {
		t.Errorf("expected percentage in result, got %q", result)
	}
}

func TestColorForPercent(t *testing.T) {
	tests := []struct {
		pct      float64
		expected string
	}{
		{0, "#50fa7b"},   // green
		{49, "#50fa7b"},  // green
		{50, "#ffb86c"},  // orange
		{79, "#ffb86c"},  // orange
		{80, "#ff5555"},  // red
		{100, "#ff5555"}, // red
	}
	for _, tt := range tests {
		got := colorForPercent(tt.pct)
		if got != tt.expected {
			t.Errorf("colorForPercent(%.0f) = %q, want %q", tt.pct, got, tt.expected)
		}
	}
}

func TestRegistryContainsAllSegments(t *testing.T) {
	expected := []string{"session", "clock", "host", "cpu", "mem", "load", "disk",
		"crewid", "task", "stats", "anchor", "crew", "events", "inbox", "harness"}
	for _, name := range expected {
		if _, ok := Registry[name]; !ok {
			t.Errorf("Registry missing segment %q", name)
		}
	}
}

func TestAllNamesMatchesRegistry(t *testing.T) {
	names := AllNames()
	if len(names) != len(Registry) {
		t.Errorf("AllNames() has %d entries, Registry has %d", len(names), len(Registry))
	}
	for _, name := range names {
		if _, ok := Registry[name]; !ok {
			t.Errorf("AllNames() contains %q but it's not in Registry", name)
		}
	}
}

func shantytownSegments() []Segment {
	return []Segment{CrewID{}, Task{}, Stats{}, Anchor{}, Crew{}, Events{}, Inbox{}, Harness{}}
}

func TestShantytownSegmentsEmptyWithoutST(t *testing.T) {
	// Without an st binary anywhere, shantytown segments render empty. This is
	// the ONE silent-empty the segments allow: shanty works without shantytown,
	// and that user must not be warned about a tool they never installed.
	//
	// SHANTY_ST_BIN pointing at nothing is how we force that state — Bin() treats
	// an override it cannot resolve as "no st" rather than searching on, so this
	// holds even on a host where st is installed.
	t.Setenv("SHANTY_ST_BIN", filepath.Join(t.TempDir(), "definitely-not-st"))
	stread.ResetBinForTest()
	defer stread.ResetBinForTest()

	for _, seg := range shantytownSegments() {
		if got := seg.Render(); got != "" {
			t.Errorf("segment %q returned %q with no st installed, want empty", seg.Name(), got)
		}
	}
}

func TestPerAgentSegmentsAreLoudWithoutIdentity(t *testing.T) {
	// The contract this asserts REPLACED an earlier one that rendered empty when
	// the identity could not be resolved. Empty was the bug: a bar segment that
	// blanks looks like "nothing to report" and gets believed, so an entire fleet
	// bar sat blank and read as healthy. With st present and no identity
	// derivable, every per-agent segment must say something.
	if !stread.Installed() {
		t.Skip("no st on this host — the loud branch requires st to be installed")
	}
	t.Setenv("SHANTY_AGENT", "")
	SetSession("") // no prefix to derive from either
	defer SetSession("")

	for _, seg := range shantytownSegments() {
		if seg.Name() == "crew" {
			continue // crew is fleet-wide; it needs no identity
		}
		got := seg.Render()
		if got == "" {
			t.Errorf("segment %q rendered empty with no identity — must be loud", seg.Name())
		}
		if !strings.Contains(got, "⚠") {
			t.Errorf("segment %q rendered %q with no identity, want a ⚠ marker", seg.Name(), got)
		}
	}
}

func TestHarnessRendersNameNotDuration(t *testing.T) {
	// Regression guard: this segment replaced one that rendered hours as
	// "4h". A harness is a name — no unit may be appended.
	t.Setenv("SHANTY_AGENT", "")
	if got := (Harness{}).Render(); strings.HasSuffix(got, "h#[default]") {
		t.Errorf("harness rendered a duration-style suffix: %q", got)
	}
}

func TestClipKeepsRunesWhole(t *testing.T) {
	// A byte-slice through a multi-byte character would put a broken glyph in the
	// status line, so the clip counts runes.
	long := "réécrire le cache très très très très long"
	got := clip(long, 12)
	if []rune(got)[len([]rune(got))-1] != '…' {
		t.Errorf("clip(%q) = %q, want a trailing ellipsis", long, got)
	}
	if n := len([]rune(got)); n != 12 {
		t.Errorf("clip produced %d runes, want 12: %q", n, got)
	}
	if short := clip("short", 12); short != "short" {
		t.Errorf("clip left a short string alone? got %q", short)
	}
}

func TestHuman(t *testing.T) {
	for n, want := range map[int]string{
		0: "0", 950: "950", 1000: "1k", 1500: "1.5k",
		12_000: "12k", 1_000_000: "1M", 3_450_000: "3.5M",
	} {
		if got := human(n); got != want {
			t.Errorf("human(%d) = %q, want %q", n, got, want)
		}
	}
}
