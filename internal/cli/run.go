package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run [objective]",
		Short: "Start a new AIR session with the given objective",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			objective := strings.Join(args, " ")

			r, _, err := buildRouter()
			if err != nil {
				return err
			}

			cp, err := r.Run(cmd.Context(), newSessionID(), objective)
			return runAndReport(cp, err)
		},
	}
}
