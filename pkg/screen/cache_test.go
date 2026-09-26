package screen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotCacheLifecycle(t *testing.T) {
	tempDir := os.TempDir()

	// Create dummy snapshot files
	f1, err := os.CreateTemp(tempDir, "grok-snapshot-test1-*.png")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f1.Name())
	f1.WriteString("1234567890") // 10 bytes
	f1.Close()

	f2, err := os.CreateTemp(tempDir, "grok-snapshot-test2-*.jpg")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f2.Name())
	f2.WriteString("abcdefghijklmnopqrst") // 20 bytes
	f2.Close()

	// Also create a non-snapshot file that must NEVER be deleted
	safeFile, err := os.CreateTemp(tempDir, "safe-file-*.txt")
	if err != nil {
		t.Fatalf("failed to create safe temp file: %v", err)
	}
	defer os.Remove(safeFile.Name())
	safeFile.WriteString("critical user data")
	safeFile.Close()

	// 1. Check stats
	stats, err := GetSnapshotCacheStats()
	if err != nil {
		t.Fatalf("GetSnapshotCacheStats failed: %v", err)
	}
	if stats.FileCount < 2 {
		t.Errorf("expected at least 2 snapshot files, got %d", stats.FileCount)
	}
	if stats.TotalBytes < 30 {
		t.Errorf("expected at least 30 bytes, got %d", stats.TotalBytes)
	}

	// 2. Test DeleteSessionTempFiles safeguard against non-temp files
	cwd, _ := os.Getwd()
	workspaceFile := filepath.Join(cwd, "test_workspace_file.txt")
	_ = os.WriteFile(workspaceFile, []byte("do not touch"), 0644)
	defer os.Remove(workspaceFile)

	err = DeleteSessionTempFiles([]string{f1.Name(), workspaceFile})
	if err != nil {
		t.Errorf("DeleteSessionTempFiles returned error: %v", err)
	}

	// Verify f1 was deleted
	if _, err := os.Stat(f1.Name()); !os.IsNotExist(err) {
		t.Errorf("expected f1 to be deleted, but still exists")
	}

	// Verify workspaceFile was protected
	if _, err := os.Stat(workspaceFile); err != nil {
		t.Errorf("expected workspaceFile to be untouched, but got error: %v", err)
	}

	// 3. Test ClearSnapshotCache
	res, err := ClearSnapshotCache()
	if err != nil {
		t.Fatalf("ClearSnapshotCache failed: %v", err)
	}
	if res.DeletedCount < 1 {
		t.Errorf("expected at least 1 file deleted (f2), got %d", res.DeletedCount)
	}

	// Verify f2 was deleted
	if _, err := os.Stat(f2.Name()); !os.IsNotExist(err) {
		t.Errorf("expected f2 to be deleted by ClearSnapshotCache")
	}
}

func TestFormatBytes(t *testing.T) {
	cases := []struct {
		bytes    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, c := range cases {
		out := formatBytes(c.bytes)
		if out != c.expected {
			t.Errorf("formatBytes(%d) = %s; want %s", c.bytes, out, c.expected)
		}
	}
}
