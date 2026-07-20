package segments

// Segment provides a status bar segment value.
type Segment interface {
	// Name returns the segment identifier.
	Name() string
	// Render returns the formatted segment string for the tmux status bar.
	Render() string
}
