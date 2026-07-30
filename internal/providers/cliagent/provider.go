package cliagent

import (
	"context"
	"sync"

	"github.com/levibmackay/air/internal/agent"
	"github.com/levibmackay/air/internal/checkpoint"
	"github.com/levibmackay/air/internal/detect"
	"github.com/levibmackay/air/internal/summarizer"
)

// Provider is a ready-made agent.Agent for any CLI that accepts a single
// non-interactive prompt via `<binary> -p "<prompt>"` and exits when its
// task is done. It's the common shape shared by Claude Code and Gemini CLI
// today; a provider that needs a different invocation can build its own
// Agent directly on top of Runner instead of using this type.
type Provider struct {
	NameValue string
	runner    Runner

	mu      sync.Mutex
	cancels map[*agent.Session]context.CancelFunc
}

// NewProvider returns a Provider that invokes binary as name's CLI.
func NewProvider(name, binary string) *Provider {
	return &Provider{
		NameValue: name,
		runner:    Runner{Binary: binary},
		cancels:   make(map[*agent.Session]context.CancelFunc),
	}
}

func (p *Provider) Name() string { return p.NameValue }

func (p *Provider) DetectInstalled() bool { return p.runner.DetectInstalled() }

func (p *Provider) DetectVersion() (string, error) { return p.runner.DetectVersion() }

func (p *Provider) IsAvailable() (bool, error) { return p.runner.DetectInstalled(), nil }

func (p *Provider) Start(ctx context.Context, task agent.Task) (*agent.Session, error) {
	prompt := task.Objective
	if task.ResumePrompt != "" {
		prompt = task.ResumePrompt
	}
	return p.launch(ctx, []string{"-p", prompt}, task.WorkDir)
}

func (p *Provider) Resume(ctx context.Context, cp *checkpoint.Checkpoint) (*agent.Session, error) {
	return p.launch(ctx, []string{"-p", summarizer.Compress(cp)}, cp.WorkDir)
}

func (p *Provider) launch(ctx context.Context, args []string, dir string) (*agent.Session, error) {
	runCtx, cancel := context.WithCancel(ctx)
	sess, err := p.runner.Launch(runCtx, args, dir)
	if err != nil {
		cancel()
		return nil, err
	}
	p.mu.Lock()
	p.cancels[sess] = cancel
	p.mu.Unlock()
	return sess, nil
}

// Stop cancels the context backing session's subprocess, terminating it.
func (p *Provider) Stop(session *agent.Session) error {
	p.mu.Lock()
	cancel, ok := p.cancels[session]
	delete(p.cancels, session)
	p.mu.Unlock()
	if ok {
		cancel()
	}
	return nil
}

// HealthCheck always reports healthy: a subprocess-driven CLI's health is
// fully captured by Session.Done()/Err(), which the router already checks.
func (p *Provider) HealthCheck(session *agent.Session) agent.HealthStatus {
	return agent.HealthStatus{Healthy: true}
}

// DetectRateLimit scans output for AIR's shared rate-limit/availability
// patterns (see internal/detect).
func (p *Provider) DetectRateLimit(output string) *agent.RateLimitInfo {
	switch detect.Classify(output) {
	case detect.RateLimit, detect.Unavailable:
		return &agent.RateLimitInfo{Message: p.NameValue + " CLI output matched a rate-limit/availability pattern"}
	default:
		return nil
	}
}

// DetectCompletion always reports false: this provider signals completion
// solely by exiting, which the router treats as success independently of
// DetectCompletion (see router.monitor).
func (p *Provider) DetectCompletion(output string) bool { return false }

var _ agent.Agent = (*Provider)(nil)
