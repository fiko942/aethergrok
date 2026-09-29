package updater

import (
	"archive/zip"
	"bytes"
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
	testFile := filepath.Join(tmpDir, "AetherGrok-Setup.exe")
	content := []byte("hello aethergrok auto-updater windows binary payload")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	hasher := sha256.New()
	hasher.Write(content)
	correctHash := hex.EncodeToString(hasher.Sum(nil))

	// 1. Test exact hash match
	valid, err := VerifyChecksum(testFile, correctHash)
	if err != nil || !valid {
		t.Fatalf("expected valid checksum, got err: %v", err)
	}

	// 2. Test sha256sum file formatting match (hash  filename)
	valid, err = VerifyChecksum(testFile, correctHash+"  AetherGrok-Setup.exe\n")
	if err != nil || !valid {
		t.Fatalf("expected valid checksum with sha256sum string, got err: %v", err)
	}

	// 3. Test sha256sum with asterisk binary flag (hash *filename)
	valid, err = VerifyChecksum(testFile, correctHash+" *AetherGrok-Setup.exe\n")
	if err != nil || !valid {
		t.Fatalf("expected valid checksum with asterisk binary notation, got err: %v", err)
	}

	// 4. Test multi-line SHA256SUMS.txt format with multiple assets
	multiLineChecksum := "1111111111111111111111111111111111111111111111111111111111111111  other-asset.zip\n" +
		correctHash + "  AetherGrok-Setup.exe\n" +
		"2222222222222222222222222222222222222222222222222222222222222222  another-asset.dmg\n"
	valid, err = VerifyChecksum(testFile, multiLineChecksum)
	if err != nil || !valid {
		t.Fatalf("expected multi-line checksum verification to pass, got err: %v", err)
	}

	// 5. Test multi-line SHA256SUMS.txt with wrong hash for this file
	multiLineWrong := "1111111111111111111111111111111111111111111111111111111111111111  other-asset.zip\n" +
		"0000000000000000000000000000000000000000000000000000000000000000  AetherGrok-Setup.exe\n"
	_, err = VerifyChecksum(testFile, multiLineWrong)
	if err == nil {
		t.Fatalf("expected checksum mismatch error on wrong file hash, got nil")
	}

	// 6. Test remote HTTP URL checksum download
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(correctHash + "  AetherGrok-Setup.exe\n"))
	}))
	defer ts.Close()

	valid, err = VerifyChecksum(testFile, ts.URL+"/AetherGrok-Setup.exe.sha256")
	if err != nil || !valid {
		t.Fatalf("expected remote URL checksum verification to pass, got err: %v", err)
	}

	// 7. Test direct hash mismatch
	_, err = VerifyChecksum(testFile, "0000000000000000000000000000000000000000000000000000000000000000")
	if err == nil {
		t.Fatalf("expected checksum mismatch error, got nil")
	}

	// 8. Test empty checksum (bypass)
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

func TestExtractZip(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test_update.zip")
	destDir := filepath.Join(tmpDir, "extracted")

	// Create a test zip file with normal files, directories, and a slip attempt
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// 1. Regular file
	f1, err := zw.Create("aethergrok.exe")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	_, _ = f1.Write([]byte("mock binary content"))

	// 2. Subdirectory file
	f2, err := zw.Create("resources/config.json")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	_, _ = f2.Write([]byte(`{"version":"1.0.2"}`))

	// 3. Zip Slip attempt (should be ignored safely)
	f3, err := zw.Create("../evil.txt")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	_, _ = f3.Write([]byte("malicious content"))

	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	if err := os.WriteFile(zipPath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("failed to write test zip: %v", err)
	}

	// Test extraction
	if err := ExtractZip(zipPath, destDir); err != nil {
		t.Fatalf("ExtractZip failed: %v", err)
	}

	// Verify regular file extracted
	extractedExe := filepath.Join(destDir, "aethergrok.exe")
	if data, err := os.ReadFile(extractedExe); err != nil || string(data) != "mock binary content" {
		t.Errorf("expected extracted exe content, got err: %v, data: %s", err, string(data))
	}

	// Verify nested file extracted
	extractedConfig := filepath.Join(destDir, "resources", "config.json")
	if data, err := os.ReadFile(extractedConfig); err != nil || string(data) != `{"version":"1.0.2"}` {
		t.Errorf("expected extracted config content, got err: %v, data: %s", err, string(data))
	}

	// Verify evil file was NOT extracted outside destDir
	evilFile := filepath.Join(tmpDir, "evil.txt")
	if _, err := os.Stat(evilFile); err == nil {
		t.Errorf("security vulnerability: evil.txt was extracted outside destDir!")
	}
}
