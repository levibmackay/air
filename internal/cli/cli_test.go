package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/levibmackay/air/internal/checkpoint"
)

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestPrintCheckpointSummary(t *testing.T) {
	cp := &checkpoint.Checkpoint{
		ID:       "2026-07-30T12-00-00",
		Provider: "claude",
		Created:  time.Now(),
		Errors:   []string{"rate limit hit"},
	}

	out := captureOutput(func() {
		printCheckpointSummary(cp)
	})

	if !strings.Contains(out, "2026-07-30T12-00-00") || !strings.Contains(out, "claude") || !strings.Contains(out, "rate limit hit") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestPrintProviderStatus(t *testing.T) {
	out := captureOutput(func() {
		err := printProviderStatus([]string{"claude", "invalid_provider_name"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "providers:") {
		t.Errorf("expected header 'providers:', got: %s", out)
	}
}
