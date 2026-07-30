// Package aider implements the agent.Agent interface for Aider,
// driven non-interactively via `aider --message "<prompt>" --yes-always`.
package aider

import (
	"context"
	"sync"

	"github.com/levibmackay/air/internal/agent"
	"github.com/levibmackay/air/internal/checkpoint"
	"github.com/levibmackay/air/internal/detect"
	"github.com/levibmackay/air/internal/providers/cliagent"
	"github.com/levibmackay/air/internal/summarizer"
)

// Provider implements agent.Agent for Aider CLI.
type Provider struct {
	name    string
	model   string
	runner  cliagent.Runner
	mu      sync.Mutex
	cancels map[*agent.Session]context.CancelFunc
}

// New returns an Aider provider. Optional model parameter can be specified.
func New(model ...string) agent.Agent {
	m := ""
	name := "aider"
	if len(model) > 0 && model[0] != "" {
		m = model[0]
		name = "aider:" + m
	}
	return &Provider{
		name:    name,
		model:   m,
		runner:  cliagent.Runner{Binary: "aider"},
		cancels: make(map[*agent.Session]context.CancelFunc),
	}
}

func (p *Provider) Name() string { return p.name }

func (p *Provider) DetectInstalled() bool { return p.runner.DetectInstalled() }

func (p *Provider) DetectVersion() (string, error) { return p.runner.DetectVersion() }

func (p *Provider) IsAvailable() (bool, error) { return p.runner.DetectInstalled(), nil }

func (p *Provider) Start(ctx context.Context, task agent.Task) (*agent.Session, error) {
	prompt := task.Objective
	if task.ResumePrompt != "" {
		prompt = task.ResumePrompt
	}
	return p.launch(ctx, prompt, task.WorkDir)
}

func (p *Provider) Resume(ctx context.Context, cp *checkpoint.Checkpoint) (*agent.Session, error) {
	return p.launch(ctx, summarizer.Compress(cp), cp.WorkDir)
}

func (p *Provider) launch(ctx context.Context, prompt, dir string) (*agent.Session, error) {
	args := []string{"--message", prompt, "--yes-always"}
	if p.model != "" {
		args = append(args, "--model", p.model)
	}

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

func (p *Provider) HealthCheck(session *agent.Session) agent.HealthStatus {
	return agent.HealthStatus{Healthy: true}
}

func (p *Provider) DetectRateLimit(output string) *agent.RateLimitInfo {
	switch detect.Classify(output) {
	case detect.RateLimit, detect.Unavailable:
		return &agent.RateLimitInfo{Message: p.name + " CLI output matched a rate-limit/availability pattern"}
	default:
		return nil
	}
}

func (p *Provider) DetectCompletion(output string) bool { return false }

var _ agent.Agent = (*Provider)(nil)
