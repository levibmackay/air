// Package gemini implements the agent.Agent interface for the Gemini CLI,
// driven non-interactively via `gemini -p "<prompt>"`.
package gemini

import (
	"github.com/levibmackay/air/internal/agent"
	"github.com/levibmackay/air/internal/providers/cliagent"
)

// New returns a Provider for the Gemini CLI.
func New() agent.Agent {
	return cliagent.NewProvider("gemini", "gemini")
}
