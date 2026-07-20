package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/scbrown/shanty/internal/segments"
	"github.com/spf13/cobra"
)

var segCmd = &cobra.Command{
	Use:   "seg <name>",
	Short: "Render a status bar segment",
	Long: `Render a single status bar segment for tmux.

tmux calls this via #(shanty seg <name>) at status-interval.
Each segment outputs a short formatted string with optional tmux color codes.

System segments: clock, host, cpu, mem, load, disk
shantytown segments: anchor, crew, events, inbox, harness
These require the st CLI on PATH and render empty without it. All but crew also
need $SHANTY_AGENT set, and render empty when it is not.
See https://github.com/scbrown/shantytown`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Optional second arg: the tmux session this render is for, passed from
		// #{session_name}. It is the per-pane identity fallback for the shared
		// fleet bar — a segment learns which agent it draws from the session it
		// is drawn in when $SHANTY_AGENT is not exported per pane.
		if len(args) == 2 {
			segments.SetSession(args[1])
		}

		if name == "list" {
			for _, n := range segments.AllNames() {
				fmt.Println(n)
			}
			return nil
		}

		seg, ok := segments.Registry[name]
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown segment: %s\nAvailable: %s\n",
				name, strings.Join(segments.AllNames(), ", "))
			return fmt.Errorf("unknown segment: %s", name)
		}

		result := seg.Render()
		if result != "" {
			fmt.Print(result)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(segCmd)
}
