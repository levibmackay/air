package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newLogsCmd() *cobra.Command {
	var sessionID string
	var follow int

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Show terminal output captured in a session's checkpoints",
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
				fmt.Println("no checkpoints found")
				return nil
			}

			if follow > 0 && follow < len(list) {
				list = list[len(list)-follow:]
			}

			for _, cp := range list {
				fmt.Printf("=== %s  %s  (session %s) ===\n", cp.ID, cp.Provider, cp.Session)
				if cp.TerminalOutput == "" {
					fmt.Println("(no output captured)")
				} else {
					fmt.Print(cp.TerminalOutput)
				}
				if len(cp.Errors) > 0 {
					fmt.Printf("--- errors: %v\n", cp.Errors)
				}
				fmt.Println()
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&sessionID, "session", "", "session ID to show logs for (defaults to the most recent session)")
	cmd.Flags().IntVarP(&follow, "last", "n", 0, "only show the last N checkpoints' output (0 = all)")
	return cmd
}
