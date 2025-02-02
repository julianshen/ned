package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConvertImagePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "root level image",
			input:    "$root/._images_/test.png",
			expected: "/images/test.png",
		},
		{
			name:     "single folder level",
			input:    "$root/folder1/._images_/test.png",
			expected: "/images/folder1/test.png",
		},
		{
			name:     "nested folders",
			input:    "$root/folder1/folder2/._images_/test.png",
			expected: "/images/folder1/folder2/test.png",
		},
		{
			name:     "no ._images_ in path",
			input:    "$root/folder1/test.png",
			expected: "$root/folder1/test.png",
		},
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertImagePath(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertImagePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestImageCommands(t *testing.T) {
	testMode = true
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create test images in different directories
	testFiles := []struct {
		path    string
		content []byte
	}{
		{
			path:    filepath.Join("._images_", "test1.jpg"),
			content: []byte("fake image 1"),
		},
		{
			path:    filepath.Join("._images_", "test2.png"),
			content: []byte("fake image 2"),
		},
		{
			path:    filepath.Join("._images_", "test3.webp"),
			content: []byte("fake webp image"),
		},
		{
			path:    filepath.Join("subfolder", "._images_", "test4.jpg"),
			content: []byte("fake image 4"),
		},
		{
			path:    filepath.Join("subfolder", "._images_", "test5.gif"),
			content: []byte("fake image 5"),
		},
		{
			path:    filepath.Join("subfolder", "nested", "._images_", "test6.jpg"),
			content: []byte("fake image 6"),
		},
	}

	// Create test images
	for _, tf := range testFiles {
		fullPath := filepath.Join(tmpDir, tf.path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory for %s: %v", tf.path, err)
		}
		if err := os.WriteFile(fullPath, tf.content, 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", tf.path, err)
		}
	}

	// Test image list command
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		contains    []string // Filenames that should be in the output
		notContains []string // Filenames that should not be in the output
	}{
		{
			name:        "list root images",
			args:        []string{},
			wantErr:     false,
			contains:    []string{"test1.jpg", "test2.png", "test3.webp"},
			notContains: []string{"test4.jpg", "test5.gif", "test6.jpg"},
		},
		{
			name:        "list subfolder images",
			args:        []string{"subfolder"},
			wantErr:     false,
			contains:    []string{"test4.jpg", "test5.gif"},
			notContains: []string{"test1.jpg", "test2.png", "test3.webp", "test6.jpg"},
		},
		{
			name:        "list nested subfolder images",
			args:        []string{"subfolder/nested"},
			wantErr:     false,
			contains:    []string{"test6.jpg"},
			notContains: []string{"test1.jpg", "test2.png", "test3.webp", "test4.jpg", "test5.gif"},
		},
		{
			name:    "list with invalid path",
			args:    []string{"../outside"},
			wantErr: true,
		},
		{
			name:    "list with dot path",
			args:    []string{"."},
			wantErr: true,
		},
		{
			name:    "list with non-existent folder",
			args:    []string{"nonexistent"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runImageList(imageListCmd, tt.args)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf strings.Builder
			if _, err := io.Copy(&buf, r); err != nil {
				t.Errorf("Failed to read captured output: %v", err)
			}
			output := buf.String()

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

			// Verify output contains expected filenames
			for _, filename := range tt.contains {
				if !strings.Contains(output, filename) {
					t.Errorf("Expected output to contain %q, but got:\n%s", filename, output)
				}
			}

			// Verify output does not contain excluded filenames
			for _, filename := range tt.notContains {
				if strings.Contains(output, filename) {
					t.Errorf("Expected output to not contain %q, but got:\n%s", filename, output)
				}
			}

			// Verify correct directory is shown in output
			var expectedDirMsg string
			if len(tt.args) == 0 {
				expectedDirMsg = "Images in root ._images_ directory:"
			} else {
				expectedDirMsg = fmt.Sprintf("Images in %s:", filepath.Join(tt.args[0], "._images_"))
			}
			if !strings.Contains(output, expectedDirMsg) {
				t.Errorf("Expected output to contain %q, but got:\n%s", expectedDirMsg, output)
			}
		})
	}

	// Test image show command
	showTests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "show root image",
			args:    []string{"test1.jpg"},
			wantErr: false,
		},
		{
			name:    "show webp image",
			args:    []string{"test3.webp"},
			wantErr: false,
		},
		{
			name:    "show subfolder image",
			args:    []string{"subfolder/test4.jpg"},
			wantErr: false,
		},
		{
			name:    "show with invalid path",
			args:    []string{"../outside/image.jpg"},
			wantErr: true,
		},
		{
			name:    "show non-existent image",
			args:    []string{"._images_/nonexistent.jpg"},
			wantErr: true,
		},
		{
			name:    "show image with explicit ._images_",
			args:    []string{"._images_/test1.jpg"},
			wantErr: true,
		},
		{
			name:    "show with absolute path",
			args:    []string{filepath.Join(tmpDir, "._images_", "test1.jpg")},
			wantErr: true,
		},
	}

	for _, tt := range showTests {
		t.Run(tt.name, func(t *testing.T) {
			if !testMode {
				t.Skip("Skipping image show test in non-test mode")
			}
			err := runImageShow(imageShowCmd, tt.args)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
