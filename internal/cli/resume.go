package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/levibmackay/air/internal/checkpoint"
	"github.com/levibmackay/air/internal/router"
	"github.com/levibmackay/air/internal/tui"
)

func newResumeCmd() *cobra.Command {
	var sessionID string
	var useTUI bool

	cmd := &cobra.Command{
		Use:   "resume",
		Short: "Resume an AIR session from its latest checkpoint",
		RunE: func(cmd *cobra.Command, args []string) error {
			if useTUI {
				events := make(chan router.Event, 64)
				r, store, err := buildRouter(events)
				if err != nil {
					return err
				}
				latest, err := resolveLatest(store, sessionID)
				if err != nil {
					return err
				}
				if latest == nil {
					return fmt.Errorf("resume: no checkpoints found")
				}
				cp, err := tui.Resume(cmd.Context(), r, latest, events)
				return runAndReport(cp, err)
			}

			r, store, err := buildRouter(nil)
			if err != nil {
				return err
			}

			latest, err := resolveLatest(store, sessionID)
			if err != nil {
				return err
			}
			if latest == nil {
				return fmt.Errorf("resume: no checkpoints found")
			}

			cp, err := r.Resume(cmd.Context(), latest)
			return runAndReport(cp, err)
		},
	}

	cmd.Flags().StringVar(&sessionID, "session", "", "session ID to resume (defaults to the most recent session)")
	cmd.Flags().BoolVar(&useTUI, "tui", false, "show a live dashboard while resuming")
	return cmd
}

// resolveLatest returns the latest checkpoint for sessionID, or across all
// sessions if sessionID is empty.
func resolveLatest(store *checkpoint.Store, sessionID string) (*checkpoint.Checkpoint, error) {
	if sessionID != "" {
		return store.Latest(sessionID)
	}
	return store.LatestAny()
}
