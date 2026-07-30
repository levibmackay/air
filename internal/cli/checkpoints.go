package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/levibmackay/air/internal/checkpoint"
	"github.com/levibmackay/air/internal/config"
)

func newCheckpointsCmd() *cobra.Command {
	var sessionID string

	cmd := &cobra.Command{
		Use:   "checkpoints",
		Short: "List saved checkpoints for the current or a given session",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return err
			}

			if sessionID == "" {
				sessions, err := store.Sessions()
				if err != nil {
					return err
				}
				if len(sessions) == 0 {
					fmt.Println("no checkpoints found")
					return nil
				}
				fmt.Println("sessions:")
				for _, s := range sessions {
					fmt.Printf("  %s\n", s)
				}
				fmt.Println("\npass --session <id> to list its checkpoints")
				return nil
			}

			list, err := store.List(sessionID)
			if err != nil {
				return err
			}
			if len(list) == 0 {
				fmt.Printf("no checkpoints found for session %s\n", sessionID)
				return nil
			}
			for _, cp := range list {
				printCheckpointSummary(cp)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&sessionID, "session", "", "session ID to list checkpoints for")
	return cmd
}

func printCheckpointSummary(cp *checkpoint.Checkpoint) {
	status := "ok"
	if len(cp.Errors) > 0 {
		status = cp.Errors[len(cp.Errors)-1]
	}
	fmt.Printf("  %s  provider=%s  %s\n", cp.ID, cp.Provider, status)
}

// openStore opens AIR's checkpoint store under ~/.air/checkpoints.
func openStore() (*checkpoint.Store, error) {
	home, err := config.AirHome()
	if err != nil {
		return nil, err
	}
	return checkpoint.NewStore(filepath.Join(home, "checkpoints")), nil
}
