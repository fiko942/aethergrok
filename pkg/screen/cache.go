package screen

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"
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

// SaveTemporaryImage writes base64 image data to a temporary file prefixed with grok-snapshot-drop-
// It compresses large images (> 650KB) to high-quality JPEG and returns a populated SnapshotResult
func SaveTemporaryImage(base64Data string, mimeType string) (*SnapshotResult, error) {
	cleanData := base64Data
	if commaIdx := strings.Index(base64Data, ","); commaIdx != -1 {
		cleanData = base64Data[commaIdx+1:]
	}

	rawBytes, err := base64.StdEncoding.DecodeString(cleanData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 image data: %w", err)
	}

	if len(rawBytes) == 0 {
		return nil, fmt.Errorf("image data buffer is empty")
	}

	finalBytes := rawBytes
	outMime := mimeType
	if outMime == "" {
		outMime = "image/png"
	}
	fileExt := "png"
	if strings.Contains(outMime, "jpeg") || strings.Contains(outMime, "jpg") {
		fileExt = "jpg"
	} else if strings.Contains(outMime, "webp") {
		fileExt = "webp"
	}

	width := 0
	height := 0
	const maxSizeBytes = 650 * 1024

	// Inspect dimensions & compress if large
	img, _, decodeErr := image.Decode(bytes.NewReader(rawBytes))
	if decodeErr == nil && img != nil {
		bounds := img.Bounds()
		width = bounds.Dx()
		height = bounds.Dy()

		if len(rawBytes) > maxSizeBytes {
			quality := 85
			for quality >= 50 {
				var buf bytes.Buffer
				err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
				if err == nil {
					compressed := buf.Bytes()
					if len(compressed) <= maxSizeBytes || quality == 50 {
						finalBytes = compressed
						outMime = "image/jpeg"
						fileExt = "jpg"
						break
					}
				}
				quality -= 10
			}
		}
	} else if fileExt == "png" {
		// Attempt png decode if general decode failed
		if pImg, pErr := png.Decode(bytes.NewReader(rawBytes)); pErr == nil && pImg != nil {
			bounds := pImg.Bounds()
			width = bounds.Dx()
			height = bounds.Dy()
		}
	}

	tmpFile, err := os.CreateTemp("", fmt.Sprintf("grok-snapshot-drop-*.%s", fileExt))
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary image file: %w", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(finalBytes); err != nil {
		_ = os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to write image data to temp file: %w", err)
	}

	encodedB64 := base64.StdEncoding.EncodeToString(finalBytes)
	dataURL := fmt.Sprintf("data:%s;base64,%s", outMime, encodedB64)

	return &SnapshotResult{
		FilePath:  tmpFile.Name(),
		DataURL:   dataURL,
		Base64:    encodedB64,
		Width:     width,
		Height:    height,
		Size:      int64(len(finalBytes)),
		Timestamp: time.Now().UnixMilli(),
	}, nil
}
