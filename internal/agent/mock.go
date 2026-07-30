package agent

import (
	"context"

	"github.com/levibmackay/air/internal/checkpoint"
)

// Mock is a scriptable Agent implementation used by router and other tests.
// Every behavior defaults to something harmless (available, healthy, never
// rate-limited, never complete) and can be overridden per-field.
type Mock struct {
	NameValue    string
	Installed    bool
	VersionValue string
	Available    bool
	AvailableErr error

	StartFunc      func(ctx context.Context, task Task) (*Session, error)
	ResumeFunc     func(ctx context.Context, cp *checkpoint.Checkpoint) (*Session, error)
	StopFunc       func(session *Session) error
	HealthFunc     func(session *Session) HealthStatus
	RateLimitFunc  func(output string) *RateLimitInfo
	CompletionFunc func(output string) bool
}

// NewMock returns a Mock that is installed, available, and behaves as a
// no-op provider until its fields are overridden.
func NewMock(name string) *Mock {
	return &Mock{NameValue: name, Installed: true, Available: true}
}

func (m *Mock) Name() string { return m.NameValue }

func (m *Mock) DetectInstalled() bool { return m.Installed }

func (m *Mock) DetectVersion() (string, error) { return m.VersionValue, nil }

func (m *Mock) IsAvailable() (bool, error) { return m.Available, m.AvailableErr }

func (m *Mock) Start(ctx context.Context, task Task) (*Session, error) {
	if m.StartFunc != nil {
		return m.StartFunc(ctx, task)
	}
	return NewSession(m.NameValue), nil
}

func (m *Mock) Resume(ctx context.Context, cp *checkpoint.Checkpoint) (*Session, error) {
	if m.ResumeFunc != nil {
		return m.ResumeFunc(ctx, cp)
	}
	return NewSession(m.NameValue), nil
}

func (m *Mock) Stop(session *Session) error {
	if m.StopFunc != nil {
		return m.StopFunc(session)
	}
	return nil
}

func (m *Mock) HealthCheck(session *Session) HealthStatus {
	if m.HealthFunc != nil {
		return m.HealthFunc(session)
	}
	return HealthStatus{Healthy: true}
}

func (m *Mock) DetectRateLimit(output string) *RateLimitInfo {
	if m.RateLimitFunc != nil {
		return m.RateLimitFunc(output)
	}
	return nil
}

func (m *Mock) DetectCompletion(output string) bool {
	if m.CompletionFunc != nil {
		return m.CompletionFunc(output)
	}
	return false
}

var _ Agent = (*Mock)(nil)
