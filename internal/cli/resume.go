package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/levibmackay/air/internal/checkpoint"
)

func newResumeCmd() *cobra.Command {
	var sessionID string

	cmd := &cobra.Command{
		Use:   "resume",
		Short: "Resume an AIR session from its latest checkpoint",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, store, err := buildRouter()
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
