package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmptyListCmd(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	// Set notes directory to the temp directory
	notesDir = tmpDir

	// Run list command on empty directory
	err = runList(listCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w.Close()

	// Read captured output
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read captured output: %v", err)
	}
	os.Stdout = oldStdout

	// Check output
	output := strings.TrimSpace(buf.String())
	expected := "Empty"

	if output != expected {
		t.Errorf("output mismatch\nwant: %q\ngot:  %q", expected, output)
	}
}

func TestListCmd(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create test directory structure
	files := []struct {
		path    string
		content string
	}{
		{"note1.md", "test content 1"},
		{"dir1/note2.md", "test content 2"},
		{"dir1/note3.md", "test content 3"},
		{"dir1/subdir/note4.md", "test content 4"},
		{"dir2/note5.md", "test content 5"},
	}

	for _, f := range files {
		path := filepath.Join(tmpDir, f.path)
		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		err = os.WriteFile(path, []byte(f.content), 0644)
		if err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	// Run list command
	err = runList(listCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w.Close()

	// Read captured output
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read captured output: %v", err)
	}
	os.Stdout = oldStdout

	// Check output
	output := buf.String()
	expected := []string{
		"Notes structure:",
		"├── dir1",
		"    ├── note2",
		"    ├── note3",
		"    └── subdir",
		"        └── note4",
		"├── dir2",
		"    └── note5",
		"└── note1",
	}

	// Convert both expected and actual output to slice of lines for comparison
	expectedLines := strings.Split(strings.TrimSpace(strings.Join(expected, "\n")), "\n")
	actualLines := strings.Split(strings.TrimSpace(output), "\n")

	if len(actualLines) != len(expectedLines) {
		t.Errorf("output line count mismatch\nwant: %d\ngot:  %d", len(expectedLines), len(actualLines))
		return
	}

	for i := range expectedLines {
		if strings.TrimSpace(actualLines[i]) != strings.TrimSpace(expectedLines[i]) {
			t.Errorf("line %d mismatch\nwant: %q\ngot:  %q", i, expectedLines[i], actualLines[i])
		}
	}
}
