package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/levibmackay/air/internal/checkpoint"
)

func newRollbackCmd() *cobra.Command {
	var sessionID string
	var yes bool

	cmd := &cobra.Command{
		Use:   "rollback [checkpoint-id]",
		Short: "Restore the working tree to a prior checkpoint (asks for confirmation)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return err
			}

			var cp *checkpoint.Checkpoint
			switch {
			case len(args) == 1 && sessionID != "":
				cp, err = store.Load(sessionID, args[0])
			case sessionID != "":
				cp, err = store.Latest(sessionID)
			default:
				cp, err = store.LatestAny()
			}
			if err != nil {
				return err
			}
			if cp == nil {
				return fmt.Errorf("rollback: no matching checkpoint found")
			}
			if cp.GitDiff == "" {
				return fmt.Errorf("rollback: checkpoint %s has no recorded git diff to restore", cp.ID)
			}
			if cp.WorkDir == "" {
				return fmt.Errorf("rollback: checkpoint %s has no recorded working directory", cp.ID)
			}

			if !yes && !confirm(fmt.Sprintf("Restore working tree in %s to checkpoint %s (session %s)? This will overwrite uncommitted changes.", cp.WorkDir, cp.ID, cp.Session)) {
				fmt.Println("rollback canceled")
				return nil
			}

			return applyDiff(cp.WorkDir, cp.GitDiff)
		},
	}

	cmd.Flags().StringVar(&sessionID, "session", "", "session to roll back (defaults to the most recently active session)")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip the confirmation prompt")
	return cmd
}

func confirm(prompt string) bool {
	fmt.Printf("%s [y/N] ", prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes"
}

func applyDiff(workDir, diff string) error {
	cmd := exec.Command("git", "-C", workDir, "apply", "--allow-empty")
	cmd.Stdin = strings.NewReader(diff)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rollback: git apply: %w", err)
	}
	fmt.Println("rollback complete")
	return nil
}
