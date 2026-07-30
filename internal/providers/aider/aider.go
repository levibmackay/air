// Package aider implements the agent.Agent interface for Aider. Not yet
// implemented — registered as a known provider so it shows up in
// `air providers`/`air doctor`, but the router will never select it.
package aider

import (
	"github.com/levibmackay/air/internal/agent"
	"github.com/levibmackay/air/internal/providers/base"
)

// New returns a not-yet-implemented placeholder for Aider.
func New() agent.Agent {
	return base.Unimplemented{NameValue: "aider", Binary: "aider"}
}
