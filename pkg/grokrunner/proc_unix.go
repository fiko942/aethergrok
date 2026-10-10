//go:build !windows

package grokrunner

import (
	"os"
	"os/exec"
	"syscall"
)

// SetSysProcGroup configures Setpgid for non-Windows platforms
func SetSysProcGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

func setSysProcGroup(cmd *exec.Cmd) {
	SetSysProcGroup(cmd)
}

// killProcessGroup kills the entire process group cleanly without ever signaling self/parent
func killProcessGroup(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil || cmd.Process.Pid <= 0 {
		return nil
	}

	myPgid, _ := syscall.Getpgid(os.Getpid())
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err == nil && pgid > 1 && pgid != myPgid {
		return syscall.Kill(-pgid, syscall.SIGKILL)
	}

	if cmd.Process.Pid != os.Getpid() {
		return cmd.Process.Kill()
	}
	return nil
}
