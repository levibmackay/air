// Package claude implements the agent.Agent interface for the Claude Code
// CLI, driven non-interactively via `claude -p "<prompt>"`.
package claude

import (
	"github.com/levibmackay/air/internal/agent"
	"github.com/levibmackay/air/internal/providers/cliagent"
)

// New returns a Provider for the Claude Code CLI.
func New() agent.Agent {
	return cliagent.NewProvider("claude", "claude")
}
