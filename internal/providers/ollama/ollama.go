// Package ollama implements the agent.Agent interface for locally-hosted
// Ollama models, driven non-interactively via `ollama run <model> "<prompt>"`.
package ollama

import (
	"context"
	"sync"

	"github.com/levibmackay/air/internal/agent"
	"github.com/levibmackay/air/internal/checkpoint"
	"github.com/levibmackay/air/internal/detect"
	"github.com/levibmackay/air/internal/providers/cliagent"
	"github.com/levibmackay/air/internal/summarizer"
)

const defaultModel = "qwen2.5-coder"

// Provider implements agent.Agent for Ollama.
type Provider struct {
	name    string
	model   string
	runner  cliagent.Runner
	mu      sync.Mutex
	cancels map[*agent.Session]context.CancelFunc
}

// New returns an Ollama provider. Model parameter specifies which Ollama model
// to invoke (e.g. "qwen2.5-coder:30b"). If omitted/empty, defaults to "qwen2.5-coder".
func New(model string) agent.Agent {
	m := model
	if m == "" {
		m = defaultModel
	}
	name := "ollama:" + m

	return &Provider{
		name:    name,
		model:   m,
		runner:  cliagent.Runner{Binary: "ollama"},
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
	args := []string{"run", p.model, prompt}

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
