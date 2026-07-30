package router

import (
	"context"
	"testing"
	"time"

	"github.com/levibmackay/air/internal/agent"
	"github.com/levibmackay/air/internal/checkpoint"
)

func TestRouterEmitsEvents(t *testing.T) {
	store := checkpoint.NewStore(t.TempDir())
	p := completesOnStart("solo", "task complete")

	events := make(chan Event, 16)
	r := New([]agent.Agent{p}, store,
		WithPollInterval(5*time.Millisecond),
		WithCheckpointInterval(time.Hour),
		WithEvents(events),
	)

	if _, err := r.Run(context.Background(), "sess-events", "build a REST API"); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	close(events)

	var types []EventType
	for e := range events {
		types = append(types, e.Type)
	}

	if len(types) < 2 {
		t.Fatalf("got %d events, want at least ProviderStarted and Completed: %v", len(types), types)
	}
	if types[0] != EventProviderStarted {
		t.Errorf("first event = %v, want EventProviderStarted", types[0])
	}
	if types[len(types)-1] != EventCompleted {
		t.Errorf("last event = %v, want EventCompleted", types[len(types)-1])
	}
}
