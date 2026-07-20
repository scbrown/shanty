package segments

// Registry maps segment names to their implementations.
var Registry = map[string]Segment{
	// Core segments
	"session": Session{},
	// System segments
	"clock": Clock{},
	"host":  Host{},
	"cpu":   CPU{},
	"mem":   Mem{},
	"load":  Load{},
	"disk":  Disk{},
	// shantytown segments (https://github.com/scbrown/shantytown) — empty without the st CLI
	"anchor":  Anchor{},
	"crew":    Crew{},
	"events":  Events{},
	"inbox":   Inbox{},
	"harness": Harness{},
}

// AllNames returns all registered segment names in display order.
func AllNames() []string {
	return []string{
		"session", "clock", "host", "cpu", "mem", "load", "disk",
		"anchor", "crew", "events", "inbox", "harness",
	}
}
