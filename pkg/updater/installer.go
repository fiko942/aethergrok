package updater

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// UpdateProgress models the progress and status during background download and installation
type UpdateProgress struct {
	DownloadedBytes int64   `json:"downloadedBytes"`
	TotalBytes      int64   `json:"totalBytes"`
	Percent         float64 `json:"percent"`
	SpeedFormatted  string  `json:"speedFormatted"`
	Stage           string  `json:"stage"` // "downloading" | "verifying" | "installing" | "ready" | "error"
	Message         string  `json:"message"`
}

// DownloadManager coordinates cancellable downloads
type DownloadManager struct {
	mu     sync.Mutex
	cancel context.CancelFunc
}

var GlobalDownloadManager = &DownloadManager{}

// SetCancel stores current active download cancel func
func (m *DownloadManager) SetCancel(cancel context.CancelFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cancel = cancel
}

// CancelActive aborts the active download if running
func (m *DownloadManager) CancelActive() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
}

// FormatSpeed formats transfer speed in KB/s or MB/s
func FormatSpeed(bytesPerSec float64) string {
	if bytesPerSec >= 1024*1024 {
		return fmt.Sprintf("%.2f MB/s", bytesPerSec/(1024*1024))
	}
	if bytesPerSec >= 1024 {
		return fmt.Sprintf("%.1f KB/s", bytesPerSec/1024)
	}
	return fmt.Sprintf("%.0f B/s", bytesPerSec)
}

// DownloadAssetWithProgress downloads a remote file to a local temp folder while streaming progress callbacks
func DownloadAssetWithProgress(ctx context.Context, downloadURL string, targetFilename string, onProgress func(UpdateProgress)) (string, error) {
	if downloadURL == "" {
		return "", fmt.Errorf("download URL cannot be empty")
	}

	tempDir := filepath.Join(os.TempDir(), "aethergrok_update")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create update temp directory: %w", err)
	}

	if targetFilename == "" {
		parts := strings.Split(downloadURL, "/")
		targetFilename = parts[len(parts)-1]
		if idx := strings.Index(targetFilename, "?"); idx != -1 {
			targetFilename = targetFilename[:idx]
		}
	}
	if targetFilename == "" {
		targetFilename = "update_package.bin"
	}

	destPath := filepath.Join(tempDir, targetFilename)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create download request: %w", err)
	}
	req.Header.Set("User-Agent", "AetherGrok-Updater")

	client := &http.Client{
		Timeout: 30 * time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to start download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("download server returned HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	totalBytes := resp.ContentLength

	out, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create target file: %w", err)
	}
	defer out.Close()

	buf := make([]byte, 64*1024) // 64KB chunk buffer
	var downloadedBytes int64
	startTime := time.Now()
	lastSampleTime := startTime
	var lastSampleBytes int64
	var currentSpeedFormatted = "0 KB/s"

	for {
		select {
		case <-ctx.Done():
			out.Close()
			os.Remove(destPath)
			return "", ctx.Err()
		default:
		}

		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return "", fmt.Errorf("error writing download chunk: %w", writeErr)
			}
			downloadedBytes += int64(n)

			now := time.Now()
			elapsedSinceSample := now.Sub(lastSampleTime)
			if elapsedSinceSample >= 400*time.Millisecond {
				bytesDiff := downloadedBytes - lastSampleBytes
				speed := float64(bytesDiff) / elapsedSinceSample.Seconds()
				currentSpeedFormatted = FormatSpeed(speed)
				lastSampleTime = now
				lastSampleBytes = downloadedBytes

				var percent float64
				if totalBytes > 0 {
					percent = float64(downloadedBytes) / float64(totalBytes) * 100.0
					if percent > 100.0 {
						percent = 100.0
					}
				}

				if onProgress != nil {
					onProgress(UpdateProgress{
						DownloadedBytes: downloadedBytes,
						TotalBytes:      totalBytes,
						Percent:         percent,
						SpeedFormatted:  currentSpeedFormatted,
						Stage:           "downloading",
						Message:         fmt.Sprintf("Downloading update (%.1f MB / %.1f MB)", float64(downloadedBytes)/(1024*1024), float64(totalBytes)/(1024*1024)),
					})
				}
			}
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return "", fmt.Errorf("error reading download stream: %w", readErr)
		}
	}

	// Final progress report for 100% download complete
	if onProgress != nil {
		onProgress(UpdateProgress{
			DownloadedBytes: downloadedBytes,
			TotalBytes:      downloadedBytes,
			Percent:         100.0,
			SpeedFormatted:  currentSpeedFormatted,
			Stage:           "downloading",
			Message:         "Download completed",
		})
	}

	return destPath, nil
}

// ComputeSHA256 returns the hexadecimal SHA256 checksum of a local file
func ComputeSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// VerifyChecksum compares local file's SHA256 against expected hash or remote checksum content
func VerifyChecksum(filePath string, expectedHashOrURL string) (bool, error) {
	cleanExpected := strings.TrimSpace(expectedHashOrURL)
	if cleanExpected == "" {
		// If no checksum is provided, consider verification bypassed
		return true, nil
	}

	actualHash, err := ComputeSHA256(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to compute file checksum: %w", err)
	}

	// If expected is a URL, fetch content first
	if strings.HasPrefix(cleanExpected, "http://") || strings.HasPrefix(cleanExpected, "https://") {
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Get(cleanExpected)
		if err != nil {
			return false, fmt.Errorf("failed to download checksum file: %w", err)
		}
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, fmt.Errorf("failed to read downloaded checksum file: %w", err)
		}
		cleanExpected = string(bodyBytes)
	}

	cleanExpected = strings.ToLower(cleanExpected)
	actualHash = strings.ToLower(actualHash)

	// Check if multi-line checksum file has a match for this specific filename
	fileName := strings.ToLower(filepath.Base(filePath))
	lines := strings.Split(cleanExpected, "\n")
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		if strings.Contains(line, fileName) {
			fields := strings.Fields(line)
			for _, field := range fields {
				cleanField := strings.Trim(field, "* \t\r\n")
				if len(cleanField) == 64 {
					if cleanField == actualHash {
						return true, nil
					}
					return false, fmt.Errorf("checksum mismatch for %s: expected '%s', got '%s'", filepath.Base(filePath), cleanField, actualHash)
				}
			}
		}
	}

	// Check if the expected string contains actualHash (handles raw hashes or simple sha256 files)
	if strings.Contains(cleanExpected, actualHash) {
		return true, nil
	}

	return false, fmt.Errorf("checksum mismatch: expected '%s', got '%s'", strings.TrimSpace(cleanExpected), actualHash)
}

// ExtractZip extracts a zip archive to the target destination directory securely.
func ExtractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer r.Close()

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	for _, f := range r.File {
		// Prevent Zip Slip vulnerability (path traversal)
		cleanName := filepath.Clean(f.Name)
		if strings.HasPrefix(cleanName, "..") || filepath.IsAbs(cleanName) || strings.Contains(cleanName, ":") {
			continue
		}

		targetPath := filepath.Join(destDir, cleanName)

		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(targetPath, f.Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for file %s: %w", targetPath, err)
		}

		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", targetPath, err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("failed to open zip entry %s: %w", f.Name, err)
		}

		_, copyErr := io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()

		if copyErr != nil {
			return fmt.Errorf("failed to extract file %s: %w", targetPath, copyErr)
		}
	}

	return nil
}

// ApplyUpdateMacOS mounts the downloaded DMG, locates AetherGrok.app, executes a detached helper script to replace the app and relaunch, then signals the host process to terminate.
func ApplyUpdateMacOS(dmgPath string, onProgress func(UpdateProgress)) error {
	if onProgress != nil {
		onProgress(UpdateProgress{
			Stage:   "installing",
			Percent: 100,
			Message: "Mounting update disk image...",
		})
	}

	// 1. Mount DMG quietly
	mountPoint := filepath.Join(os.TempDir(), fmt.Sprintf("aethergrok_mnt_%d", time.Now().UnixNano()))
	_ = os.MkdirAll(mountPoint, 0755)

	cmd := exec.Command("hdiutil", "attach", dmgPath, "-mountpoint", mountPoint, "-nobrowse", "-noverify", "-noautoopen", "-quiet")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to mount DMG (%s): %s", err, string(out))
	}

	// Ensure cleanup helper will unmount upon failure
	unmount := func() {
		_ = exec.Command("hdiutil", "detach", mountPoint, "-force", "-quiet").Run()
	}

	// 2. Locate .app inside mount point
	entries, err := os.ReadDir(mountPoint)
	if err != nil {
		unmount()
		return fmt.Errorf("failed to read mount point: %w", err)
	}

	var appName string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".app") {
			appName = entry.Name()
			break
		}
	}

	if appName == "" {
		appName = "AetherGrok.app"
	}
	sourceAppPath := filepath.Join(mountPoint, appName)
	if _, err := os.Stat(sourceAppPath); err != nil {
		unmount()
		return fmt.Errorf("application bundle %s not found in disk image", appName)
	}

	// 3. Determine current application install target path
	currentExec, err := os.Executable()
	if err != nil {
		unmount()
		return fmt.Errorf("failed to determine executable path: %w", err)
	}

	// Find the .app directory ancestor
	targetAppPath := "/Applications/AetherGrok.app"
	parts := strings.Split(filepath.Clean(currentExec), string(filepath.Separator))
	for i := len(parts) - 1; i >= 0; i-- {
		if strings.HasSuffix(parts[i], ".app") {
			targetAppPath = "/" + filepath.Join(parts[:i+1]...)
			break
		}
	}

	if onProgress != nil {
		onProgress(UpdateProgress{
			Stage:   "installing",
			Percent: 100,
			Message: "Applying update and preparing relaunch...",
		})
	}

	// 4. Create detached updater shell script
	scriptPath := filepath.Join(os.TempDir(), fmt.Sprintf("aethergrok_apply_%d.sh", time.Now().UnixNano()))
	pid := os.Getpid()

	scriptContent := fmt.Sprintf(`#!/bin/bash
# Wait for parent process to exit
while kill -0 %d 2>/dev/null; do
    sleep 0.2
done

# Copy new .app bundle over old one
rm -rf "%s"
cp -R "%s" "%s"

# Remove quarantine attribute if present
xattr -rd com.apple.quarantine "%s" 2>/dev/null || true

# Unmount DMG
hdiutil detach "%s" -force -quiet 2>/dev/null || true
rm -f "%s"

# Relaunch newly installed application
open -n "%s"

# Self-delete script
rm -f "%s"
`, pid, targetAppPath, sourceAppPath, targetAppPath, targetAppPath, mountPoint, dmgPath, targetAppPath, scriptPath)

	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		unmount()
		return fmt.Errorf("failed to create update script: %w", err)
	}

	// 5. Spawn updater script detached
	updaterCmd := exec.Command("/bin/bash", scriptPath)
	if err := updaterCmd.Start(); err != nil {
		unmount()
		return fmt.Errorf("failed to start detached updater script: %w", err)
	}

	if onProgress != nil {
		onProgress(UpdateProgress{
			Stage:   "ready",
			Percent: 100,
			Message: "Update ready. Restarting AetherGrok...",
		})
	}

	return nil
}

// ApplyUpdateWindows executes the downloaded NSIS setup executable or extracts a portable zip archive
func ApplyUpdateWindows(filePath string, onProgress func(UpdateProgress)) error {
	cleanPath := filepath.Clean(filePath)
	lowerName := strings.ToLower(cleanPath)

	if strings.HasSuffix(lowerName, ".zip") {
		if onProgress != nil {
			onProgress(UpdateProgress{
				Stage:   "installing",
				Percent: 100,
				Message: "Extracting portable package...",
			})
		}

		// Destination directory next to zip
		baseDir := filepath.Dir(cleanPath)
		targetFolderName := strings.TrimSuffix(filepath.Base(cleanPath), filepath.Ext(cleanPath))
		destDir := filepath.Join(baseDir, targetFolderName)

		if err := ExtractZip(cleanPath, destDir); err != nil {
			return fmt.Errorf("failed to extract portable zip: %w", err)
		}

		// Reveal extracted folder in Windows Explorer
		_ = exec.Command("explorer.exe", filepath.Clean(destDir)).Start()

		if onProgress != nil {
			onProgress(UpdateProgress{
				Stage:   "ready",
				Percent: 100,
				Message: fmt.Sprintf("Update extracted to %s. Opened folder in Explorer.", filepath.Base(destDir)),
			})
		}
		return nil
	}

	// Executable setup or standalone .exe
	if onProgress != nil {
		onProgress(UpdateProgress{
			Stage:   "installing",
			Percent: 100,
			Message: "Launching Windows installer...",
		})
	}

	// If installer executable (setup or installer)
	if strings.Contains(lowerName, "setup") || strings.Contains(lowerName, "installer") {
		cmd := exec.Command(cleanPath, "/S")
		if err := cmd.Start(); err != nil {
			// Fallback without /S for interactive wizard mode
			cmd = exec.Command(cleanPath)
			if err := cmd.Start(); err != nil {
				return fmt.Errorf("failed to launch Windows installer: %w", err)
			}
		}
	} else {
		// Standalone executable
		cmd := exec.Command(cleanPath)
		if err := cmd.Start(); err != nil {
			_ = exec.Command("explorer.exe", fmt.Sprintf("/select,%s", cleanPath)).Start()
			return fmt.Errorf("failed to execute binary: %w", err)
		}
	}

	if onProgress != nil {
		onProgress(UpdateProgress{
			Stage:   "ready",
			Percent: 100,
			Message: "Installer started. Closing application for update...",
		})
	}

	return nil
}
