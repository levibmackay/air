// Package antigravity implements the agent.Agent interface for Antigravity,
// driven non-interactively via `antigravity -p "<prompt>"` — the same
// invocation shape as Claude Code and Gemini CLI. This hasn't been verified
// against a real Antigravity install (not present on the machine this was
// built on); if the actual CLI uses different flags, update the binary name
// or invocation here.
package antigravity

import (
	"github.com/levibmackay/air/internal/agent"
	"github.com/levibmackay/air/internal/providers/cliagent"
)

// New returns a Provider for the Antigravity CLI.
func New() agent.Agent {
	return cliagent.NewProvider("antigravity", "antigravity")
}
