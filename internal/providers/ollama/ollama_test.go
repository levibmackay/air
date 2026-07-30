package ollama_test

import (
	"testing"

	"github.com/levibmackay/air/internal/providers/ollama"
)

func TestNew(t *testing.T) {
	agentDefault := ollama.New("")
	if agentDefault == nil {
		t.Fatal("expected non-nil agent")
	}
	if agentDefault.Name() != "ollama:qwen2.5-coder" {
		t.Errorf("expected name 'ollama:qwen2.5-coder', got %q", agentDefault.Name())
	}

	agentModel := ollama.New("qwen3-coder:30b")
	if agentModel.Name() != "ollama:qwen3-coder:30b" {
		t.Errorf("expected name 'ollama:qwen3-coder:30b', got %q", agentModel.Name())
	}
}
