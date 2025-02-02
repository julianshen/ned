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

	// Create test images
	testImages := []struct {
		path    string
		content []byte
	}{
		{
			path:    filepath.Join("._images_", "test1.jpg"),
			content: []byte("fake image 1"),
		},
		{
			path:    filepath.Join("subfolder", "._images_", "test2.png"),
			content: []byte("fake image 2"),
		},
	}

	for _, img := range testImages {
		fullPath := filepath.Join(tmpDir, img.path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory for %s: %v", img.path, err)
		}
		if err := os.WriteFile(fullPath, img.content, 0644); err != nil {
			t.Fatalf("Failed to create test image %s: %v", img.path, err)
		}
	}

	noteContent := `# Test Note
This is a test note with images:
![Local Image](test1.jpg)
![Subfolder Image](subfolder/test2.png)`
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

	expectedHTML := "<h1>Test Note</h1>"
	if !strings.Contains(htmlContent, expectedHTML) {
		t.Errorf("HTML content mismatch\nwant to contain: %q\ngot: %q", expectedHTML, htmlContent)
	}

	// Check if image paths are transformed correctly
	expectedLocalPath := filepath.Join(tmpDir, "._images_", "test1.jpg")
	expectedSubfolderPath := filepath.Join(tmpDir, "subfolder", "._images_", "test2.png")

	if !strings.Contains(htmlContent, expectedLocalPath) {
		t.Errorf("Local image path not transformed correctly\nwant to contain: %q\ngot: %q", expectedLocalPath, htmlContent)
	}

	if !strings.Contains(htmlContent, expectedSubfolderPath) {
		t.Errorf("Subfolder image path not transformed correctly\nwant to contain: %q\ngot: %q", expectedSubfolderPath, htmlContent)
	}
}

func TestTransformImagePaths(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name     string
		content  string
		noteDir  string
		expected string
	}{
		{
			name:     "local image",
			content:  "![Alt](test.jpg)",
			noteDir:  tmpDir,
			expected: "![Alt](" + strings.ReplaceAll(filepath.Join(tmpDir, "._images_", "test.jpg"), "\\", "\\\\") + ")",
		},
		{
			name:     "subfolder image",
			content:  "![Alt](folder/test.png)",
			noteDir:  tmpDir,
			expected: "![Alt](" + strings.ReplaceAll(filepath.Join(notesDir, "folder", "._images_", "test.png"), "\\", "\\\\") + ")",
		},
		{
			name:     "multiple images",
			content:  "![One](test1.jpg)\n![Two](folder/test2.png)",
			noteDir:  tmpDir,
			expected: "![One](" + strings.ReplaceAll(filepath.Join(tmpDir, "._images_", "test1.jpg"), "\\", "\\\\") + ")\n![Two](" + strings.ReplaceAll(filepath.Join(notesDir, "folder", "._images_", "test2.png"), "\\", "\\\\") + ")",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformImagePaths(tt.content, tt.noteDir)
			if result != tt.expected {
				t.Errorf("transformImagePaths() = %v, want %v", result, tt.expected)
			}
		})
	}
}
