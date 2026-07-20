package cmd

import (
	"fmt"

	"github.com/scbrown/shanty/internal/session"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List shanty sessions",
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
		for _, s := range sessions {
			fmt.Println(s)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
}
