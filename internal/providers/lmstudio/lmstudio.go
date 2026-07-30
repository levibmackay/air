// Package lmstudio implements the agent.Agent interface for LM Studio. Not
// yet implemented — registered as a known provider so entries like
// "lmstudio:<model>" resolve to a placeholder that shows up in
// `air providers`/`air doctor`, but the router will never select it.
package lmstudio

import (
	"github.com/levibmackay/air/internal/agent"
	"github.com/levibmackay/air/internal/providers/base"
)

// New returns a not-yet-implemented placeholder for the given LM Studio model.
func New(model string) agent.Agent {
	name := "lmstudio"
	if model != "" {
		name = "lmstudio:" + model
	}
	return base.Unimplemented{NameValue: name, Binary: "lmstudio"}
}
