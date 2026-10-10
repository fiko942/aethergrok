//go:build !windows

package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
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

// killChildProcessesOfUnix finds and terminates child processes of parentPID
func killChildProcessesOfUnix(parentPID int, sig syscall.Signal) error {
	if parentPID <= 0 || parentPID == os.Getpid() {
		return nil
	}

	out, err := exec.Command("pgrep", "-P", strconv.Itoa(parentPID)).Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if childPID, err := strconv.Atoi(line); err == nil && childPID > 1 && childPID != parentPID && childPID != os.Getpid() {
			_ = killChildProcessesOfUnix(childPID, sig)
			_ = syscall.Kill(childPID, sig)
		}
	}
	return nil
}

// interruptProcess sends Ctrl+C and terminates any running foreground child processes
func interruptProcess(cmd *exec.Cmd, ptmx *osFileWrapper) error {
	if ptmx != nil {
		_, _ = ptmx.Write([]byte{3}) // \x03 (Ctrl+C)
	}

	if cmd != nil && cmd.Process != nil && cmd.Process.Pid > 0 && cmd.Process.Pid != os.Getpid() {
		pid := cmd.Process.Pid
		myPgid, _ := syscall.Getpgid(os.Getpid())
		pgid, err := syscall.Getpgid(pid)
		if err == nil && pgid > 1 && pgid != myPgid {
			_ = syscall.Kill(-pgid, syscall.SIGINT)
		}
		_ = killChildProcessesOfUnix(pid, syscall.SIGINT)
	}

	return nil
}

// killProcessTree terminates the command and its full process tree safely
func killProcessTree(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil || cmd.Process.Pid <= 0 || cmd.Process.Pid == os.Getpid() {
		return nil
	}

	pid := cmd.Process.Pid
	myPgid, _ := syscall.Getpgid(os.Getpid())
	pgid, err := syscall.Getpgid(pid)
	if err == nil && pgid > 1 && pgid != myPgid {
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
	}

	_ = killChildProcessesOfUnix(pid, syscall.SIGKILL)
	_ = cmd.Process.Kill()
	return nil
}
