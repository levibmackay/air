package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/levibmackay/air/internal/checkpoint"
)

func TestStatusLoadSessionCheckpoints(t *testing.T) {
	dir := t.TempDir()
	store := checkpoint.NewStore(dir)

	cp := &checkpoint.Checkpoint{
		Session:   "sess-123",
		Provider:  "claude",
		Created:   time.Now(),
		Objective: "Test Status",
		WorkDir:   dir,
	}

	if err := store.Save(cp); err != nil {
		t.Fatalf("failed to save checkpoint: %v", err)
	}

	cps, err := loadSessionCheckpoints(store, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cps) != 1 {
		t.Fatalf("expected 1 checkpoint, got %d", len(cps))
	}
	if cps[0].Session != "sess-123" {
		t.Errorf("expected session 'sess-123', got %q", cps[0].Session)
	}

	_ = os.RemoveAll(filepath.Join(dir, "sess-123"))
}
