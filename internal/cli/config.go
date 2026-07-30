package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "View or edit AIR's configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("config: not yet implemented (config package lands in Phase 2)")
		},
	}
}
