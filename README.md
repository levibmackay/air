# AIR — AI Router

AIR is a load balancer for AI coding agent CLIs. Run one command, and AIR picks
the best available agent, monitors it, checkpoints progress, and switches to
another agent on failure or rate limit — without losing your work.

```bash
air run "Build me a REST API"
```

```bash
air resume
```

## Status

Early development. `run`, `resume`, `doctor`, `providers`, and `checkpoints`
are wired end-to-end against the Claude Code and Gemini CLIs; the rest of
the providers are registered but not yet implemented (see below). See
[`docs/superpowers/specs/2026-07-30-air-architecture-design.md`](docs/superpowers/specs/2026-07-30-air-architecture-design.md)
for the architecture and phased delivery plan.

## Supported providers

| Provider     | Status          |
| ------------ | --------------- |
| Claude Code  | implemented     |
| Gemini CLI   | implemented     |
| OpenAI Codex | not implemented |
| OpenCode     | not implemented |
| Aider        | not implemented |
| Antigravity  | not implemented |
| Ollama       | not implemented |
| LM Studio    | not implemented |

Every provider is a plugin behind the common `agent.Agent` interface
(`internal/agent/agent.go`); "not implemented" providers register a
placeholder (`internal/providers/base`) so they show up in `air doctor`/
`air providers` without being selectable by the router yet.

## Building

```bash
go build ./...
```

## Commands

`air run [--tui]`, `air resume [--tui]`, `air status`, `air doctor`,
`air providers`, `air checkpoints`, `air rollback`, `air config`,
`air config init`. `air logs`, `air benchmark`, and `air update` are still
stubs.

## Documentation

- [Architecture design](docs/superpowers/specs/2026-07-30-air-architecture-design.md)
- [Developer guide](docs/developer-guide.md) — building, testing, repo layout
- [Plugin guide](docs/plugin-guide.md) — adding a new provider

## Contributing

Adding a provider requires only implementing `agent.Agent` — no router
changes needed. See the [plugin guide](docs/plugin-guide.md) and
`internal/providers/` for existing plugins.

## License

MIT — see [LICENSE](LICENSE).
