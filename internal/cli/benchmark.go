package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newBenchmarkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "benchmark",
		Short: "Run the same objective across configured providers and compare results",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("benchmark: not yet implemented")
		},
	}
}
