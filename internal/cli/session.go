package cli

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/levibmackay/air/internal/agent"
	"github.com/levibmackay/air/internal/checkpoint"
	"github.com/levibmackay/air/internal/config"
	"github.com/levibmackay/air/internal/providers"
	"github.com/levibmackay/air/internal/router"
)

// newSessionID mints a sortable, human-readable session identifier.
func newSessionID() string {
	return "sess-" + time.Now().UTC().Format("20060102T150405")
}

// buildRouter loads config, resolves its provider list through the built-in
// registry, and returns a Router plus the checkpoint store it's using —
// the wiring shared by `air run`, `air resume`, `air doctor`, and
// `air providers`. Pass a non-nil events channel to drive a live TUI (see
// internal/tui); doing so also suppresses the structured logger, since its
// JSON output would otherwise corrupt the dashboard's terminal rendering.
func buildRouter(events chan router.Event) (*router.Router, *checkpoint.Store, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("load config: %w", err)
	}

	agents, err := agent.BuildProviders(cfg.Providers, providers.Registry())
	if err != nil {
		return nil, nil, fmt.Errorf("resolve providers: %w", err)
	}

	store, err := openStore()
	if err != nil {
		return nil, nil, err
	}

	opts := []router.Option{router.WithCheckpointInterval(cfg.CheckpointInterval)}
	if events != nil {
		opts = append(opts, router.WithEvents(events), router.WithLogger(zap.NewNop()))
	} else if logger, err := zap.NewProduction(); err == nil {
		opts = append(opts, router.WithLogger(logger))
	}

	r := router.New(agents, store, opts...)
	return r, store, nil
}

// runAndReport drives a Router run/resume to completion and prints a
// one-line result to stdout, returning a non-nil error if the run didn't
// complete successfully.
func runAndReport(cp *checkpoint.Checkpoint, runErr error) error {
	if cp != nil {
		fmt.Printf("session %s: last active provider %s\n", cp.Session, cp.Provider)
	}
	if runErr != nil {
		return runErr
	}
	fmt.Println("done.")
	return nil
}
