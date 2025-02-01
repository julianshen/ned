package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestEnv creates a temporary directory for testing and updates notesDir
func setupTestEnv(t *testing.T) (string, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ned-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Store original notesDir
	originalNotesDir := notesDir

	// Set notesDir to temp directory
	notesDir = tmpDir

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tmpDir)
		notesDir = originalNotesDir
	}

	return tmpDir, cleanup
}

// createTestNote creates a test note file with given content
func createTestNote(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		t.Fatalf("failed to create directories: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	return path
}

// assertFileContent checks if a file contains expected content
func assertFileContent(t *testing.T, path, expected string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(content) != expected {
		t.Errorf("file content mismatch\nwant: %q\ngot:  %q", expected, string(content))
	}
}

// assertFileExists checks if a file exists
func assertFileExists(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Errorf("file does not exist: %s", path)
	}
}
