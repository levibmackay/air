// Package tui implements AIR's Bubble Tea dashboard: current provider and
// task, elapsed time, checkpoint count, recent status, latest output, and
// keyboard shortcuts. It renders a live view of a Router run driven from
// internal/router's Event stream, without altering the router's own
// orchestration logic.
package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/levibmackay/air/internal/checkpoint"
	"github.com/levibmackay/air/internal/router"
)

// runner is the subset of Router's API the TUI drives; satisfied by
// *router.Router; a narrow interface here keeps this package testable
// without spinning up a real Router.
type runner interface {
	Run(ctx context.Context, sessionID, objective string) (*checkpoint.Checkpoint, error)
	Resume(ctx context.Context, cp *checkpoint.Checkpoint) (*checkpoint.Checkpoint, error)
}

// Run shows a live dashboard while r runs objective, and returns exactly
// what r.Run would have returned. events must be the same channel r was
// constructed with via router.WithEvents.
func Run(ctx context.Context, r runner, sessionID, objective string, events chan router.Event) (*checkpoint.Checkpoint, error) {
	return drive(ctx, sessionID, objective, events, func(runCtx context.Context) (*checkpoint.Checkpoint, error) {
		return r.Run(runCtx, sessionID, objective)
	})
}

// Resume shows a live dashboard while r resumes from cp, and returns
// exactly what r.Resume would have returned. events must be the same
// channel r was constructed with via router.WithEvents.
func Resume(ctx context.Context, r runner, cp *checkpoint.Checkpoint, events chan router.Event) (*checkpoint.Checkpoint, error) {
	return drive(ctx, cp.Session, cp.Objective, events, func(runCtx context.Context) (*checkpoint.Checkpoint, error) {
		return r.Resume(runCtx, cp)
	})
}

func drive(ctx context.Context, sessionID, objective string, events chan router.Event, run func(context.Context) (*checkpoint.Checkpoint, error)) (*checkpoint.Checkpoint, error) {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	m := initialModel(sessionID, objective, events, cancel)
	program := tea.NewProgram(m)

	go func() {
		cp, err := run(runCtx)
		close(events)
		program.Send(resultMsg{checkpoint: cp, err: err})
	}()

	finalModel, err := program.Run()
	if err != nil {
		return nil, err
	}

	fm := finalModel.(model)
	return fm.result, fm.runErr
}
