# AIR (AI Router) — Architecture Design

Date: 2026-07-30
Status: Approved

## Vision

AIR is a load balancer / orchestrator for AI coding agent CLIs (Claude Code, Codex CLI,
Gemini CLI, OpenCode, Aider, Antigravity, Ollama, LM Studio, ...). The user runs
`air "Build me a REST API"` and AIR picks an available provider, monitors it, checkpoints
progress, and transparently switches providers on failure/rate-limit/quota without losing
progress.

## Module

`github.com/levibmackay/air` — single Go binary, Cobra CLI, Bubble Tea TUI, Viper config,
Zap logging.

## Repository layout

```
air/
  cmd/air/            main.go — thin entrypoint, calls cli.Execute()
  internal/
    cli/              Cobra commands (run, resume, status, doctor, providers, checkpoints,
                       rollback, logs, config, benchmark, update)
    router/           core router: event loop, provider queue, retry/switch logic
    agent/            Agent interface + registry (plugin contract) + mock provider for tests
    providers/
      claude/         Claude Code plugin
      codex/          OpenAI Codex CLI plugin
      gemini/         Gemini CLI plugin
      opencode/
      aider/
      antigravity/
      ollama/
      lmstudio/
    checkpoint/       checkpoint struct, JSON store, save/load/rollback
    detect/           shared heuristics for rate-limit, quota, context-overflow, crash
                       detection (providers compose these; router stays provider-agnostic)
    summarizer/       context-compression: builds the resume prompt from a checkpoint
    config/           Viper-backed YAML config loading + validation
    tui/              Bubble Tea dashboard (model/update/view, components)
    cost/             token/cost tracking, usage ledger
  pkg/                reserved, empty until a public API is intentional
  docs/
  .github/workflows/
  go.mod
```

`internal/` is used for everything except `cmd/air`: nothing here is meant to be imported by
other Go modules. The plugin contract for third parties is "implement `agent.Agent`," not
"import our internals."

## Agent interface (`internal/agent`)

```go
type Agent interface {
    Name() string
    DetectInstalled() bool
    DetectVersion() (string, error)
    IsAvailable() (bool, error)
    Start(ctx context.Context, task Task) (*Session, error)
    Resume(ctx context.Context, checkpoint *checkpoint.Checkpoint) (*Session, error)
    Stop(session *Session) error
    HealthCheck(session *Session) HealthStatus
    DetectRateLimit(output string) *RateLimitInfo
    DetectCompletion(output string) bool
}
```

Providers only see `Task`, `Session`, `checkpoint.Checkpoint`, `HealthStatus`,
`RateLimitInfo` — plain structs owned by `internal/agent`. Each provider package is
self-contained: it knows how to shell out to its own CLI, parse its own stdout format, and
answer these questions. The router never branches on provider name.

## Router core loop (`internal/router`)

- Holds an ordered `[]agent.Agent`, built from the config `providers:` list (including
  `ollama:<model>` parsed into per-model agent instances).
- `Run(ctx, objective string)`: picks first available provider, `Start()`s it, streams
  stdout/stderr through `detect` heuristics on a ticker. On rate-limit/crash/quota/timeout:
  `checkpoint.Save()`, `summarizer.Compress()`, advance to next available provider,
  `Resume()`.
- Checkpoints save on a timer (`checkpoint_interval`) and on every detected
  failure/switch, so `air resume`/`air rollback` always have recent state regardless of
  which trigger fired.
- Router is pure orchestration — no provider-specific string matching lives here; that
  belongs to `detect` plus each provider's `DetectRateLimit`/`DetectCompletion`.

## Checkpoint format (`internal/checkpoint`)

One JSON file per checkpoint under `~/.air/checkpoints/<session-id>/<timestamp>.json`,
containing: objective, completed work, remaining tasks, git diff, modified files, terminal
output tail, conversation summary, errors, timestamp, provider used. `air checkpoints`
lists them; `air rollback <id>` restores the working tree described by the checkpoint's
git diff, behind a confirmation prompt since it mutates the user's working tree.

## Config (`internal/config`)

Viper loads `~/.air/config.yaml` (global) with an optional `./air.yaml` project override,
matching the shape:

```yaml
providers:
  - claude
  - codex
  - gemini
  - antigravity
  - ollama:qwen3-coder:30b
retry_failed: true
checkpoint_interval: 2m
summary_model: local
parallel_review: false
```

Struct-validated at load time so bad provider names/durations fail fast via `air doctor`/
`air config` rather than at runtime mid-session.

## CLI (`internal/cli`)

One Cobra command file per subcommand (`run`, `resume`, `status`, `doctor`, `providers`,
`checkpoints`, `rollback`, `logs`, `config`, `benchmark`, `update`). Commands stay thin —
real logic lives in `router`/`checkpoint`/`config` and is unit-tested independently of
Cobra.

## Testing approach

- `agent` ships a `mock` provider (in-package, not a separate provider dir) used by router
  tests — scriptable Start/Resume/failure injection.
- Router, checkpoint, config, and detect all get unit tests with no real CLI subprocess
  involved in Phase 1–3.
- Real provider packages (Phase 4+) get integration tests gated behind a build tag /
  `-short` skip, since they require the actual CLI installed on the test machine.

## Phased delivery

1. Repo scaffold + architecture (this doc, go.mod, package skeletons, CI, must compile)
2. Core router (event loop, checkpoint save/load, failure-detection interfaces)
3. Provider abstraction (Agent interface, registry, mock provider)
4. Claude Code provider
5. Gemini CLI provider
6. Remaining providers, TUI, cost tracking, benchmark, releases — one phase at a time,
   repository never left in a broken/non-compiling state between phases.
