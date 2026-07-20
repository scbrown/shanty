// shanty is a terminal multiplexer wrapper with Dracula theme and agent pane control.
package main

import (
	"os"

	"github.com/scbrown/shanty/internal/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
