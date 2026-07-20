package cmd

import (
	"fmt"
	"strings"

	"github.com/scbrown/shanty/internal/session"
	"github.com/spf13/cobra"
)

var lsPlain bool

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List shanty sessions — crew-oriented when shantytown is present",
	Long: `List shanty sessions.

When shantytown (st) is installed and reports crew state, each row is a crew
selector: name, what the agent is working on, its state (busy/idle/waiting/
saturated — st's own verdict, not a second opinion), and its settings currency.
The ones that need attention (blocked/waiting first) are sorted to the top, so
choosing who to attach to is informed, not a raw session list.

Without st, or with only personal sessions, it falls back to a plain name list.
Use --plain to force the plain list (e.g. for scripting).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := session.NewManager()
		sessions, err := mgr.List()
		if err != nil {
			return err
		}
		if len(sessions) == 0 {
			fmt.Println("No active sessions.")
			return nil
		}

		if !lsPlain {
			if rows, ok := session.EnrichedRows(sessions); ok {
				printCrewTable(rows)
				return nil
			}
		}

		// Fallback: the plain name list (and the only output under --plain).
		for _, s := range sessions {
			fmt.Println(s)
		}
		return nil
	},
}

// printCrewTable renders the crew-oriented picker. A leading marker flags the
// rows that want a human — the sort already floated them up, the marker makes
// them scannable.
func printCrewTable(rows []session.CrewRow) {
	fmt.Printf("  %-2s %-12s %-16s %-16s %s\n", "", "CREW", "STATE", "WORKING ON", "SETTINGS")
	for _, r := range rows {
		mark := " "
		if needsAttention(r.State) {
			mark = "⚠"
		}
		fmt.Printf("  %-2s %-12s %-16s %-16s %s\n",
			mark, r.Name, dash(r.State), dash(r.Item), dash(r.Currency))
	}
}

// needsAttention is a display concern only — it does not decide state, it marks
// the verdicts st already sorted to the top.
func needsAttention(stateCell string) bool {
	for _, w := range []string{"waiting", "wedged", "queued", "saturated"} {
		if strings.HasPrefix(stateCell, w) {
			return true
		}
	}
	return false
}

func dash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func init() {
	lsCmd.Flags().BoolVar(&lsPlain, "plain", false, "plain name list only (no crew enrichment)")
	rootCmd.AddCommand(lsCmd)
}
