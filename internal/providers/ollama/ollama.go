// Package ollama implements the agent.Agent interface for locally-hosted
// Ollama models. Not yet implemented — registered as a known provider so
// entries like "ollama:qwen3-coder:30b" resolve to a placeholder that shows
// up in `air providers`/`air doctor`, but the router will never select it.
//
// `ollama agent` (its agentic, tool-using mode — the closest match to
// Claude Code/Gemini CLI) requires a real TTY and errors out under
// cliagent.Runner's piped stdout/stderr, so it can't be wired up the same
// way as those. Making this real needs either a pty (e.g.
// github.com/creack/pty) inside Runner, or falling back to plain
// `ollama run <model> <prompt>`, which works headless but has no tool use.
package ollama

import (
	"github.com/levibmackay/air/internal/agent"
	"github.com/levibmackay/air/internal/providers/base"
)

// New returns a not-yet-implemented placeholder for the given Ollama model.
func New(model string) agent.Agent {
	name := "ollama"
	if model != "" {
		name = "ollama:" + model
	}
	return base.Unimplemented{NameValue: name, Binary: "ollama"}
}
