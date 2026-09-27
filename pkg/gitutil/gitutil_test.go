package gitutil

import (
	"os/exec"
	"testing"
)

func TestExecutable(t *testing.T) {
	gitPath := Executable()
	if gitPath == "" {
		t.Fatalf("expected git executable to not be empty")
	}

	cmd := exec.Command(gitPath, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run git (%s): %v, output: %s", gitPath, err, string(out))
	}
	if len(out) == 0 {
		t.Fatalf("expected output from git --version")
	}
}
