package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestBenchmarkCmd(t *testing.T) {
	var buf bytes.Buffer
	err := runBenchmark(&buf, "Test benchmark objective", []string{"unsupported_provider"}, 1*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Benchmarking objective") {
		t.Errorf("expected header in output, got: %s", out)
	}
	if !strings.Contains(out, "unsupported_provider") {
		t.Errorf("expected provider name in output, got: %s", out)
	}
}
