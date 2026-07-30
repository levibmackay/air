// Package providers wires every built-in provider package into a single
// agent.Registry. Adding a new provider means adding one Register call here
// and a package implementing agent.Agent — no router or CLI changes needed.
package providers

import (
	"github.com/levibmackay/air/internal/agent"
	"github.com/levibmackay/air/internal/providers/aider"
	"github.com/levibmackay/air/internal/providers/antigravity"
	"github.com/levibmackay/air/internal/providers/claude"
	"github.com/levibmackay/air/internal/providers/codex"
	"github.com/levibmackay/air/internal/providers/gemini"
	"github.com/levibmackay/air/internal/providers/lmstudio"
	"github.com/levibmackay/air/internal/providers/ollama"
	"github.com/levibmackay/air/internal/providers/opencode"
)

// Registry returns an agent.Registry populated with every built-in
// provider. Config keys with a ":<param>" suffix (e.g. "ollama:qwen3-coder:30b")
// are handled by passing the param through to providers that accept one;
// providers that don't ignore it.
func Registry() *agent.Registry {
	r := agent.NewRegistry()

	r.Register("claude", func(string) (agent.Agent, error) { return claude.New(), nil })
	r.Register("codex", func(string) (agent.Agent, error) { return codex.New(), nil })
	r.Register("gemini", func(string) (agent.Agent, error) { return gemini.New(), nil })
	r.Register("opencode", func(string) (agent.Agent, error) { return opencode.New(), nil })
	r.Register("aider", func(string) (agent.Agent, error) { return aider.New(), nil })
	r.Register("antigravity", func(string) (agent.Agent, error) { return antigravity.New(), nil })
	r.Register("ollama", func(param string) (agent.Agent, error) { return ollama.New(param), nil })
	r.Register("lmstudio", func(param string) (agent.Agent, error) { return lmstudio.New(param), nil })

	return r
}
