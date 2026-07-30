package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newResumeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resume",
		Short: "Resume the most recent AIR session from its latest checkpoint",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("resume: not yet implemented (checkpoint store lands in Phase 2)")
		},
	}
}
