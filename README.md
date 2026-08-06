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
are wired end-to-end against the Claude Code, Gemini CLI, and Antigravity
CLIs; the rest of the providers are registered but not yet implemented (see
below). See
[`docs/superpowers/specs/2026-07-30-air-architecture-design.md`](docs/superpowers/specs/2026-07-30-air-architecture-design.md)
for the architecture and phased delivery plan.

## Supported providers

| Provider     | Status          |
| ------------ | --------------- |
| Claude Code  | implemented     |
| Gemini CLI   | implemented     |
| Antigravity  | implemented, unverified — see below |
| OpenAI Codex | not implemented |
| OpenCode     | not implemented |
| Aider        | not implemented |
| Ollama       | not implemented — `ollama agent` needs a real TTY, so it can't be driven the same way as the others; see `internal/providers/ollama` |
| LM Studio    | not implemented |

**Antigravity note:** it isn't installed on the machine this was built on,
so `internal/providers/antigravity` assumes the same `antigravity -p
"<prompt>"` invocation as Claude Code and Gemini CLI, verified only at the
Go-wiring level (a fake `antigravity` binary standing in for the real one).
If the real CLI's flags differ, that's the one line to fix
(`internal/providers/antigravity/antigravity.go`).

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
`air config init`, `air logs`. `air benchmark` and `air update` are still
stubs.

## Installing

Not yet published. Once `levibmackay/homebrew-tap` and
`levibmackay/scoop-bucket` exist, pushing a `v*` tag
(`.github/workflows/release.yml`) will publish cross-platform binaries via
GoReleaser, a Homebrew cask, a Scoop manifest, and a
`ghcr.io/levibmackay/air` Docker image. Until then, build from source:

```bash
go install github.com/levibmackay/air/cmd/air@latest
```

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

**Last updated:** 2026-08-05 19:46 PDT

