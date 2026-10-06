package grokrunner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveSessionArgs_ContinuationExistingFolder(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	sessionsDir := filepath.Join(home, ".grok", "sessions")
	wsDir := filepath.Join(t.TempDir(), "ws-resume-test")
	_ = os.MkdirAll(wsDir, 0755)

	targetWsDir := ResolveWorkspaceSessionsDir(sessionsDir, wsDir)
	_ = os.MkdirAll(targetWsDir, 0755)
	defer os.RemoveAll(targetWsDir)

	sessUUID := "12345678-1234-1234-1234-123456789abc"
	sessFolder := filepath.Join(targetWsDir, sessUUID)
	_ = os.MkdirAll(sessFolder, 0755)
	_ = os.WriteFile(filepath.Join(sessFolder, "chat_history.jsonl"), []byte("{}\n"), 0644)

	args, isResume, resolvedID := ResolveSessionArgs(sessionsDir, targetWsDir, sessUUID)

	if !isResume {
		t.Fatalf("Expected isResume=true for existing session directory, got false")
	}
	if resolvedID != sessUUID {
		t.Errorf("Expected resolvedID '%s', got '%s'", sessUUID, resolvedID)
	}
	if len(args) != 2 || args[0] != "--resume" || args[1] != sessUUID {
		t.Errorf("Expected args ['--resume', '%s'], got %v", sessUUID, args)
	}
}

func TestResolveSessionArgs_NewSessionWithUUID(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	sessionsDir := filepath.Join(home, ".grok", "sessions")
	wsDir := filepath.Join(t.TempDir(), "ws-new-test")
	_ = os.MkdirAll(wsDir, 0755)

	targetWsDir := ResolveWorkspaceSessionsDir(sessionsDir, wsDir)
	sessUUID := "87654321-4321-4321-4321-cba987654321"

	args, isResume, resolvedID := ResolveSessionArgs(sessionsDir, targetWsDir, sessUUID)

	if isResume {
		t.Fatalf("Expected isResume=false for non-existent session directory, got true")
	}
	if resolvedID != sessUUID {
		t.Errorf("Expected resolvedID '%s', got '%s'", sessUUID, resolvedID)
	}
	if len(args) != 2 || args[0] != "--session-id" || args[1] != sessUUID {
		t.Errorf("Expected args ['--session-id', '%s'], got %v", sessUUID, args)
	}
}

func TestResolveSessionArgs_NewSessionEmpty(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	sessionsDir := filepath.Join(home, ".grok", "sessions")
	wsDir := filepath.Join(t.TempDir(), "ws-empty-test")
	_ = os.MkdirAll(wsDir, 0755)

	targetWsDir := ResolveWorkspaceSessionsDir(sessionsDir, wsDir)

	args, isResume, resolvedID := ResolveSessionArgs(sessionsDir, targetWsDir, "")

	if isResume {
		t.Fatalf("Expected isResume=false for empty session ID, got true")
	}
	if !isUUID(resolvedID) {
		t.Errorf("Expected generated UUID, got '%s'", resolvedID)
	}
	if len(args) != 2 || args[0] != "--session-id" || args[1] != resolvedID {
		t.Errorf("Expected args ['--session-id', '%s'], got %v", resolvedID, args)
	}
}

func TestResolveSessionArgs_IsolationBetweenWorkspaces(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	sessionsDir := filepath.Join(home, ".grok", "sessions")
	ws1 := filepath.Join(t.TempDir(), "ws1")
	ws2 := filepath.Join(t.TempDir(), "ws2")
	_ = os.MkdirAll(ws1, 0755)
	_ = os.MkdirAll(ws2, 0755)

	ws1Dir := ResolveWorkspaceSessionsDir(sessionsDir, ws1)
	ws2Dir := ResolveWorkspaceSessionsDir(sessionsDir, ws2)
	_ = os.MkdirAll(ws1Dir, 0755)
	_ = os.MkdirAll(ws2Dir, 0755)
	defer os.RemoveAll(ws1Dir)
	defer os.RemoveAll(ws2Dir)

	sessUUID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	sessFolder1 := filepath.Join(ws1Dir, sessUUID)
	_ = os.MkdirAll(sessFolder1, 0755)
	_ = os.WriteFile(filepath.Join(sessFolder1, "chat_history.jsonl"), []byte("{}\n"), 0644)

	// ws1 should resume sessUUID
	args1, isResume1, _ := ResolveSessionArgs(sessionsDir, ws1Dir, sessUUID)
	if !isResume1 || args1[0] != "--resume" {
		t.Fatalf("Expected ws1 to resume sessUUID, got args: %v", args1)
	}

	// ws2 must NOT resume sessUUID under any circumstances
	args2, isResume2, _ := ResolveSessionArgs(sessionsDir, ws2Dir, sessUUID)
	if isResume2 || args2[0] != "--session-id" {
		t.Fatalf("Expected ws2 to NOT resume sessUUID from ws1, got args: %v", args2)
	}
}
