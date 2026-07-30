package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update AIR to the latest release",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("update: not yet implemented")
		},
	}
}
