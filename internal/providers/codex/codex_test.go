package codex_test

import (
	"testing"

	"github.com/levibmackay/air/internal/providers/codex"
)

func TestNew(t *testing.T) {
	agent := codex.New()
	if agent == nil {
		t.Fatal("expected non-nil agent")
	}
	if agent.Name() != "codex" {
		t.Errorf("expected name 'codex', got %q", agent.Name())
	}

	agentWithModel := codex.New("o3-mini")
	if agentWithModel.Name() != "codex:o3-mini" {
		t.Errorf("expected name 'codex:o3-mini', got %q", agentWithModel.Name())
	}
}
