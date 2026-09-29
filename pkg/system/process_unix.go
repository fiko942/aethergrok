//go:build !windows

package system

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func initAppProcessGroup() error {
	return nil
}

func cleanupOrphanProcesses() {
	currentPID := os.Getpid()
	if currentPID <= 0 {
		return
	}

	out, err := exec.Command("pgrep", "-P", strconv.Itoa(currentPID)).Output()
	if err != nil {
		return
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if childPID, err := strconv.Atoi(line); err == nil && childPID > 0 && childPID != currentPID {
			_ = syscall.Kill(childPID, syscall.SIGKILL)
		}
	}
}
