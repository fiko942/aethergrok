//go:build windows

package terminal

import (
	"fmt"
	"os/exec"
	"strconv"
	"syscall"
)

func startPty(cmd *exec.Cmd, rows, cols int) (*osFileWrapper, error) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to open stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to open stdout pipe: %w", err)
	}

	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	return &osFileWrapper{
		readCloser:  stdout,
		writeCloser: stdin,
		file:        nil,
	}, nil
}

func resizePty(file *osFileWrapper, rows, cols int) error {
	// PTY resize on Windows pipe fallback is a no-op
	return nil
}

func killProcessTree(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	// Terminate entire process tree using taskkill /T /F /PID <pid>
	killCmd := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
	if err := killCmd.Run(); err != nil {
		return cmd.Process.Kill()
	}

	return nil
}
