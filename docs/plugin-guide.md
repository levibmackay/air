# Plugin Guide: Adding a Provider

Adding a provider to AIR means implementing one interface and registering
one factory function. The router never branches on a provider's identity,
so nothing outside your new package needs to change.

## 1. Implement `agent.Agent`

```go
type Agent interface {
	Name() string
	DetectInstalled() bool
	DetectVersion() (string, error)
	IsAvailable() (bool, error)
	Start(ctx context.Context, task Task) (*Session, error)
	Resume(ctx context.Context, cp *checkpoint.Checkpoint) (*Session, error)
	Stop(session *Session) error
	HealthCheck(session *Session) HealthStatus
	DetectRateLimit(output string) *RateLimitInfo
	DetectCompletion(output string) bool
}
```

(`internal/agent/agent.go` has the full doc comments for each method.)

### If your provider is a non-interactive CLI you invoke as `<binary> -p "<prompt>"`

You don't need to implement `Agent` from scratch. Use
`internal/providers/cliagent.Provider`, which already handles process
launch, output streaming into a `Session`, exit-code-based completion, and
context-cancellation-based `Stop`:

```go
// internal/providers/yourtool/yourtool.go
package yourtool

import (
	"github.com/levibmackay/air/internal/agent"
	"github.com/levibmackay/air/internal/providers/cliagent"
)

func New() agent.Agent {
	return cliagent.NewProvider("yourtool", "yourtool-cli-binary")
}
```

That's the entire Claude Code and Gemini CLI implementations
(`internal/providers/claude`, `internal/providers/gemini`) — they're
one-line wrappers around `cliagent.Provider`.

### If your provider needs different invocation semantics

Build on `cliagent.Runner` directly (process launch + output streaming,
without the `-p <prompt>` argument shape `cliagent.Provider` assumes), or
implement `Agent` entirely by hand if the provider isn't process-driven at
all (e.g. a provider that talks to a local HTTP daemon like Ollama or LM
Studio would poll an API instead of scanning subprocess output).

### Completion semantics

For a process that exits when its task is done, a clean exit (no error
from `Session.Err()`) is treated by the router as success automatically —
you don't need `DetectCompletion` to return `true` for that. `Session.Err()`
non-nil is always a failure ("unexpected exit"). `DetectCompletion` exists
for providers that can signal completion *while still running* (e.g. a
long-lived interactive session AIR keeps open across tasks); if that
doesn't apply to your provider, just return `false`.

## 2. Register it

Add one line to `internal/providers/registry.go`:

```go
r.Register("yourtool", func(string) (agent.Agent, error) { return yourtool.New(), nil })
```

If your provider takes a parameter from the config entry (like
`ollama:qwen3-coder:30b` → param `"qwen3-coder:30b"`), use it:

```go
r.Register("yourtool", func(param string) (agent.Agent, error) { return yourtool.New(param), nil })
```

## 3. That's it

Users add `yourtool` to their `providers:` list in `air.yaml` /
`~/.air/config.yaml` and it's live — `air doctor`/`air providers` will pick
it up automatically, and the router will try it in whatever order the
config lists it.

## Not-yet-implemented providers

Six providers (`codex`, `opencode`, `aider`, `antigravity`, `ollama`,
`lmstudio`) are registered today as `internal/providers/base.Unimplemented`
placeholders: they report `IsAvailable() == false, nil` so the router
silently skips them, but they still show up in `air doctor`/`air providers`
so users can see AIR knows about them. Replacing one is exactly the process
above — swap the `base.Unimplemented{...}` in that package's `New()` for a
real implementation.
