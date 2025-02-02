package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWelcomePage(t *testing.T) {
	// Set test mode to prevent browser from opening
	testMode = true
	defer func() { testMode = false }()

	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create some test notes
	testNotes := []struct {
		name    string
		content string
	}{
		{
			name:    "note1.md",
			content: "# Note 1\nThis is test note 1",
		},
		{
			name:    "folder/note2.md",
			content: "# Note 2\nThis is test note 2",
		},
		{
			name:    "folder/subfolder/note3.md",
			content: "# Note 3\nThis is test note 3",
		},
	}

	for _, note := range testNotes {
		fullPath := filepath.Join(tmpDir, note.name)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory for %s: %v", note.name, err)
		}
		if err := os.WriteFile(fullPath, []byte(note.content), 0644); err != nil {
			t.Fatalf("Failed to create test note %s: %v", note.name, err)
		}
	}

	// Create test server
	r, err := setupServer("")
	if err != nil {
		t.Fatalf("Failed to setup server: %v", err)
	}
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Test welcome page
	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("Failed to get welcome page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	htmlContent := string(body)

	// Check that all notes are listed
	expectedNotes := []string{
		"note1",
		"folder/note2",
		"folder/subfolder/note3",
	}
	for _, note := range expectedNotes {
		if !strings.Contains(htmlContent, fmt.Sprintf("href=\"/notes/%s\"", note)) {
			t.Errorf("Expected welcome page to contain link to %s", note)
		}
		if !strings.Contains(htmlContent, note) {
			t.Errorf("Expected welcome page to contain note name %s", note)
		}
	}
}

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

	// Create test server
	r, err := setupServer(noteName)
	if err != nil {
		t.Fatalf("Failed to setup server: %v", err)
	}
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Test running view command without arguments
	err = runView(viewCmd, []string{})
	if err != nil {
		t.Errorf("Expected no error when running view without arguments, got: %v", err)
	}

	// Test note endpoint
	resp, err := http.Get(ts.URL + "/notes/" + noteName)
	if err != nil {
		t.Fatalf("Failed to get note: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	htmlContent := string(body)

	// Check basic HTML structure
	expectedHTML := "<h1>Test Note</h1>"
	if !strings.Contains(htmlContent, expectedHTML) {
		t.Errorf("HTML content mismatch\nwant to contain: %q\ngot: %q", expectedHTML, htmlContent)
	}

	// Check image paths in HTML
	expectedPaths := []string{
		fmt.Sprintf("/images/%s/test1.jpg", noteName),
		"/images/subfolder/test2.png",
	}
	for _, path := range expectedPaths {
		if !strings.Contains(htmlContent, path) {
			t.Errorf("Expected image path %s not found in HTML", path)
		}
	}

	// Test image endpoints
	imageTests := []struct {
		path     string
		expected []byte
	}{
		{
			path:     fmt.Sprintf("/images/%s/test1.jpg", noteName),
			expected: []byte("fake image 1"),
		},
		{
			path:     "/images/subfolder/test2.png",
			expected: []byte("fake image 2"),
		},
	}

	for _, tt := range imageTests {
		resp, err := http.Get(ts.URL + tt.path)
		if err != nil {
			t.Errorf("Failed to get image %s: %v", tt.path, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status OK for %s; got %v", tt.path, resp.Status)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Failed to read image %s: %v", tt.path, err)
			continue
		}

		if string(body) != string(tt.expected) {
			t.Errorf("Image content mismatch for %s", tt.path)
		}
	}
}

func TestTransformImagePaths(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		notePath string
		expected string
	}{
		{
			name:     "root note with local image",
			content:  "![Alt](test.jpg)",
			notePath: filepath.Join(notesDir, "note.md"),
			expected: "![Alt](/images/test.jpg)",
		},
		{
			name:     "note in folder with local image",
			content:  "![Alt](test.jpg)",
			notePath: filepath.Join(notesDir, "folder1/note.md"),
			expected: "![Alt](/images/folder1/test.jpg)",
		},
		{
			name:     "note in nested folder with local image",
			content:  "![Alt](test.jpg)",
			notePath: filepath.Join(notesDir, "folder1/subfolder/note.md"),
			expected: "![Alt](/images/folder1/subfolder/test.jpg)",
		},
		{
			name:     "absolute path in image",
			content:  "![Alt](folder/test.png)",
			notePath: filepath.Join(notesDir, "folder1/note.md"),
			expected: "![Alt](/images/folder/test.png)",
		},
		{
			name:     "multiple images with mixed paths",
			content:  "![One](test1.jpg)\n![Two](folder/test2.png)",
			notePath: filepath.Join(notesDir, "folder1/note.md"),
			expected: "![One](/images/folder1/test1.jpg)\n![Two](/images/folder/test2.png)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformImagePaths(tt.content, tt.notePath)
			if result != tt.expected {
				t.Errorf("transformImagePaths() = %v, want %v", result, tt.expected)
			}
		})
	}
}
