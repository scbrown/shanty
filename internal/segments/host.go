package segments

import (
	"fmt"
	"os"
)

// Host renders the hostname.
type Host struct{}

func (h Host) Name() string {
	return "host"
}

func (h Host) Render() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return fmt.Sprintf("#[fg=#50fa7b]%s#[default]", name)
}
