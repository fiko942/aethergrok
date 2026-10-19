package grokrunner

import (
	"context"
	"os/exec"
	"testing"
)

func TestRunner_ActiveSessionTracking(t *testing.T) {
	r := &Runner{
		sessions: make(map[string]*ActiveSession),
	}

	if r.HasRunningSessions() {
		t.Fatalf("expected HasRunningSessions to be false initially")
	}
	if count := r.GetRunningSessionCount(); count != 0 {
		t.Fatalf("expected 0 running sessions, got %d", count)
	}

	ctx, cancel := context.WithCancel(context.Background())
	_ = ctx
	defer cancel()

	r.sessions["sess-1"] = &ActiveSession{
		SessionID: "sess-1",
		Cancel:    cancel,
	}

	if !r.HasRunningSessions() {
		t.Fatalf("expected HasRunningSessions to be true after adding active session")
	}
	if count := r.GetRunningSessionCount(); count != 1 {
		t.Fatalf("expected 1 running session, got %d", count)
	}

	delete(r.sessions, "sess-1")

	if r.HasRunningSessions() {
		t.Fatalf("expected HasRunningSessions to be false after deletion")
	}
	if count := r.GetRunningSessionCount(); count != 0 {
		t.Fatalf("expected 0 running sessions after deletion, got %d", count)
	}
}

func TestRunner_Cancel_ResilientLookup(t *testing.T) {
	r := NewRunner()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.Command("sleep", "60")
	if err := cmd.Start(); err != nil {
		t.Skip("sleep command unavailable")
	}
	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}()

	active := &ActiveSession{
		SessionID:     "sess_custom_abc",
		GrokSessionID: "01a15062-c047-7d91-a7e1-10e862e35c64",
		Cmd:           cmd,
		Cancel:        cancel,
		Done:          make(chan struct{}),
	}

	r.mu.Lock()
	r.sessions["sess_custom_abc"] = active
	r.mu.Unlock()

	// 1. Cancel by GrokSessionID
	err := r.Cancel("01a15062-c047-7d91-a7e1-10e862e35c64")
	if err != nil {
		t.Fatalf("unexpected cancel error: %v", err)
	}

	// Verify context is cancelled
	if ctx.Err() == nil {
		t.Errorf("expected session context to be cancelled")
	}

	// 2. Single running fallback test
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	cmd2 := exec.Command("sleep", "60")
	if err := cmd2.Start(); err == nil {
		defer func() {
			if cmd2.Process != nil {
				_ = cmd2.Process.Kill()
			}
		}()
		active2 := &ActiveSession{
			SessionID:     "sess_only_one",
			GrokSessionID: "uuid-single",
			Cmd:           cmd2,
			Cancel:        cancel2,
			Done:          make(chan struct{}),
		}
		r.mu.Lock()
		r.sessions = map[string]*ActiveSession{
			"sess_only_one": active2,
		}
		r.mu.Unlock()

		// Cancel with unmapped/unknown ID when exactly one session is active
		_ = r.Cancel("[object PointerEvent]")
		if ctx2.Err() == nil {
			t.Errorf("expected single active session to be cancelled via fallback")
		}
	}
}
