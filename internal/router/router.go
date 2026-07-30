// Package router is AIR's core orchestration loop: it launches providers,
// monitors their output, detects failures and completion, saves
// checkpoints, and switches providers when one becomes unavailable. Router
// depends only on internal/agent's interface, so it never branches on a
// specific provider's identity.
package router

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"go.uber.org/zap"

	"github.com/levibmackay/air/internal/agent"
	"github.com/levibmackay/air/internal/checkpoint"
)

const (
	defaultCheckpointInterval = 2 * time.Minute
	defaultPollInterval       = 2 * time.Second
)

// outcome is the result of monitoring a single provider session.
type outcome int

const (
	outcomeCompleted outcome = iota
	outcomeFailed
)

// Router launches providers in order, monitors them, and switches to the
// next available provider on failure, carrying state forward via
// checkpoints.
type Router struct {
	providers []agent.Agent
	store     *checkpoint.Store

	checkpointInterval time.Duration
	pollInterval       time.Duration
	logger             *zap.Logger
	events             chan<- Event
}

// Option configures a Router.
type Option func(*Router)

// WithCheckpointInterval overrides how often the router saves a checkpoint
// while a provider is running, independent of failure detection.
func WithCheckpointInterval(d time.Duration) Option {
	return func(r *Router) { r.checkpointInterval = d }
}

// WithPollInterval overrides how often the router polls a running session's
// output for completion, rate limits, and health.
func WithPollInterval(d time.Duration) Option {
	return func(r *Router) { r.pollInterval = d }
}

// WithLogger overrides the router's logger. Defaults to zap.NewNop().
func WithLogger(l *zap.Logger) Option {
	return func(r *Router) { r.logger = l }
}

// New builds a Router that tries providers in the given order.
func New(providers []agent.Agent, store *checkpoint.Store, opts ...Option) *Router {
	r := &Router{
		providers:          providers,
		store:              store,
		checkpointInterval: defaultCheckpointInterval,
		pollInterval:       defaultPollInterval,
		logger:             zap.NewNop(),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Run drives objective to completion, trying each configured provider in
// order and switching on failure. It returns the final checkpoint (whether
// the run succeeded or every provider was exhausted) alongside an error
// describing why the run didn't complete, if it didn't.
func (r *Router) Run(ctx context.Context, sessionID, objective string) (*checkpoint.Checkpoint, error) {
	return r.run(ctx, sessionID, objective, nil)
}

// Resume continues a previously-checkpointed session: the first available
// provider is asked to Resume from cp rather than Start fresh, and the
// router falls through to later providers exactly as Run does if that
// provider fails too.
func (r *Router) Resume(ctx context.Context, cp *checkpoint.Checkpoint) (*checkpoint.Checkpoint, error) {
	if cp == nil {
		return nil, errors.New("router: cannot resume a nil checkpoint")
	}
	return r.run(ctx, cp.Session, cp.Objective, cp)
}

// run is the shared implementation behind Run and Resume. startCP is nil
// for a fresh Run, or the checkpoint to Resume from for the first provider
// tried; either way, once a provider fails, later providers always resume
// from whatever checkpoint that failure produced.
func (r *Router) run(ctx context.Context, sessionID, objective string, startCP *checkpoint.Checkpoint) (*checkpoint.Checkpoint, error) {
	if len(r.providers) == 0 {
		return nil, errors.New("router: no providers configured")
	}

	task := agent.Task{Objective: objective}
	if startCP != nil {
		task.WorkDir = startCP.WorkDir
	}
	lastCP := startCP
	var lastErr error

	for _, p := range r.providers {
		available, err := p.IsAvailable()
		if err != nil {
			r.logger.Warn("provider availability check failed", zap.String("provider", p.Name()), zap.Error(err))
			lastErr = err
			continue
		}
		if !available {
			r.logger.Info("provider unavailable, skipping", zap.String("provider", p.Name()))
			continue
		}

		var sess *agent.Session
		if lastCP != nil {
			sess, err = p.Resume(ctx, lastCP)
		} else {
			sess, err = p.Start(ctx, task)
		}
		if err != nil {
			r.logger.Warn("provider failed to start", zap.String("provider", p.Name()), zap.Error(err))
			lastErr = err
			continue
		}
		r.emit(Event{Type: EventProviderStarted, Provider: p.Name()})

		cp, out, err := r.monitor(ctx, p, sess, sessionID, objective, task.WorkDir)
		if err != nil {
			return lastCP, err
		}

		switch out {
		case outcomeCompleted:
			r.emit(Event{Type: EventCompleted, Provider: p.Name(), Checkpoint: cp})
			return cp, nil
		case outcomeFailed:
			lastCP = cp
			r.emit(Event{Type: EventProviderFailed, Provider: p.Name(), Checkpoint: cp})
			if n := len(cp.Errors); n > 0 {
				lastErr = fmt.Errorf("provider %s failed: %s", p.Name(), cp.Errors[n-1])
			} else {
				lastErr = fmt.Errorf("provider %s failed", p.Name())
			}
		}
	}

	if lastErr == nil {
		lastErr = errors.New("no configured provider is available")
	}
	return lastCP, fmt.Errorf("router: exhausted all providers: %w", lastErr)
}

// monitor watches a single running session until it completes, fails, or
// the context is canceled, saving checkpoints along the way.
func (r *Router) monitor(ctx context.Context, p agent.Agent, sess *agent.Session, sessionID, objective, workDir string) (*checkpoint.Checkpoint, outcome, error) {
	pollTicker := time.NewTicker(r.pollInterval)
	defer pollTicker.Stop()
	checkpointTicker := time.NewTicker(r.checkpointInterval)
	defer checkpointTicker.Stop()

	finish := func(reason string, out outcome) (*checkpoint.Checkpoint, outcome, error) {
		cp := r.buildCheckpoint(sessionID, p.Name(), objective, sess.Output(), workDir, reason)
		if err := r.store.Save(cp); err != nil {
			r.logger.Warn("failed to save checkpoint", zap.Error(err))
		}
		r.emit(Event{Type: EventCheckpoint, Provider: p.Name(), Checkpoint: cp})
		return cp, out, nil
	}

	for {
		select {
		case <-ctx.Done():
			_ = p.Stop(sess)
			cp := r.buildCheckpoint(sessionID, p.Name(), objective, sess.Output(), workDir, "canceled: "+ctx.Err().Error())
			_ = r.store.Save(cp)
			return cp, outcomeFailed, ctx.Err()

		case <-sess.Done():
			// A clean process exit (err == nil) is the authoritative success
			// signal for process-driven providers: DetectCompletion exists to
			// catch an explicit signal *while* a provider is still running
			// (see the poll case below), not to second-guess a clean exit.
			// The one thing we still check is whether the final output
			// carries a failure signal that didn't produce a non-zero exit
			// code.
			output := sess.Output()
			if err := sess.Err(); err != nil {
				return finish("unexpected exit: "+err.Error(), outcomeFailed)
			}
			if rl := p.DetectRateLimit(output); rl != nil {
				return finish("rate limit: "+rl.Message, outcomeFailed)
			}
			return finish("", outcomeCompleted)

		case <-checkpointTicker.C:
			cp := r.buildCheckpoint(sessionID, p.Name(), objective, sess.Output(), workDir, "")
			if err := r.store.Save(cp); err != nil {
				r.logger.Warn("failed to save periodic checkpoint", zap.Error(err))
			}
			r.emit(Event{Type: EventCheckpoint, Provider: p.Name(), Checkpoint: cp})

		case <-pollTicker.C:
			output := sess.Output()

			if rl := p.DetectRateLimit(output); rl != nil {
				_ = p.Stop(sess)
				return finish("rate limit: "+rl.Message, outcomeFailed)
			}
			if p.DetectCompletion(output) {
				_ = p.Stop(sess)
				return finish("", outcomeCompleted)
			}
			if health := p.HealthCheck(sess); !health.Healthy {
				_ = p.Stop(sess)
				return finish("unhealthy: "+health.Reason, outcomeFailed)
			}
		}
	}
}

// buildCheckpoint snapshots the current state of a run. If reason is
// non-empty it's appended to Errors; an empty reason marks a clean
// checkpoint (periodic save or successful completion).
func (r *Router) buildCheckpoint(sessionID, provider, objective, output, workDir, reason string) *checkpoint.Checkpoint {
	cp := &checkpoint.Checkpoint{
		Session:        sessionID,
		Provider:       provider,
		Created:        time.Now(),
		Objective:      objective,
		WorkDir:        workDir,
		TerminalOutput: output,
		GitDiff:        captureGitDiff(workDir),
	}
	if reason != "" {
		cp.Errors = []string{reason}
	}
	return cp
}

// captureGitDiff returns `git diff` output for workDir, or "" if workDir is
// unset or isn't a git repository. Best-effort: errors are swallowed since a
// missing diff shouldn't block checkpointing.
func captureGitDiff(workDir string) string {
	if workDir == "" {
		return ""
	}
	cmd := exec.Command("git", "-C", workDir, "diff")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}
