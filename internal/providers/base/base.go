// Package base provides Unimplemented, a placeholder agent.Agent for
// providers that are registered but not yet built out. It reports itself as
// unavailable so the router silently skips it rather than erroring, letting
// `air doctor`/`air providers` still list it as a known-but-unimplemented
// provider.
package base

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/levibmackay/air/internal/agent"
	"github.com/levibmackay/air/internal/checkpoint"
)

// Unimplemented is an agent.Agent stand-in for a provider whose real
// implementation hasn't landed yet.
type Unimplemented struct {
	NameValue string
	Binary    string
}

func (u Unimplemented) Name() string { return u.NameValue }

func (u Unimplemented) DetectInstalled() bool {
	if u.Binary == "" {
		return false
	}
	_, err := exec.LookPath(u.Binary)
	return err == nil
}

func (u Unimplemented) DetectVersion() (string, error) {
	return "", fmt.Errorf("%s: not yet implemented", u.NameValue)
}

// IsAvailable always reports false: an unimplemented provider is never a
// candidate for the router to actually run.
func (u Unimplemented) IsAvailable() (bool, error) { return false, nil }

func (u Unimplemented) Start(ctx context.Context, task agent.Task) (*agent.Session, error) {
	return nil, fmt.Errorf("%s: not yet implemented", u.NameValue)
}

func (u Unimplemented) Resume(ctx context.Context, cp *checkpoint.Checkpoint) (*agent.Session, error) {
	return nil, fmt.Errorf("%s: not yet implemented", u.NameValue)
}

func (u Unimplemented) Stop(session *agent.Session) error { return nil }

func (u Unimplemented) HealthCheck(session *agent.Session) agent.HealthStatus {
	return agent.HealthStatus{Healthy: false, Reason: "not yet implemented"}
}

func (u Unimplemented) DetectRateLimit(output string) *agent.RateLimitInfo { return nil }

func (u Unimplemented) DetectCompletion(output string) bool { return false }

var _ agent.Agent = Unimplemented{}
