package antigravity_test

import (
	"testing"

	"github.com/levibmackay/air/internal/providers/antigravity"
)

func TestNew(t *testing.T) {
	agent := antigravity.New()
	if agent == nil {
		t.Fatal("expected non-nil agent")
	}
	if agent.Name() != "antigravity" {
		t.Errorf("expected name 'antigravity', got %q", agent.Name())
	}
}
