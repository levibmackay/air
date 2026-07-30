package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/levibmackay/air/internal/agent"
	"github.com/levibmackay/air/internal/config"
	"github.com/levibmackay/air/internal/providers"
)

type benchmarkResult struct {
	Provider string
	Status   string
	Duration time.Duration
	Error    error
}

func newBenchmarkCmd() *cobra.Command {
	var targetProviders string
	var timeoutSec int

	cmd := &cobra.Command{
		Use:   "benchmark [objective]",
		Short: "Run the same objective across configured providers and compare results",
		RunE: func(cmd *cobra.Command, args []string) error {
			objective := "Generate a hello world function"
			if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
				objective = args[0]
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			providerList := cfg.Providers
			if targetProviders != "" {
				providerList = strings.Split(targetProviders, ",")
			}

			return runBenchmark(cmd.OutOrStdout(), objective, providerList, time.Duration(timeoutSec)*time.Second)
		},
	}

	cmd.Flags().StringVarP(&targetProviders, "providers", "p", "", "Comma-separated list of providers to benchmark (defaults to configured providers)")
	cmd.Flags().IntVarP(&timeoutSec, "timeout", "t", 60, "Timeout per provider in seconds")

	return cmd
}

func runBenchmark(w io.Writer, objective string, providerEntries []string, timeout time.Duration) error {
	reg := providers.Registry()
	fmt.Fprintf(w, "Benchmarking objective: %q\n", objective)
	fmt.Fprintf(w, "%-20s %-12s %-10s\n", "PROVIDER", "STATUS", "DURATION")
	fmt.Fprintln(w, strings.Repeat("-", 44))

	workDir, _ := os.Getwd()

	for _, entry := range providerEntries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		key, param := agent.ParseProviderKey(entry)
		a, err := reg.Build(key, param)
		if err != nil {
			fmt.Fprintf(w, "%-20s %-12s %-10s\n", entry, "error", "N/A")
			continue
		}

		avail, _ := a.IsAvailable()
		if !avail {
			fmt.Fprintf(w, "%-20s %-12s %-10s\n", a.Name(), "unavailable", "N/A")
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		task := agent.Task{
			Objective: objective,
			WorkDir:   workDir,
		}

		start := time.Now()
		sess, err := a.Start(ctx, task)
		if err != nil {
			cancel()
			fmt.Fprintf(w, "%-20s %-12s %-10s\n", a.Name(), "failed", "N/A")
			continue
		}

		var runErr error
		select {
		case <-sess.Done():
			runErr = sess.Err()
		case <-ctx.Done():
			runErr = ctx.Err()
			_ = a.Stop(sess)
		}
		cancel()

		dur := time.Since(start).Round(10 * time.Millisecond)

		status := "success"
		if runErr != nil {
			status = "failed"
		}

		fmt.Fprintf(w, "%-20s %-12s %-10s\n", a.Name(), status, dur.String())
	}

	return nil
}
