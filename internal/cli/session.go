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
// `air providers`.
func buildRouter() (*router.Router, *checkpoint.Store, error) {
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

	logger, err := zap.NewProduction()
	if err != nil {
		logger = zap.NewNop()
	}

	r := router.New(agents, store, router.WithCheckpointInterval(cfg.CheckpointInterval), router.WithLogger(logger))
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
