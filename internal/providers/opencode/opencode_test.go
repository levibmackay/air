package opencode_test

import (
	"testing"

	"github.com/levibmackay/air/internal/providers/opencode"
)

func TestNew(t *testing.T) {
	agent := opencode.New()
	if agent == nil {
		t.Fatal("expected non-nil agent")
	}
	if agent.Name() != "opencode" {
		t.Errorf("expected name 'opencode', got %q", agent.Name())
	}

	agentWithModel := opencode.New("claude-3-5-sonnet")
	if agentWithModel.Name() != "opencode:claude-3-5-sonnet" {
		t.Errorf("expected name 'opencode:claude-3-5-sonnet', got %q", agentWithModel.Name())
	}
}
