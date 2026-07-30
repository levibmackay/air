package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCheckpointsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "checkpoints",
		Short: "List saved checkpoints for the current or a given session",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("checkpoints: not yet implemented (checkpoint store lands in Phase 2)")
		},
	}
}
