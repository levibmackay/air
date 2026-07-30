package aider_test

import (
	"testing"

	"github.com/levibmackay/air/internal/providers/aider"
)

func TestNew(t *testing.T) {
	agent := aider.New()
	if agent == nil {
		t.Fatal("expected non-nil agent")
	}
	if agent.Name() != "aider" {
		t.Errorf("expected name 'aider', got %q", agent.Name())
	}

	agentWithModel := aider.New("gpt-4o")
	if agentWithModel.Name() != "aider:gpt-4o" {
		t.Errorf("expected name 'aider:gpt-4o', got %q", agentWithModel.Name())
	}
}
