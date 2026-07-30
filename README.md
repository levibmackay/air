# AIR — AI Router

AIR is a load balancer for AI coding agent CLIs. Run one command, and AIR picks
the best available agent, monitors it, checkpoints progress, and switches to
another agent on failure or rate limit — without losing your work.

```bash
air "Build me a REST API"
```

```bash
air resume
```

## Status

Early development. See [`docs/superpowers/specs/2026-07-30-air-architecture-design.md`](docs/superpowers/specs/2026-07-30-air-architecture-design.md)
for the architecture and phased delivery plan.

## Supported providers (planned)

Claude Code, OpenAI Codex CLI, Gemini CLI, OpenCode, Aider, Antigravity,
Ollama, LM Studio — each implemented as a plugin behind a common
`agent.Agent` interface (`internal/agent/agent.go`).

## Building

```bash
go build ./...
```

## Commands

`air run`, `air resume`, `air status`, `air doctor`, `air providers`,
`air checkpoints`, `air rollback`, `air logs`, `air config`, `air benchmark`,
`air update`.

## Contributing

Adding a provider requires only implementing `agent.Agent` — no router
changes needed. See `internal/providers/` for existing plugins.

## License

MIT — see [LICENSE](LICENSE).
