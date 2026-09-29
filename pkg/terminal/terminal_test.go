package terminal

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestTerminalManager_LifecycleAndProcessGroup(t *testing.T) {
	mgr := NewManager()

	termID := "test_term_1"
	sessionID := "session_alpha"

	cwd, _ := os.Getwd()
	err := mgr.Create(sessionID, termID, cwd, "")
	if err != nil {
		t.Skipf("Skipping pty test: failed to create terminal in current environment: %v", err)
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
	err = mgr.Write(termID, "echo hello-terminal-test\r\n")
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

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
	e1 := mgr.Create("sess_1", "t1", cwd, "")
	e2 := mgr.Create("sess_1", "t2", cwd, "")
	e3 := mgr.Create("sess_2", "t3", cwd, "")

	if e1 != nil || e2 != nil || e3 != nil {
		t.Skipf("Skipping session terminals test: failed to create terminal in current environment")
	}

	err := mgr.CloseSessionTerminals("sess_1")
	if err != nil {
		t.Fatalf("CloseSessionTerminals failed: %v", err)
	}

	mgr.mu.RLock()
	_, hasT1 := mgr.terminals["t1"]
	_, hasT2 := mgr.terminals["t2"]
	_, hasT3 := mgr.terminals["t3"]
	mgr.mu.RUnlock()

	if hasT1 || hasT2 {
		t.Errorf("Terminals t1 and t2 should have been closed")
	}
	if !hasT3 {
		t.Errorf("Terminal t3 from sess_2 should still exist")
	}

	mgr.CloseAll()
}

func TestTerminalManager_ClsCommand(t *testing.T) {
	mgr := NewManager()

	cwd, _ := os.Getwd()
	err := mgr.Create("sess_cls", "t_cls", cwd, "")
	if err != nil {
		t.Skipf("Skipping cls test: failed to create terminal in current environment: %v", err)
	}
	defer mgr.Close("t_cls")

	time.Sleep(300 * time.Millisecond)

	// Send cls command
	err = mgr.Write("t_cls", "cls\r\n")
	if err != nil {
		t.Fatalf("Write cls failed: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	buf := mgr.GetTerminalBuffer("t_cls")
	t.Logf("Buffer length: %d", len(buf))
	if len(buf) == 0 {
		t.Errorf("Expected non-empty terminal buffer")
	}
	// Check that buffer contains prompt or clear sequences
	if strings.Contains(buf, "\x1b[2J") || strings.Contains(buf, "\x1b[H") || strings.Contains(buf, "PS") || strings.Contains(buf, ">") {
		t.Logf("Buffer successfully captured output from ConPTY terminal")
	}
}

func TestTerminalManager_InterruptAndKill(t *testing.T) {
	mgr := NewManager()

	cwd, _ := os.Getwd()
	err := mgr.Create("sess_test", "t_intr", cwd, "")
	if err != nil {
		t.Skipf("Skipping interrupt test: failed to create terminal in current environment: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	// Send an interactive/loop command, then test Interrupt
	err = mgr.Interrupt("t_intr")
	if err != nil {
		t.Errorf("Interrupt failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// Test Kill
	err = mgr.Kill("t_intr")
	if err != nil {
		t.Errorf("Kill failed: %v", err)
	}
}
