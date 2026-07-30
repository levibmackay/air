package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newLogsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logs",
		Short: "Show AIR's structured logs for past and current sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("logs: not yet implemented (logging lands in Phase 2)")
		},
	}
}
