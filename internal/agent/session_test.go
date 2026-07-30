package agent

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestSessionAppendAndOutput(t *testing.T) {
	s := NewSession("mock")
	s.AppendOutput("hello ")
	s.AppendOutput("world")
	if got := s.Output(); got != "hello world" {
		t.Errorf("Output() = %q, want %q", got, "hello world")
	}
}

func TestSessionMarkDoneIsIdempotent(t *testing.T) {
	s := NewSession("mock")
	err1 := errors.New("boom")
	s.MarkDone(err1)
	s.MarkDone(errors.New("ignored"))

	select {
	case <-s.Done():
	default:
		t.Fatal("Done() channel should be closed after MarkDone")
	}
	if s.Err() != err1 {
		t.Errorf("Err() = %v, want %v (first MarkDone should win)", s.Err(), err1)
	}
}

func TestSessionConcurrentAppendAndRead(t *testing.T) {
	s := NewSession("mock")
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.AppendOutput("x")
			_ = s.Output()
		}()
	}
	wg.Wait()
	if len(s.Output()) != 50 {
		t.Errorf("Output() length = %d, want 50", len(s.Output()))
	}
}

func TestSessionDoneBlocksUntilMarked(t *testing.T) {
	s := NewSession("mock")
	select {
	case <-s.Done():
		t.Fatal("Done() should not be closed before MarkDone")
	case <-time.After(10 * time.Millisecond):
	}
	s.MarkDone(nil)
	select {
	case <-s.Done():
	case <-time.After(time.Second):
		t.Fatal("Done() should close promptly after MarkDone")
	}
}
