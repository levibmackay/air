package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/levibmackay/air/internal/checkpoint"
	"github.com/levibmackay/air/internal/router"
)

func newTestModel() model {
	events := make(chan router.Event)
	return initialModel("sess-1", "build a REST API", events, func() {})
}

func TestModelViewNeverPanicsInitially(t *testing.T) {
	m := newTestModel()
	view := m.View()
	if !strings.Contains(view, "sess-1") {
		t.Errorf("View() = %q, want it to contain the session ID", view)
	}
	if !strings.Contains(view, "build a REST API") {
		t.Errorf("View() missing objective")
	}
}

func TestModelHandlesProviderStartedEvent(t *testing.T) {
	m := newTestModel()
	updated, _ := m.Update(router.Event{Type: router.EventProviderStarted, Provider: "claude"})
	m2 := updated.(model)

	if m2.provider != "claude" {
		t.Errorf("provider = %q, want claude", m2.provider)
	}
	if !strings.Contains(m2.View(), "claude") {
		t.Errorf("View() should mention the active provider")
	}
}

func TestModelHandlesCheckpointEvent(t *testing.T) {
	m := newTestModel()
	cp := &checkpoint.Checkpoint{TerminalOutput: "line one\nline two\n"}
	updated, _ := m.Update(router.Event{Type: router.EventCheckpoint, Provider: "claude", Checkpoint: cp})
	m2 := updated.(model)

	if m2.checkpoints != 1 {
		t.Errorf("checkpoints = %d, want 1", m2.checkpoints)
	}
	if len(m2.outputTail) != 2 || m2.outputTail[0] != "line one" || m2.outputTail[1] != "line two" {
		t.Errorf("outputTail = %v, want [line one, line two]", m2.outputTail)
	}
}

func TestModelHandlesFailureEvent(t *testing.T) {
	m := newTestModel()
	cp := &checkpoint.Checkpoint{Errors: []string{"rate limit: quota exceeded"}}
	updated, _ := m.Update(router.Event{Type: router.EventProviderFailed, Provider: "claude", Checkpoint: cp})
	m2 := updated.(model)

	if m2.lastError != "rate limit: quota exceeded" {
		t.Errorf("lastError = %q, want %q", m2.lastError, "rate limit: quota exceeded")
	}
}

func TestModelHandlesResultAndCompletionView(t *testing.T) {
	m := newTestModel()
	m.provider = "claude"
	updated, _ := m.Update(resultMsg{checkpoint: &checkpoint.Checkpoint{}, err: nil})
	m2 := updated.(model)

	if !m2.done {
		t.Fatal("done should be true after resultMsg")
	}
	if !strings.Contains(m2.View(), "completed") {
		t.Errorf("View() after success should mention completion: %s", m2.View())
	}
}

func TestModelQuitCallsCancel(t *testing.T) {
	canceled := false
	events := make(chan router.Event)
	m := initialModel("sess-1", "obj", events, func() { canceled = true })

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if !canceled {
		t.Error("ctrl+c should call the cancel func")
	}
	if cmd == nil {
		t.Error("ctrl+c should return a quit command")
	}
}
