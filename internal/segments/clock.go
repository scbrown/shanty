package segments

import (
	"fmt"
	"time"
)

// Clock renders the current time as HH:MM.
type Clock struct{}

func (c Clock) Name() string {
	return "clock"
}

func (c Clock) Render() string {
	return fmt.Sprintf("#[fg=#8be9fd]%s#[default]", time.Now().Format("15:04"))
}
