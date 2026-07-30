package lmstudio_test

import (
	"testing"

	"github.com/levibmackay/air/internal/providers/lmstudio"
)

func TestNew(t *testing.T) {
	agentDefault := lmstudio.New("")
	if agentDefault == nil {
		t.Fatal("expected non-nil agent")
	}
	if agentDefault.Name() != "lmstudio:local-model" {
		t.Errorf("expected name 'lmstudio:local-model', got %q", agentDefault.Name())
	}

	agentModel := lmstudio.New("qwen2.5-coder-7b")
	if agentModel.Name() != "lmstudio:qwen2.5-coder-7b" {
		t.Errorf("expected name 'lmstudio:qwen2.5-coder-7b', got %q", agentModel.Name())
	}
}
