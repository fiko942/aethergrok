package grokrunner

import (
	"context"
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
