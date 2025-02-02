package cmd

import (
	"os"
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

// assertFileExists checks if a file exists
func assertFileExists(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Errorf("file does not exist: %s", path)
	}
}
