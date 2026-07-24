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
	"crewid":  CrewID{}, // who this pane is: mark, name, role, state
	"task":    Task{},   // what they hold: item id + title
	"stats":   Stats{},  // what they did: activity, files, tokens
	"anchor":  Anchor{}, // just the item id (superseded by task)
	"crew":    Crew{},
	"events":  Events{},
	"inbox":   Inbox{},
	"harness": Harness{},
}

// AllNames returns all registered segment names in display order.
func AllNames() []string {
	return []string{
		"session", "clock", "host", "cpu", "mem", "load", "disk",
		"crewid", "task", "stats", "anchor", "crew", "events", "inbox", "harness",
	}
}
