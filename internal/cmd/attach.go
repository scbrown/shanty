package cmd

import (
	"github.com/scbrown/shanty/internal/session"
	"github.com/spf13/cobra"
)

var attachCmd = &cobra.Command{
	Use:   "attach [name]",
	Short: "Attach to an existing shanty session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := session.NewManager()
		return mgr.Attach(args[0])
	},
}

func init() {
	rootCmd.AddCommand(attachCmd)
}
