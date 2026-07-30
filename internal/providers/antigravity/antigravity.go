// Package antigravity implements the agent.Agent interface for Antigravity,
// driven non-interactively via `agy -p "<prompt>"`.
package antigravity

import (
	"github.com/levibmackay/air/internal/agent"
	"github.com/levibmackay/air/internal/providers/cliagent"
)

// New returns a Provider for the Antigravity CLI (`agy`).
func New() agent.Agent {
	return cliagent.NewProvider("antigravity", "agy")
}
