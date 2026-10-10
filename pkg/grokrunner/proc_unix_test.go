//go:build !windows

package grokrunner

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
)

func TestKillProcessGroup_NilAndInvalidSafe(t *testing.T) {
	if err := killProcessGroup(nil); err != nil {
		t.Fatalf("expected nil for nil cmd, got: %v", err)
	}

	cmdNoProcess := &exec.Cmd{}
	if err := killProcessGroup(cmdNoProcess); err != nil {
		t.Fatalf("expected nil for cmd without process, got: %v", err)
	}

	cmdZeroPid := &exec.Cmd{Process: &os.Process{Pid: 0}}
	if err := killProcessGroup(cmdZeroPid); err != nil {
		t.Fatalf("expected nil for cmd with pid 0, got: %v", err)
	}
}

func TestKillProcessGroup_SelfPgidProtected(t *testing.T) {
	// A command representing current process should NOT be killed
	myPid := os.Getpid()
	myPgid, err := syscall.Getpgid(myPid)
	if err != nil {
		t.Fatalf("failed to get my pgid: %v", err)
	}
	if myPgid <= 0 {
		t.Fatalf("invalid process group id: %d", myPgid)
	}

	cmdSelf := &exec.Cmd{Process: &os.Process{Pid: myPid}}
	// killProcessGroup must detect that pgid == myPgid and refrain from calling syscall.Kill(-pgid, ...)
	// We verify that the caller survives this call.
	_ = killProcessGroup(cmdSelf)

	// Verify we are still alive
	currPid := os.Getpid()
	if currPid != myPid {
		t.Fatalf("process mutated unexpectedly")
	}
}
