//go:build windows

package grokrunner

import (
	"os/exec"
	"strconv"
	"syscall"
)

// SetSysProcGroup configures creation flags for Windows
func SetSysProcGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		HideWindow:    true,
	}
}

func setSysProcGroup(cmd *exec.Cmd) {
	SetSysProcGroup(cmd)
}

// killProcessGroup kills the process and its child processes on Windows using taskkill
func killProcessGroup(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	killCmd := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
	killCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := killCmd.Run(); err != nil {
		_ = cmd.Process.Kill()
	}
	return nil
}
