package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_workspace_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test subfolder and files
	subDir := filepath.Join(tempDir, "src")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}

	testFile := filepath.Join(tempDir, "package.json")
	if err := os.WriteFile(testFile, []byte(`{"name":"test"}`), 0644); err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}

	items, err := ReadDirectory(tempDir, "")
	if err != nil {
		t.Fatalf("ReadDirectory failed: %v", err)
	}

	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}

	// Directories first
	if !items[0].IsDir || items[0].Name != "src" {
		t.Errorf("expected first item to be directory 'src', got %+v", items[0])
	}
	if items[1].IsDir || items[1].Name != "package.json" {
		t.Errorf("expected second item to be file 'package.json', got %+v", items[1])
	}
}

func TestReadFileContent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_workspace_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "hello.txt")
	content := "Hello, World!"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	read, err := ReadFileContent(tempDir, "hello.txt")
	if err != nil {
		t.Fatalf("ReadFileContent failed: %v", err)
	}

	if read != content {
		t.Errorf("expected %q, got %q", content, read)
	}
}
