package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRollbackCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rollback [checkpoint-id]",
		Short: "Restore the working tree to a prior checkpoint (asks for confirmation)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("rollback: not yet implemented (checkpoint store lands in Phase 2)")
		},
	}
}
