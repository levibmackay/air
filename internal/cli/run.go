package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/levibmackay/air/internal/router"
	"github.com/levibmackay/air/internal/tui"
)

func newRunCmd() *cobra.Command {
	var useTUI bool

	cmd := &cobra.Command{
		Use:   "run [objective]",
		Short: "Start a new AIR session with the given objective",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			objective := strings.Join(args, " ")
			sessionID := newSessionID()

			if useTUI {
				events := make(chan router.Event, 64)
				r, _, err := buildRouter(events)
				if err != nil {
					return err
				}
				cp, err := tui.Run(cmd.Context(), r, sessionID, objective, events)
				return runAndReport(cp, err)
			}

			r, _, err := buildRouter(nil)
			if err != nil {
				return err
			}
			cp, err := r.Run(cmd.Context(), sessionID, objective)
			return runAndReport(cp, err)
		},
	}

	cmd.Flags().BoolVar(&useTUI, "tui", false, "show a live dashboard while the session runs")
	return cmd
}
