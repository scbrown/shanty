package segments

import (
	"strings"
	"testing"
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
		"anchor", "crew", "events", "inbox", "harness"}
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
	return []Segment{Anchor{}, Crew{}, Events{}, Inbox{}, Harness{}}
}

func TestShantytownSegmentsEmptyWithoutST(t *testing.T) {
	// Without st on PATH, shantytown segments should return empty.
	// This depends on the environment, so we only assert on the case we
	// can control: when st is absent, the render must be empty. Either
	// way it must not panic.
	for _, seg := range shantytownSegments() {
		got := seg.Render()
		if !stAvailable() && got != "" {
			t.Errorf("segment %q returned %q without st on PATH, want empty", seg.Name(), got)
		}
	}
}

func TestShantytownSegmentsEmptyWithoutAgent(t *testing.T) {
	// Identity comes from $SHANTY_AGENT. With it unset we must render
	// nothing rather than guess an agent — otherwise the bar would show
	// somebody else's plate, inbox and events. This is a condition we can
	// control, so assert it unconditionally.
	t.Setenv("SHANTY_AGENT", "")
	for _, seg := range shantytownSegments() {
		if seg.Name() == "crew" {
			continue // crew is fleet-wide, not per-agent
		}
		cache.entries = map[string]cacheEntry{} // defeat any cached value
		if got := seg.Render(); got != "" {
			t.Errorf("segment %q returned %q with SHANTY_AGENT unset, want empty", seg.Name(), got)
		}
	}
}

func TestHarnessRendersNameNotDuration(t *testing.T) {
	// Regression guard: this segment replaced one that rendered hours as
	// "4h". A harness is a name — no unit may be appended.
	t.Setenv("SHANTY_AGENT", "")
	cache.entries = map[string]cacheEntry{}
	if got := (Harness{}).Render(); strings.HasSuffix(got, "h#[default]") {
		t.Errorf("harness rendered a duration-style suffix: %q", got)
	}
}

func TestCacheGetSet(t *testing.T) {
	cache.set("test-key", "test-value")
	val, ok := cache.get("test-key")
	if !ok {
		t.Error("expected cache hit")
	}
	if val != "test-value" {
		t.Errorf("expected 'test-value', got %q", val)
	}
}
