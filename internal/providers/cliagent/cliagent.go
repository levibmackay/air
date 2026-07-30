// Package cliagent is a shared helper for provider plugins that drive an
// external, non-interactive CLI binary as a subprocess. It owns process
// lifecycle (launch, stream output into an agent.Session, detect exit) so
// individual provider packages only need to supply their binary name and
// argument shape.
package cliagent

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/levibmackay/air/internal/agent"
)

// Runner launches a single CLI binary and wires its output into an
// agent.Session.
type Runner struct {
	// Binary is the executable name (or path) to invoke.
	Binary string
}

// DetectInstalled reports whether Binary is present on PATH.
func (r Runner) DetectInstalled() bool {
	_, err := exec.LookPath(r.Binary)
	return err == nil
}

// DetectVersion runs "<binary> --version" and returns its trimmed output.
func (r Runner) DetectVersion() (string, error) {
	out, err := exec.Command(r.Binary, "--version").Output()
	if err != nil {
		return "", fmt.Errorf("%s --version: %w", r.Binary, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// Launch starts the binary with args in dir (the process's working
// directory; empty means inherit the AIR process's own). Combined
// stdout/stderr are streamed line-by-line into the returned Session as they
// arrive. Canceling ctx terminates the process.
func (r Runner) Launch(ctx context.Context, args []string, dir string) (*agent.Session, error) {
	cmd := exec.CommandContext(ctx, r.Binary, args...)
	if dir != "" {
		cmd.Dir = dir
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("%s: stdout pipe: %w", r.Binary, err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("%s: stderr pipe: %w", r.Binary, err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("%s: start: %w", r.Binary, err)
	}

	sess := agent.NewSession(r.Binary)
	sess.PID = cmd.Process.Pid

	var wg sync.WaitGroup
	wg.Add(2)
	go streamInto(sess, stdout, &wg)
	go streamInto(sess, stderr, &wg)

	go func() {
		wg.Wait()
		sess.MarkDone(cmd.Wait())
	}()

	return sess, nil
}

// streamInto copies r line-by-line into sess's output buffer until EOF.
func streamInto(sess *agent.Session, r io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		sess.AppendOutput(scanner.Text() + "\n")
	}
}
