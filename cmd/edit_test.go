package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditCmd(t *testing.T) {
	// Skip test if running in a CI environment or without a terminal
	if os.Getenv("CI") != "" || os.Getenv("TERM") == "" {
		t.Skip("Skipping test in non-terminal environment")
	}

	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		content string
		verify  func(t *testing.T, path string)
		wantErr bool
	}{
		{
			name: "edit existing file with stdin",
			setup: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "test.md")
				err := os.WriteFile(path, []byte("original content"), 0644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return "test" // Test without .md extension
			},
			content: "new content from stdin\n",
			verify: func(t *testing.T, path string) {
				if !strings.HasSuffix(path, ".md") {
					path += ".md"
				}
				path = filepath.Join(tmpDir, path)
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if string(content) != "new content from stdin\n" {
					t.Errorf("content mismatch\nwant: %q\ngot:  %q", "new content from stdin\n", string(content))
				}
			},
			wantErr: false,
		},
		{
			name: "edit non-existent file",
			setup: func(t *testing.T) string {
				return "nonexistent" // Test without .md extension
			},
			content: "",
			verify:  func(t *testing.T, path string) {},
			wantErr: true,
		},
		{
			name: "edit file in subdirectory",
			setup: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "subdir", "test.md")
				err := os.MkdirAll(filepath.Dir(path), 0755)
				if err != nil {
					t.Fatalf("failed to create directory: %v", err)
				}
				err = os.WriteFile(path, []byte("original content"), 0644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return "subdir/test" // Test without .md extension
			},
			content: "new subdir content\n",
			verify: func(t *testing.T, path string) {
				if !strings.HasSuffix(path, ".md") {
					path += ".md"
				}
				path = filepath.Join(tmpDir, path)
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if string(content) != "new subdir content\n" {
					t.Errorf("content mismatch\nwant: %q\ngot:  %q", "new subdir content\n", string(content))
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := tt.setup(t)

			if tt.content != "" {
				// Create pipe for stdin
				r, w, err := os.Pipe()
				if err != nil {
					t.Fatalf("failed to create pipe: %v", err)
				}
				oldStdin := os.Stdin
				os.Stdin = r
				defer func() { os.Stdin = oldStdin }()

				// Write content to pipe
				errCh := make(chan error, 1)
				go func() {
					_, err := io.WriteString(w, tt.content)
					errCh <- err
					w.Close()
				}()

				if err := <-errCh; err != nil {
					t.Fatalf("failed to write to pipe: %v", err)
				}
			}

			err := runEdit(editCmd, []string{filename})

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
