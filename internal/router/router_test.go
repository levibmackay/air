package router

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/levibmackay/air/internal/agent"
	"github.com/levibmackay/air/internal/checkpoint"
)

func newTestRouter(t *testing.T, providers []agent.Agent) *Router {
	t.Helper()
	store := checkpoint.NewStore(t.TempDir())
	return New(providers, store,
		WithPollInterval(5*time.Millisecond),
		WithCheckpointInterval(time.Hour), // keep periodic saves out of the way of assertions
	)
}

func completesOnStart(name, marker string) *agent.Mock {
	m := agent.NewMock(name)
	m.CompletionFunc = func(output string) bool { return strings.Contains(output, marker) }
	m.StartFunc = func(ctx context.Context, task agent.Task) (*agent.Session, error) {
		sess := agent.NewSession(name)
		go func() {
			sess.AppendOutput(marker)
			sess.MarkDone(nil)
		}()
		return sess, nil
	}
	return m
}

func TestRouterCompletesWithFirstProvider(t *testing.T) {
	p := completesOnStart("solo", "task complete")
	r := newTestRouter(t, []agent.Agent{p})

	cp, err := r.Run(context.Background(), "sess-1", "build a REST API")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if cp.Provider != "solo" {
		t.Errorf("cp.Provider = %q, want %q", cp.Provider, "solo")
	}
	if len(cp.Errors) != 0 {
		t.Errorf("cp.Errors = %v, want empty on success", cp.Errors)
	}
}

func TestRouterSwitchesOnRateLimit(t *testing.T) {
	p1 := agent.NewMock("limited")
	p1.RateLimitFunc = func(output string) *agent.RateLimitInfo {
		return &agent.RateLimitInfo{Message: "quota exceeded"}
	}
	p1.StartFunc = func(ctx context.Context, task agent.Task) (*agent.Session, error) {
		return agent.NewSession("limited"), nil // never marks done; router must detect via polling
	}

	p2 := agent.NewMock("backup")
	p2.CompletionFunc = func(output string) bool { return strings.Contains(output, "done") }
	p2.ResumeFunc = func(ctx context.Context, cp *checkpoint.Checkpoint) (*agent.Session, error) {
		if cp == nil || cp.Provider != "limited" {
			t.Errorf("Resume() checkpoint = %+v, want one carried from provider %q", cp, "limited")
		}
		sess := agent.NewSession("backup")
		go func() {
			sess.AppendOutput("done")
			sess.MarkDone(nil)
		}()
		return sess, nil
	}

	r := newTestRouter(t, []agent.Agent{p1, p2})
	cp, err := r.Run(context.Background(), "sess-2", "build a REST API")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if cp.Provider != "backup" {
		t.Errorf("cp.Provider = %q, want %q", cp.Provider, "backup")
	}
}

func TestRouterSkipsUnavailableProvider(t *testing.T) {
	p1 := agent.NewMock("down")
	p1.Available = false

	p2 := completesOnStart("up", "done")

	r := newTestRouter(t, []agent.Agent{p1, p2})
	cp, err := r.Run(context.Background(), "sess-3", "build a REST API")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if cp.Provider != "up" {
		t.Errorf("cp.Provider = %q, want %q", cp.Provider, "up")
	}
}

func TestRouterAllProvidersExhausted(t *testing.T) {
	fails := func(name string) *agent.Mock {
		m := agent.NewMock(name)
		crash := func(sess *agent.Session) {
			go func() {
				sess.MarkDone(errors.New("crashed"))
			}()
		}
		m.StartFunc = func(ctx context.Context, task agent.Task) (*agent.Session, error) {
			sess := agent.NewSession(name)
			crash(sess)
			return sess, nil
		}
		m.ResumeFunc = func(ctx context.Context, cp *checkpoint.Checkpoint) (*agent.Session, error) {
			sess := agent.NewSession(name)
			crash(sess)
			return sess, nil
		}
		return m
	}

	p1, p2 := fails("p1"), fails("p2")
	r := newTestRouter(t, []agent.Agent{p1, p2})

	cp, err := r.Run(context.Background(), "sess-4", "build a REST API")
	if err == nil {
		t.Fatal("Run() error = nil, want an error since every provider crashed")
	}
	if cp == nil || cp.Provider != "p2" {
		t.Errorf("cp = %+v, want last-attempted provider p2's checkpoint", cp)
	}
	if len(cp.Errors) == 0 || !strings.Contains(cp.Errors[0], "unexpected exit") {
		t.Errorf("cp.Errors = %v, want an unexpected exit entry", cp.Errors)
	}
}

func TestRouterNoProvidersConfigured(t *testing.T) {
	r := newTestRouter(t, nil)
	_, err := r.Run(context.Background(), "sess-5", "build a REST API")
	if err == nil {
		t.Fatal("Run() with no providers should error")
	}
}

func TestRouterContextCancellation(t *testing.T) {
	p := agent.NewMock("slow")
	p.StartFunc = func(ctx context.Context, task agent.Task) (*agent.Session, error) {
		return agent.NewSession("slow"), nil // never completes on its own
	}
	r := newTestRouter(t, []agent.Agent{p})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := r.Run(ctx, "sess-6", "build a REST API")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Run() error = %v, want context.DeadlineExceeded", err)
	}
}
