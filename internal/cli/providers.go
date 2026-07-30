package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newProvidersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "providers",
		Short: "List configured providers and their availability",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("providers: not yet implemented (registry lands in Phase 3)")
		},
	}
}
