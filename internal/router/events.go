package router

import "github.com/levibmackay/air/internal/checkpoint"

// EventType classifies a router Event.
type EventType int

const (
	// EventProviderStarted fires when a provider is launched or resumed.
	EventProviderStarted EventType = iota
	// EventCheckpoint fires whenever a checkpoint is saved, periodic or not.
	EventCheckpoint
	// EventProviderFailed fires when a provider's run ends in failure and
	// the router is about to try the next one (or give up).
	EventProviderFailed
	// EventCompleted fires once, when a run finishes successfully.
	EventCompleted
)

// Event is a progress notification emitted by a running Router, primarily
// for internal/tui to render live. Consuming events is optional: a Router
// with no configured channel runs exactly as before.
type Event struct {
	Type       EventType
	Provider   string
	Checkpoint *checkpoint.Checkpoint
}

// WithEvents makes the Router send an Event to ch at each significant step
// of a run. Sends are non-blocking: if ch is full, the event is dropped
// rather than stalling the run, so a slow or absent consumer never affects
// orchestration.
func WithEvents(ch chan<- Event) Option {
	return func(r *Router) { r.events = ch }
}

func (r *Router) emit(e Event) {
	if r.events == nil {
		return
	}
	select {
	case r.events <- e:
	default:
	}
}
