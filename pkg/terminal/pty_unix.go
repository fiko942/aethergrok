//go:build !windows

package terminal

import (
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/creack/pty"
)

func startPty(cmd *exec.Cmd, rows, cols int) (*osFileWrapper, error) {
	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
	if err != nil {
		return nil, err
	}

	if cmd.Process != nil {
		_ = syscall.Setpgid(cmd.Process.Pid, cmd.Process.Pid)
	}

	return &osFileWrapper{
		readCloser:  ptmx,
		writeCloser: ptmx,
		file:        ptmx,
	}, nil
}

func resizePty(file *osFileWrapper, rows, cols int) error {
	if file == nil || file.file == nil {
		return nil
	}
	return pty.Setsize(file.file, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
}

func killProcessTree(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	pid := cmd.Process.Pid
	pgid, err := syscall.Getpgid(pid)
	if err == nil && pgid > 0 {
		// Send SIGTERM to entire process group (-pgid)
		_ = syscall.Kill(-pgid, syscall.SIGTERM)

		done := make(chan error, 1)
		go func() {
			state, _ := cmd.Process.Wait()
			if state != nil && state.Exited() {
				done <- nil
			} else {
				done <- fmt.Errorf("still running")
			}
		}()

		select {
		case <-done:
			// Terminated cleanly
		case <-time.After(150 * time.Millisecond):
			// Escalate to SIGKILL for all child/grandchild processes
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		}
	} else {
		_ = cmd.Process.Kill()
	}

	return nil
}
