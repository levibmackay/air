package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the current session's provider, elapsed time, cost, and tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("status: not yet implemented (router lands in Phase 2)")
		},
	}
}
