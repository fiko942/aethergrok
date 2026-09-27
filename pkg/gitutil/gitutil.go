package gitutil

import (
	"os/exec"
	"runtime"
)

// Executable returns the most appropriate path or command name for git on the current OS.
// On macOS (darwin), /usr/bin/git is an xcrun wrapper shim that dynamically links against
// CommandLineTools libxcrun.dylib. If an app runs in x86_64 or arm64 where CommandLineTools
// only provides arm64/arm64e libraries, invoking "/usr/bin/git" triggers xcrun which fails with:
// "unable to load libxcrun ... fat file, but missing compatible architecture".
// Checking direct binary locations such as CommandLineTools, Homebrew, or Xcode avoids the xcrun shim.
// On Windows and Linux, it returns "git" or finds git in PATH.
func Executable() string {
	if runtime.GOOS == "darwin" {
		candidates := []string{
			"/opt/homebrew/bin/git",
			"/usr/local/bin/git",
			"/Library/Developer/CommandLineTools/usr/bin/git",
			"/Applications/Xcode.app/Contents/Developer/usr/bin/git",
		}
		for _, candidate := range candidates {
			if path, err := exec.LookPath(candidate); err == nil {
				return path
			}
		}
	}
	return "git"
}
