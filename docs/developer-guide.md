# Developer Guide

## Prerequisites

- Go 1.26+
- `git` on PATH (used for `air rollback` and checkpoint diffs)
- Whichever provider CLIs you want to actually run (`claude`, `gemini`, ...)
  — everything else works without them installed.

## Building and testing

```bash
go build ./...          # must always succeed on main
go vet ./...
go test ./...            # add -race locally before pushing anything touching
                          # internal/router, internal/agent, or internal/tui
```

CI (`.github/workflows/ci.yml`) runs all three on Linux, macOS, and Windows
for every push and PR.

## Repository layout

See [`docs/superpowers/specs/2026-07-30-air-architecture-design.md`](superpowers/specs/2026-07-30-air-architecture-design.md)
for the full rationale. In short:

- `cmd/air` — entrypoint, nothing but `cli.Execute()`.
- `internal/cli` — one file per subcommand; commands are thin, logic lives
  below.
- `internal/agent` — the `Agent` plugin interface, `Session`, `Registry`,
  and `Mock` (for tests).
- `internal/router` — orchestration: launches providers, polls for
  completion/rate-limit/health, saves checkpoints, switches providers,
  emits `Event`s for the TUI.
- `internal/checkpoint` — the `Checkpoint` struct and its JSON file store.
- `internal/detect` — shared regex heuristics for classifying provider
  output (rate limit, context overflow, connection lost, ...).
- `internal/summarizer` — turns a `Checkpoint` into a resume prompt.
- `internal/config` — Viper-backed YAML config, layered
  defaults → `~/.air/config.yaml` → `./air.yaml`.
- `internal/cost` — elapsed-time/token/cost ledger built from a session's
  checkpoint history.
- `internal/tui` — the Bubble Tea dashboard (`air run --tui`).
- `internal/providers/*` — one package per provider; see
  [plugin-guide.md](plugin-guide.md) to add one.

## Testing patterns

- **Router tests** (`internal/router/router_test.go`) never touch a real
  process. They script `agent.Mock` to simulate completion, rate limits,
  crashes, and provider switching, then assert on the router's outcome and
  the checkpoint it produced.
- **cliagent tests** (`internal/providers/cliagent/cliagent_test.go`) do
  spawn real (trivial, `sh -c ...`) subprocesses to verify output streaming
  and exit-code handling; they skip on Windows since they rely on `sh`.
- **TUI tests** (`internal/tui/model_test.go`) exercise `model.Update`/
  `model.View` directly — no real terminal or `tea.Program` involved, since
  Bubble Tea needs a real TTY that CI doesn't have.
- Every package that isn't a thin CLI wrapper has tests; `go test -race` is
  clean across the whole module.

## Local end-to-end smoke testing

Since `air run` really shells out to whatever provider you've configured,
don't smoke-test it against a real CLI casually — it'll actually try to do
the work. To exercise the full pipeline safely, point PATH at a fake binary:

```bash
mkdir -p /tmp/fakebin
cat > /tmp/fakebin/claude <<'EOF'
#!/bin/sh
echo "fake claude got args: $@"
exit 0
EOF
chmod +x /tmp/fakebin/claude

PATH="/tmp/fakebin:$PATH" go run ./cmd/air run "test objective"
```

## Logging

`internal/cli` wires a `zap.NewProduction()` logger into the router for
`air run`/`air resume`, except when `--tui` is passed — the dashboard owns
the terminal, so the logger is swapped for `zap.NewNop()` to avoid
corrupting the rendered UI. If you need router-level diagnostics while
using `--tui`, check `~/.air/checkpoints/<session>/*.json` after the fact
instead.
