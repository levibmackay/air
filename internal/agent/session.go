package agent

import (
	"strings"
	"sync"
	"time"
)

// Session is a handle to a running provider process. Providers create a
// Session in Start/Resume and append output to it as they read their
// subprocess's stdout/stderr; the router polls Output and Done to drive its
// monitor loop without knowing how any given provider works internally.
type Session struct {
	Provider  string
	StartedAt time.Time
	PID       int

	mu     sync.Mutex
	output strings.Builder
	done   chan struct{}
	err    error
	closed bool
}

// NewSession creates a running Session for the given provider name.
func NewSession(provider string) *Session {
	return &Session{
		Provider:  provider,
		StartedAt: time.Now(),
		done:      make(chan struct{}),
	}
}

// AppendOutput appends a chunk of the provider's stdout/stderr to the
// session's accumulated output. Safe to call concurrently with Output.
func (s *Session) AppendOutput(chunk string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.output.WriteString(chunk)
}

// Output returns everything appended so far.
func (s *Session) Output() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.output.String()
}

// MarkDone signals that the provider's process has exited, with a non-nil
// err if it exited abnormally. MarkDone is idempotent; only the first call
// has an effect.
func (s *Session) MarkDone(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	s.err = err
	close(s.done)
}

// Done returns a channel that closes once the session has exited.
func (s *Session) Done() <-chan struct{} {
	return s.done
}

// Err returns the exit error recorded by MarkDone, if any. Only meaningful
// after Done has closed.
func (s *Session) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}
