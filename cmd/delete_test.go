package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestDeleteCmd(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name       string
		setup      func(t *testing.T) string
		withForce  bool
		withSilent bool
		input      string
		wantErr    bool
		verify     func(t *testing.T, path string)
	}{
		{
			name: "delete file",
			setup: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "test.md")
				err := os.WriteFile(path, []byte("test content"), 0644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return "test" // Test without .md extension
			},
			withSilent: true,
			wantErr:    false,
			verify: func(t *testing.T, path string) {
				fullPath := filepath.Join(tmpDir, path+".md")
				if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
					t.Errorf("file still exists: %s", path)
				}
			},
		},
		{
			name: "delete empty directory",
			setup: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "testdir")
				err := os.MkdirAll(path, 0755)
				if err != nil {
					t.Fatalf("failed to create directory: %v", err)
				}
				return "testdir"
			},
			withSilent: true,
			wantErr:    false,
			verify: func(t *testing.T, path string) {
				fullPath := filepath.Join(tmpDir, path)
				if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
					t.Errorf("directory still exists: %s", path)
				}
			},
		},
		{
			name: "delete non-empty directory without force",
			setup: func(t *testing.T) string {
				dir := filepath.Join(tmpDir, "nonempty")
				err := os.MkdirAll(dir, 0755)
				if err != nil {
					t.Fatalf("failed to create directory: %v", err)
				}
				err = os.WriteFile(filepath.Join(dir, "file.md"), []byte("content"), 0644)
				if err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				return "nonempty"
			},
			withSilent: true,
			wantErr:    true,
			verify: func(t *testing.T, path string) {
				fullPath := filepath.Join(tmpDir, path)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Errorf("directory was deleted when it shouldn't be: %s", path)
				}
			},
		},
		{
			name: "delete non-empty directory with force",
			setup: func(t *testing.T) string {
				dir := filepath.Join(tmpDir, "forcedir")
				err := os.MkdirAll(dir, 0755)
				if err != nil {
					t.Fatalf("failed to create directory: %v", err)
				}
				err = os.WriteFile(filepath.Join(dir, "file.md"), []byte("content"), 0644)
				if err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				return "forcedir"
			},
			withForce:  true,
			withSilent: true,
			wantErr:    false,
			verify: func(t *testing.T, path string) {
				fullPath := filepath.Join(tmpDir, path)
				if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
					t.Errorf("directory still exists: %s", path)
				}
			},
		},
		{
			name: "delete with user confirmation - yes",
			setup: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "confirm.md")
				err := os.WriteFile(path, []byte("test content"), 0644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return "confirm"
			},
			input:   "y\n",
			wantErr: false,
			verify: func(t *testing.T, path string) {
				fullPath := filepath.Join(tmpDir, path+".md")
				if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
					t.Errorf("file still exists: %s", path)
				}
			},
		},
		{
			name: "delete with user confirmation - no",
			setup: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "keep.md")
				err := os.WriteFile(path, []byte("test content"), 0644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return "keep"
			},
			input:   "n\n",
			wantErr: false,
			verify: func(t *testing.T, path string) {
				fullPath := filepath.Join(tmpDir, path+".md")
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Errorf("file was deleted when it shouldn't be: %s", path)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags
			force = tt.withForce
			silent = tt.withSilent

			filename := tt.setup(t)

			// Mock stdin if input provided
			var stdin bytes.Buffer
			if tt.input != "" {
				stdin.WriteString(tt.input)
				oldStdin := os.Stdin
				r, w, _ := os.Pipe()
				os.Stdin = r
				go func() {
					io.WriteString(w, tt.input)
					w.Close()
				}()
				defer func() { os.Stdin = oldStdin }()
			}

			err := runDelete(deleteCmd, []string{filename})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			tt.verify(t, filename)
		})
	}
}
