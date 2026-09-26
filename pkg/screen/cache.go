package screen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CacheStats represents statistics about screenshot cache on disk
type CacheStats struct {
	TotalBytes    int64  `json:"totalBytes"`
	FileCount     int    `json:"fileCount"`
	FormattedSize string `json:"formattedSize"`
}

// ClearCacheResult represents the outcome of clearing the cache
type ClearCacheResult struct {
	FreedBytes    int64  `json:"freedBytes"`
	DeletedCount  int    `json:"deletedCount"`
	FormattedSize string `json:"formattedSize"`
}

// formatBytes returns human-readable file size string (B, KB, MB, GB)
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetSnapshotCacheStats calculates the size and count of grok snapshot temporary files
func GetSnapshotCacheStats() (*CacheStats, error) {
	tempDir := os.TempDir()
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return &CacheStats{TotalBytes: 0, FileCount: 0, FormattedSize: "0 B"}, nil
	}

	var totalBytes int64
	var fileCount int

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "grok-snapshot-") {
			if info, err := entry.Info(); err == nil {
				totalBytes += info.Size()
				fileCount++
			}
		}
	}

	return &CacheStats{
		TotalBytes:    totalBytes,
		FileCount:     fileCount,
		FormattedSize: formatBytes(totalBytes),
	}, nil
}

// ClearSnapshotCache deletes all grok snapshot temporary files from the system temp directory
func ClearSnapshotCache() (*ClearCacheResult, error) {
	tempDir := os.TempDir()
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return &ClearCacheResult{FreedBytes: 0, DeletedCount: 0, FormattedSize: "0 B"}, nil
	}

	var freedBytes int64
	var deletedCount int

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "grok-snapshot-") {
			filePath := filepath.Join(tempDir, name)
			if info, err := entry.Info(); err == nil {
				if removeErr := os.Remove(filePath); removeErr == nil || os.IsNotExist(removeErr) {
					freedBytes += info.Size()
					deletedCount++
				}
			}
		}
	}

	return &ClearCacheResult{
		FreedBytes:    freedBytes,
		DeletedCount:  deletedCount,
		FormattedSize: formatBytes(freedBytes),
	}, nil
}

// DeleteSessionTempFiles removes temporary files referenced by a deleted session
// Safeguard: Only deletes files inside os.TempDir() or files with prefix "grok-snapshot-"
func DeleteSessionTempFiles(filePaths []string) error {
	tempDir := filepath.Clean(os.TempDir())

	for _, rawPath := range filePaths {
		clean := filepath.Clean(strings.TrimSpace(rawPath))
		if clean == "" {
			continue
		}

		// Safeguard check: must either reside inside system temp directory OR have grok-snapshot- prefix
		base := filepath.Base(clean)
		isGrokSnapshot := strings.HasPrefix(base, "grok-snapshot-")
		isInsideTemp := strings.HasPrefix(clean, tempDir)

		if isGrokSnapshot || isInsideTemp {
			// Best-effort removal without throwing if file is locked or already gone
			_ = os.Remove(clean)
		}
	}

	return nil
}
