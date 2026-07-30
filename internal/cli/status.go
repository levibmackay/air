package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/levibmackay/air/internal/checkpoint"
	"github.com/levibmackay/air/internal/cost"
)

func newStatusCmd() *cobra.Command {
	var sessionID string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the current session's provider, elapsed time, cost, and tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return err
			}

			list, err := loadSessionCheckpoints(store, sessionID)
			if err != nil {
				return err
			}
			if len(list) == 0 {
				fmt.Println("no active or past session found")
				return nil
			}

			latest := list[len(list)-1]
			ledger := cost.FromCheckpoints(list)

			fmt.Printf("session:     %s\n", latest.Session)
			fmt.Printf("provider:    %s\n", latest.Provider)
			fmt.Printf("objective:   %s\n", latest.Objective)
			fmt.Printf("checkpoints: %d\n", len(list))
			fmt.Printf("elapsed:     %s\n", ledger.TotalDuration())

			inTok, outTok := ledger.TotalTokens()
			totalCost := ledger.TotalCostUSD()
			if inTok == 0 && outTok == 0 && totalCost == 0 {
				fmt.Println("tokens/cost: not reported by this provider")
			} else {
				fmt.Printf("tokens:      %d in / %d out\n", inTok, outTok)
				fmt.Printf("cost:        $%.4f\n", totalCost)
			}

			if len(latest.Errors) > 0 {
				fmt.Printf("last error:  %s\n", latest.Errors[len(latest.Errors)-1])
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&sessionID, "session", "", "session ID to report on (defaults to the most recent session)")
	return cmd
}

// loadSessionCheckpoints returns every checkpoint for sessionID, or for the
// most recently active session if sessionID is empty.
func loadSessionCheckpoints(store *checkpoint.Store, sessionID string) ([]*checkpoint.Checkpoint, error) {
	if sessionID == "" {
		latest, err := store.LatestAny()
		if err != nil {
			return nil, err
		}
		if latest == nil {
			return nil, nil
		}
		sessionID = latest.Session
	}
	return store.List(sessionID)
}
