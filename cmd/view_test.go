package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestViewCmd(t *testing.T) {
	// Set test mode to prevent browser from opening
	testMode = true
	defer func() { testMode = false }()

	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	noteContent := "# Test Note\nThis is a test note."
	noteName := "test-note"
	notePath := filepath.Join(tmpDir, noteName+".md")

	err := os.WriteFile(notePath, []byte(noteContent), 0644)
	if err != nil {
		t.Fatalf("failed to create note file: %v", err)
	}

	// Run view command
	err = runView(viewCmd, []string{noteName})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("Temp HTML file: %s", testFile)

	// Read HTML content
	htmlContentBytes, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read HTML file: %v", err)
	}
	htmlContent := string(htmlContentBytes)

	t.Logf("HTML Content:\\n%s", htmlContent)

	expectedHTML := "<h1>Test Note</h1>\n\n<p>This is a test note.</p>\n"
	if !strings.Contains(htmlContent, expectedHTML) {
		t.Errorf("HTML content mismatch\\nwant to contain: %q\\n got: %q", expectedHTML, htmlContent)
	}
}
