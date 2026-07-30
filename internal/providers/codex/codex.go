// Package codex implements the agent.Agent interface for the OpenAI Codex
// CLI. Not yet implemented — registered as a known provider so it shows up
// in `air providers`/`air doctor`, but the router will never select it.
package codex

import (
	"github.com/levibmackay/air/internal/agent"
	"github.com/levibmackay/air/internal/providers/base"
)

// New returns a not-yet-implemented placeholder for the Codex CLI.
func New() agent.Agent {
	return base.Unimplemented{NameValue: "codex", Binary: "codex"}
}
