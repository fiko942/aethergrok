package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestFormatSpeed(t *testing.T) {
	tests := []struct {
		bytesPerSec float64
		expected    string
	}{
		{500, "500 B/s"},
		{1024, "1.0 KB/s"},
		{1536, "1.5 KB/s"},
		{1024 * 1024, "1.00 MB/s"},
		{10.5 * 1024 * 1024, "10.50 MB/s"},
	}

	for _, tt := range tests {
		got := FormatSpeed(tt.bytesPerSec)
		if got != tt.expected {
			t.Errorf("FormatSpeed(%v) = %s; want %s", tt.bytesPerSec, got, tt.expected)
		}
	}
}

func TestVerifyChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("hello aethergrok auto-updater")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	hasher := sha256.New()
	hasher.Write(content)
	correctHash := hex.EncodeToString(hasher.Sum(nil))

	// Test exact hash match
	valid, err := VerifyChecksum(testFile, correctHash)
	if err != nil || !valid {
		t.Fatalf("expected valid checksum, got err: %v", err)
	}

	// Test sha256sum file formatting match (hash  filename)
	valid, err = VerifyChecksum(testFile, correctHash+"  test.txt\n")
	if err != nil || !valid {
		t.Fatalf("expected valid checksum with sha256sum string, got err: %v", err)
	}

	// Test mismatch
	_, err = VerifyChecksum(testFile, "0000000000000000000000000000000000000000000000000000000000000000")
	if err == nil {
		t.Fatalf("expected checksum mismatch error, got nil")
	}

	// Test empty checksum (bypass)
	valid, err = VerifyChecksum(testFile, "")
	if err != nil || !valid {
		t.Fatalf("expected empty hash to pass verification bypass, got err: %v", err)
	}
}

func TestDownloadAssetWithProgress(t *testing.T) {
	payload := strings.Repeat("AetherGrok update package content block. ", 1000)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(payload))
	}))
	defer server.Close()

	var recordedProgress []UpdateProgress
	progressCb := func(p UpdateProgress) {
		recordedProgress = append(recordedProgress, p)
	}

	filePath, err := DownloadAssetWithProgress(context.Background(), server.URL, "test_asset.bin", progressCb)
	if err != nil {
		t.Fatalf("DownloadAssetWithProgress failed: %v", err)
	}
	defer os.Remove(filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(data) != payload {
		t.Fatalf("downloaded content mismatch")
	}

	if len(recordedProgress) == 0 {
		t.Fatalf("expected progress callbacks to be invoked")
	}
	last := recordedProgress[len(recordedProgress)-1]
	if last.Percent != 100.0 {
		t.Errorf("expected final progress to be 100%%, got %.2f%%", last.Percent)
	}
}
