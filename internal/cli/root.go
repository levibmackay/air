// Package cli defines the AIR command-line interface. Each subcommand is a
// thin wrapper: real logic lives in internal/router, internal/checkpoint, and
// internal/config so it can be tested without going through Cobra.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is set at build time via -ldflags.
var version = "dev"

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "air",
		Short:         "AIR routes your prompt to whichever AI coding agent is available",
		Long:          "AIR (AI Router) launches, monitors, checkpoints, and switches between AI coding agent CLIs so you never have to think about which one is running.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		newRunCmd(),
		newResumeCmd(),
		newStatusCmd(),
		newDoctorCmd(),
		newProvidersCmd(),
		newCheckpointsCmd(),
		newRollbackCmd(),
		newLogsCmd(),
		newConfigCmd(),
		newBenchmarkCmd(),
		newUpdateCmd(),
	)

	return root
}

// Execute runs the AIR CLI and exits the process on error.
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "air:", err)
		os.Exit(1)
	}
}
