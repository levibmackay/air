package summarizer

import (
	"strings"
	"testing"

	"github.com/levibmackay/air/internal/checkpoint"
)

func TestCompressIncludesKeySections(t *testing.T) {
	cp := &checkpoint.Checkpoint{
		Objective:       "Build a REST API",
		CompletedWork:   []string{"Scaffolded router", "Added /health endpoint"},
		RemainingTasks:  []string{"Add auth middleware"},
		ModifiedFiles:   []string{"main.go", "router.go"},
		GitDiff:         "diff --git a/main.go b/main.go",
		Errors:          []string{"rate limit: quota exceeded"},
		ConversationSum: "User asked for a REST API in Go.",
		TerminalOutput:  "server listening on :8080",
	}

	got := Compress(cp)

	for _, want := range []string{
		cp.Objective,
		"Scaffolded router",
		"Add auth middleware",
		"main.go",
		"diff --git a/main.go b/main.go",
		"rate limit: quota exceeded",
		"User asked for a REST API in Go.",
		"server listening on :8080",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Compress() missing %q in output:\n%s", want, got)
		}
	}
}

func TestCompressHandlesEmptyCheckpoint(t *testing.T) {
	cp := &checkpoint.Checkpoint{Objective: "Build a REST API"}
	got := Compress(cp)
	if !strings.Contains(got, cp.Objective) {
		t.Errorf("Compress() on minimal checkpoint missing objective: %s", got)
	}
}

func TestTailOfTruncatesFromEnd(t *testing.T) {
	s := "abcdef"
	got := tailOf(s, 3)
	if got != "def" {
		t.Errorf("tailOf(%q, 3) = %q, want %q", s, got, "def")
	}
	if tailOf(s, 100) != s {
		t.Errorf("tailOf with n > len should return input unchanged")
	}
}
