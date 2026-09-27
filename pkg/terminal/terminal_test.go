package terminal

import (
	"os"
	"testing"
	"time"
)

func TestTerminalManager_LifecycleAndProcessGroup(t *testing.T) {
	mgr := NewManager()

	termID := "test_term_1"
	sessionID := "session_alpha"

	cwd, _ := os.Getwd()
	err := mgr.Create(sessionID, termID, cwd, "/bin/sh")
	if err != nil {
		t.Fatalf("Failed to create terminal: %v", err)
	}

	// Verify terminal is in manager
	mgr.mu.RLock()
	inst, ok := mgr.terminals[termID]
	mgr.mu.RUnlock()

	if !ok || inst == nil {
		t.Fatalf("Terminal instance not stored in manager")
	}

	// Resize check
	err = mgr.Resize(termID, 100, 30)
	if err != nil {
		t.Errorf("Resize failed: %v", err)
	}

	// Write check (echo test)
	err = mgr.Write(termID, "echo hello-terminal-test\n")
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Close terminal and ensure process group cleanup
	err = mgr.Close(termID)
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Should not be in manager anymore
	mgr.mu.RLock()
	_, stillExists := mgr.terminals[termID]
	mgr.mu.RUnlock()

	if stillExists {
		t.Errorf("Terminal should have been removed from manager on Close()")
	}
}

func TestTerminalManager_CloseSessionTerminals(t *testing.T) {
	mgr := NewManager()

	cwd, _ := os.Getwd()
	_ = mgr.Create("sess_1", "t1", cwd, "/bin/sh")
	_ = mgr.Create("sess_1", "t2", cwd, "/bin/sh")
	_ = mgr.Create("sess_2", "t3", cwd, "/bin/sh")

	err := mgr.CloseSessionTerminals("sess_1")
	if err != nil {
		t.Fatalf("CloseSessionTerminals failed: %v", err)
	}

	mgr.mu.RLock()
	t1Exists := mgr.terminals["t1"] != nil
	t2Exists := mgr.terminals["t2"] != nil
	t3Exists := mgr.terminals["t3"] != nil
	mgr.mu.RUnlock()

	if t1Exists {
		t.Errorf("t1 should have been closed")
	}
	if t2Exists {
		t.Errorf("t2 should have been closed")
	}
	if !t3Exists {
		t.Errorf("t3 should still be alive")
	}
	_ = mgr.Close("t3")
}
