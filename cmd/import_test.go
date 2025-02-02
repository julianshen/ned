package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestImportCmd(t *testing.T) {
	// Initialize test environment
	testMode = true
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create test file within notes directory
	testImageContent := []byte("fake image content")
	testFilename := "test.jpg"
	testImagePath := filepath.Join(tmpDir, "source", testFilename) // source dir within notes dir
	if err := os.MkdirAll(filepath.Dir(testImagePath), 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	if err := os.WriteFile(testImagePath, testImageContent, 0644); err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	// Get relative path for test file
	relTestImagePath := "source/" + testFilename

	// Create HTTP test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fake remote image content"))
	}))
	defer ts.Close()

	// Log test setup
	t.Logf("Test setup:")
	t.Logf("- Notes directory: %s", notesDir)
	t.Logf("- Source directory: %s", filepath.Join(tmpDir, "source"))
	t.Logf("- Test image: %s", relTestImagePath)
	t.Logf("- Test server: %s", ts.URL)

	tests := []struct {
		name     string
		args     []string
		wantPath string
		wantErr  bool
	}{
		{
			name:     "local image to root",
			args:     []string{relTestImagePath},
			wantPath: filepath.Join(notesDir, "._images_", testFilename),
			wantErr:  false,
		},
		{
			name:     "local image to subdirectory",
			args:     []string{relTestImagePath, "photos"},
			wantPath: filepath.Join(notesDir, "photos", "._images_", testFilename),
			wantErr:  false,
		},
		{
			name:     "remote image to root",
			args:     []string{fmt.Sprintf("%s/image.jpg", ts.URL)},
			wantPath: filepath.Join(notesDir, "._images_", "image.jpg"),
			wantErr:  false,
		},
		{
			name:     "remote image to subdirectory",
			args:     []string{fmt.Sprintf("%s/image.jpg", ts.URL), "photos"},
			wantPath: filepath.Join(notesDir, "photos", "._images_", "image.jpg"),
			wantErr:  false,
		},
		{
			name:    "non-existent file",
			args:    []string{"source/nonexistent.jpg"},
			wantErr: true,
		},
		{
			name:    "source from images directory",
			args:    []string{"folder/._images_/test.jpg"},
			wantErr: true,
		},
		{
			name:    "source from root images directory",
			args:    []string{"._images_/test.jpg"},
			wantErr: true,
		},
		{
			name:    "source from nested images directory",
			args:    []string{"notes/subfolder/._images_/test.jpg"},
			wantErr: true,
		},
		{
			name:    "invalid URL",
			args:    []string{"http://invalid.local/test.jpg"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Test case: %s", tt.name)
			t.Logf("- Arguments: %v", tt.args)
			t.Logf("- Expected path: %s", tt.wantPath)

			err := runImport(importCmd, tt.args)
			t.Logf("- Result: %v", err)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify file existence
			info, err := os.Stat(tt.wantPath)
			if err != nil {
				t.Errorf("Failed to stat imported file: %v", err)
				return
			}

			t.Logf("- Created file: %s (%d bytes)", tt.wantPath, info.Size())

			// Verify content
			content, err := os.ReadFile(tt.wantPath)
			if err != nil {
				t.Errorf("Failed to read imported file: %v", err)
				return
			}

			// Check content based on source
			isURL := len(tt.args) > 0 && (tt.args[0] == fmt.Sprintf("%s/image.jpg", ts.URL))
			expectedContent := testImageContent
			if isURL {
				expectedContent = []byte("fake remote image content")
			}

			if string(content) != string(expectedContent) {
				t.Errorf("Content mismatch\nwant: %q\ngot:  %q", string(expectedContent), string(content))
			}
		})
	}
}
