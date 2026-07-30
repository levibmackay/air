package cliagent

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"
)

func skipOnWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("test relies on sh -c, not available on windows")
	}
}

func TestRunnerLaunchCapturesOutputAndCleanExit(t *testing.T) {
	skipOnWindows(t)
	r := Runner{Binary: "sh"}

	sess, err := r.Launch(context.Background(), []string{"-c", "echo hello; echo world"}, "")
	if err != nil {
		t.Fatalf("Launch() error = %v", err)
	}

	select {
	case <-sess.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("session did not finish in time")
	}

	if err := sess.Err(); err != nil {
		t.Fatalf("Err() = %v, want nil for a clean exit", err)
	}
	out := sess.Output()
	if !strings.Contains(out, "hello") || !strings.Contains(out, "world") {
		t.Errorf("Output() = %q, want it to contain both lines", out)
	}
}

func TestRunnerLaunchCapturesNonZeroExit(t *testing.T) {
	skipOnWindows(t)
	r := Runner{Binary: "sh"}

	sess, err := r.Launch(context.Background(), []string{"-c", "exit 1"}, "")
	if err != nil {
		t.Fatalf("Launch() error = %v", err)
	}

	<-sess.Done()
	if sess.Err() == nil {
		t.Fatal("Err() = nil, want a non-nil error for exit code 1")
	}
}

func TestRunnerLaunchCancelKillsProcess(t *testing.T) {
	skipOnWindows(t)
	r := Runner{Binary: "sh"}
	ctx, cancel := context.WithCancel(context.Background())

	sess, err := r.Launch(ctx, []string{"-c", "sleep 30"}, "")
	if err != nil {
		t.Fatalf("Launch() error = %v", err)
	}

	cancel()

	select {
	case <-sess.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("session did not exit promptly after context cancellation")
	}
	if sess.Err() == nil {
		t.Error("Err() = nil, want a non-nil error after being killed")
	}
}

func TestRunnerDetectInstalled(t *testing.T) {
	skipOnWindows(t)
	if !(Runner{Binary: "sh"}).DetectInstalled() {
		t.Error("DetectInstalled() = false for sh, want true")
	}
	if (Runner{Binary: "definitely-not-a-real-binary-xyz"}).DetectInstalled() {
		t.Error("DetectInstalled() = true for a nonexistent binary, want false")
	}
}
