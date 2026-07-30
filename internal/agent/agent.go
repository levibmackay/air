// Package agent defines the plugin contract that every AI coding agent
// provider must implement. The router (internal/router) depends only on this
// interface and never branches on a provider's identity.
package agent

import (
	"context"
	"time"

	"github.com/levibmackay/air/internal/checkpoint"
)

// Task describes the work AIR is asking a provider to perform.
type Task struct {
	Objective string
	// ResumePrompt is set instead of Objective when starting a provider from
	// a compressed checkpoint rather than a fresh objective.
	ResumePrompt string
	WorkDir      string
}

// Session is a handle to a running provider process.
type Session struct {
	Provider  string
	StartedAt time.Time
	PID       int
}

// HealthStatus reports whether a running session is progressing normally.
type HealthStatus struct {
	Healthy bool
	Reason  string
}

// RateLimitInfo describes a detected rate limit, quota, or availability
// failure so the router can decide whether and when to retry this provider.
type RateLimitInfo struct {
	Message    string
	RetryAfter time.Duration // zero if unknown
}

// Agent is the contract every provider plugin implements. Providers own all
// CLI-specific behavior: how to invoke their binary, how to parse their
// stdout, and how to recognize their own rate-limit and completion signals.
type Agent interface {
	// Name is the provider's stable identifier, as used in air.yaml.
	Name() string

	// DetectInstalled reports whether the provider's CLI is present on PATH.
	DetectInstalled() bool

	// DetectVersion returns the installed CLI's version string.
	DetectVersion() (string, error)

	// IsAvailable reports whether the provider is installed and currently
	// usable (not mid-cooldown from a prior rate limit, etc.).
	IsAvailable() (bool, error)

	// Start launches the provider on a fresh task.
	Start(ctx context.Context, task Task) (*Session, error)

	// Resume launches the provider continuing from a checkpoint.
	Resume(ctx context.Context, cp *checkpoint.Checkpoint) (*Session, error)

	// Stop terminates a running session.
	Stop(session *Session) error

	// HealthCheck reports whether a session is still progressing.
	HealthCheck(session *Session) HealthStatus

	// DetectRateLimit inspects a chunk of provider output and returns
	// non-nil if it recognizes a rate limit, quota, or availability failure.
	DetectRateLimit(output string) *RateLimitInfo

	// DetectCompletion inspects a chunk of provider output and reports
	// whether the provider considers its task finished.
	DetectCompletion(output string) bool
}
