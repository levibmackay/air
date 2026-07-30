package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run [objective]",
		Short: "Start a new AIR session with the given objective",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("run: not yet implemented (router lands in Phase 2)")
		},
	}
}
