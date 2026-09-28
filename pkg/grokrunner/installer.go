package grokrunner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"aethergrok/pkg/logger"
)

type GrokInstallStatus struct {
	Installed  bool   `json:"installed"`
	Version    string `json:"version"`
	BinaryPath string `json:"binaryPath"`
	Platform   string `json:"platform"`
	Error      string `json:"error,omitempty"`
}

type GrokInstallProgress struct {
	Stage   string `json:"stage"`   // "preparing", "downloading", "installing", "verifying", "completed", "error"
	Percent int    `json:"percent"` // 0 - 100
	Message string `json:"message"`
	LogLine string `json:"logLine,omitempty"`
}

var versionRegex = regexp.MustCompile(`(?i)grok\s+v?([0-9]+\.[0-9]+(\.[0-9]+)?)`)

// DetectGrokInstallation checks whether the grok CLI binary is available and executable.
func DetectGrokInstallation() GrokInstallStatus {
	binPath := ResolveGrokBinary()

	// If resolved path is just fallback "grok" or does not exist on disk
	if binPath == "" || binPath == "grok" {
		if _, err := exec.LookPath("grok"); err != nil {
			logger.GetDiskLogger().Append(logger.LogEntry{
				Timestamp: time.Now().UnixMilli(),
				Level:     "WARN",
				Category:  "grok_installer",
				Message:   "Grok CLI detection: binary not found",
				Details: map[string]interface{}{
					"platform": runtime.GOOS,
					"binPath":  binPath,
				},
			})
			return GrokInstallStatus{
				Installed: false,
				Platform:  runtime.GOOS,
				Error:     "Grok CLI binary not found in PATH or standard install directories",
			}
		}
	} else if _, err := os.Stat(binPath); err != nil {
		logger.GetDiskLogger().Append(logger.LogEntry{
			Timestamp: time.Now().UnixMilli(),
			Level:     "WARN",
			Category:  "grok_installer",
			Message:   fmt.Sprintf("Grok CLI detection: binary path %s does not exist", binPath),
			Details: map[string]interface{}{
				"platform": runtime.GOOS,
				"binPath":  binPath,
				"error":    err.Error(),
			},
		})
		return GrokInstallStatus{
			Installed: false,
			Platform:  runtime.GOOS,
			Error:     fmt.Sprintf("Resolved binary path '%s' not accessible: %v", binPath, err),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binPath, "--version")
	outBytes, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(outBytes))

	if err != nil {
		errMsg := fmt.Sprintf("failed executing %s --version: %v (output: %s)", binPath, err, output)
		logger.GetDiskLogger().Append(logger.LogEntry{
			Timestamp: time.Now().UnixMilli(),
			Level:     "ERROR",
			Category:  "grok_installer",
			Message:   "Grok CLI detection: version check failed",
			Details: map[string]interface{}{
				"platform": runtime.GOOS,
				"binPath":  binPath,
				"error":    errMsg,
			},
		})
		return GrokInstallStatus{
			Installed:  false,
			BinaryPath: binPath,
			Platform:   runtime.GOOS,
			Error:      errMsg,
		}
	}

	version := parseGrokVersion(output)
	logger.GetDiskLogger().Append(logger.LogEntry{
		Timestamp: time.Now().UnixMilli(),
		Level:     "INFO",
		Category:  "grok_installer",
		Message:   "Grok CLI detected successfully",
		Details: map[string]interface{}{
			"platform": runtime.GOOS,
			"binPath":  binPath,
			"version":  version,
			"raw":      output,
		},
	})

	return GrokInstallStatus{
		Installed:  true,
		Version:    version,
		BinaryPath: binPath,
		Platform:   runtime.GOOS,
	}
}

func parseGrokVersion(raw string) string {
	matches := versionRegex.FindStringSubmatch(raw)
	if len(matches) >= 2 {
		return matches[1]
	}
	// Fallback to first line if regex did not match
	lines := strings.Split(raw, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return raw
}

// InstallGrokCLI executes automated installation on macOS and emits progress updates.
func InstallGrokCLI(ctx context.Context, onProgress func(GrokInstallProgress)) (*GrokInstallStatus, error) {
	if onProgress == nil {
		onProgress = func(GrokInstallProgress) {}
	}

	if runtime.GOOS != "darwin" {
		err := fmt.Errorf("automated installation is currently only supported on macOS (detected %s)", runtime.GOOS)
		onProgress(GrokInstallProgress{
			Stage:   "error",
			Percent: 0,
			Message: err.Error(),
			LogLine: err.Error(),
		})
		logger.GetDiskLogger().Append(logger.LogEntry{
			Timestamp: time.Now().UnixMilli(),
			Level:     "ERROR",
			Category:  "grok_installer",
			Message:   "Installation aborted: unsupported platform",
			Details:   map[string]interface{}{"platform": runtime.GOOS},
		})
		return nil, err
	}

	onProgress(GrokInstallProgress{
		Stage:   "preparing",
		Percent: 5,
		Message: "Checking system prerequisites and package managers...",
	})

	logger.GetDiskLogger().Append(logger.LogEntry{
		Timestamp: time.Now().UnixMilli(),
		Level:     "INFO",
		Category:  "grok_installer",
		Message:   "Beginning Grok CLI installation sequence",
		Details:   map[string]interface{}{"platform": runtime.GOOS},
	})

	// Check Homebrew availability
	brewPath := ""
	for _, path := range []string{"/opt/homebrew/bin/brew", "/usr/local/bin/brew"} {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			brewPath = path
			break
		}
	}
	if brewPath == "" {
		if p, err := exec.LookPath("brew"); err == nil {
			brewPath = p
		}
	}

	var installErr error
	var recentLogLines []string
	recordLog := func(line string) {
		recentLogLines = append(recentLogLines, line)
		if len(recentLogLines) > 50 {
			recentLogLines = recentLogLines[len(recentLogLines)-50:]
		}
	}

	if brewPath != "" {
		onProgress(GrokInstallProgress{
			Stage:   "downloading",
			Percent: 20,
			Message: "Homebrew detected. Running `brew install grok`...",
		})
		logger.GetDiskLogger().Append(logger.LogEntry{
			Timestamp: time.Now().UnixMilli(),
			Level:     "INFO",
			Category:  "grok_installer",
			Message:   fmt.Sprintf("Attempting Strategy 1: Homebrew installation using %s", brewPath),
		})

		cmd := exec.CommandContext(ctx, brewPath, "install", "grok")
		installErr = streamCommand(cmd, func(line string) {
			recordLog(line)
			logger.GetDiskLogger().Append(logger.LogEntry{
				Timestamp: time.Now().UnixMilli(),
				Level:     "DEBUG",
				Category:  "grok_installer",
				Message:   line,
			})
			onProgress(GrokInstallProgress{
				Stage:   "installing",
				Percent: 45,
				Message: "Installing Grok via Homebrew...",
				LogLine: line,
			})
		})
	}

	// Strategy 2: If Homebrew was not found or failed, use official install script
	if brewPath == "" || installErr != nil {
		if installErr != nil {
			logger.GetDiskLogger().Append(logger.LogEntry{
				Timestamp: time.Now().UnixMilli(),
				Level:     "WARN",
				Category:  "grok_installer",
				Message:   fmt.Sprintf("Homebrew installation failed: %v. Falling back to official install script.", installErr),
			})
		}

		onProgress(GrokInstallProgress{
			Stage:   "downloading",
			Percent: 30,
			Message: "Running official install script (https://x.ai/cli/install.sh)...",
		})
		logger.GetDiskLogger().Append(logger.LogEntry{
			Timestamp: time.Now().UnixMilli(),
			Level:     "INFO",
			Category:  "grok_installer",
			Message:   "Attempting Strategy 2: curl -fsSL https://x.ai/cli/install.sh | bash",
		})

		cmd := exec.CommandContext(ctx, "bash", "-c", "curl -fsSL https://x.ai/cli/install.sh | bash")
		installErr = streamCommand(cmd, func(line string) {
			recordLog(line)
			logger.GetDiskLogger().Append(logger.LogEntry{
				Timestamp: time.Now().UnixMilli(),
				Level:     "DEBUG",
				Category:  "grok_installer",
				Message:   line,
			})
			onProgress(GrokInstallProgress{
				Stage:   "installing",
				Percent: 65,
				Message: "Running installer script...",
				LogLine: line,
			})
		})
	}

	onProgress(GrokInstallProgress{
		Stage:   "verifying",
		Percent: 85,
		Message: "Verifying Grok CLI installation...",
	})

	// Allow a brief moment for filesystem flush
	time.Sleep(500 * time.Millisecond)

	status := DetectGrokInstallation()
	if status.Installed {
		onProgress(GrokInstallProgress{
			Stage:   "completed",
			Percent: 100,
			Message: fmt.Sprintf("Grok CLI v%s installed successfully at %s", status.Version, status.BinaryPath),
			LogLine: fmt.Sprintf("Installation completed: version %s", status.Version),
		})
		logger.GetDiskLogger().Append(logger.LogEntry{
			Timestamp: time.Now().UnixMilli(),
			Level:     "INFO",
			Category:  "grok_installer",
			Message:   "Grok CLI installation verified successfully",
			Details: map[string]interface{}{
				"version":    status.Version,
				"binaryPath": status.BinaryPath,
			},
		})
		return &status, nil
	}

	lastOutput := strings.Join(recentLogLines, "\n")
	errFinal := fmt.Errorf("installation failed or grok binary not found: %s\nRecent logs:\n%s", status.Error, lastOutput)
	onProgress(GrokInstallProgress{
		Stage:   "error",
		Percent: 100,
		Message: "Grok CLI installation verification failed",
		LogLine: errFinal.Error(),
	})
	logger.GetDiskLogger().Append(logger.LogEntry{
		Timestamp: time.Now().UnixMilli(),
		Level:     "ERROR",
		Category:  "grok_installer",
		Message:   "Grok CLI installation verification failed",
		Details: map[string]interface{}{
			"error":      errFinal.Error(),
			"recentLogs": recentLogLines,
		},
	})
	return &status, errFinal
}

func streamCommand(cmd *exec.Cmd, lineHandler func(string)) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	reader := io.MultiReader(stdout, stderr)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		text := strings.TrimRight(scanner.Text(), "\r\n")
		if text != "" {
			lineHandler(text)
		}
	}

	return cmd.Wait()
}
