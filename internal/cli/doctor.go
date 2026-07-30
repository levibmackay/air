package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose AIR's config and check which providers are installed and available",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("doctor: not yet implemented (config validation lands in Phase 2, provider detection in Phase 3)")
		},
	}
}
