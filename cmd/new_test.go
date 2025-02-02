package cmd

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestNewCmd(t *testing.T) {
	testMode = true
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name     string
		args     []string
		title    string
		content  string
		wantFile string
		wantErr  bool
	}{
		{
			name:     "create note with filename",
			args:     []string{"test.md"},
			title:    "Test Title",
			content:  "Test content\n",
			wantFile: "test.md",
			wantErr:  false,
		},
		{
			name:     "create note in subdirectory",
			args:     []string{"subdir/test.md"},
			title:    "Sub Test",
			content:  "Subdir content\n",
			wantFile: "subdir/test.md",
			wantErr:  false,
		},
		{
			name:     "create note without md extension",
			args:     []string{"test"},
			title:    "No Extension",
			content:  "Content here\n",
			wantFile: "test.md",
			wantErr:  false,
		},
		{
			name:    "create note in folder with double dot",
			args:    []string{"../test.md"},
			title:   "Invalid",
			content: "Should fail\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up command
			cmd := newCmd
			if tt.title != "" {
				title = tt.title
			} else {
				title = ""
			}

			// Create pipe for stdin if content provided
			oldStdin := os.Stdin
			if tt.content != "" {
				r, w, err := os.Pipe()
				if err != nil {
					t.Fatalf("failed to create pipe: %v", err)
				}
				os.Stdin = r

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
			defer func() { os.Stdin = oldStdin }()

			// Execute command
			err := runNew(cmd, tt.args)

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

			// Check if file was created
			expectedPath := filepath.Join(tmpDir, tt.wantFile)
			assertFileExists(t, expectedPath)

			// Check file content
			content, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("failed to read created file: %v", err)
			}

			// Verify title if specified
			if tt.title != "" {
				expectedContent := "# " + tt.title + "\n\n" + tt.content
				if string(content) != expectedContent {
					t.Errorf("content mismatch\nwant: %q\ngot:  %q", expectedContent, string(content))
				}
			} else {
				if string(content) != tt.content {
					t.Errorf("content mismatch\nwant: %q\ngot:  %q", tt.content, string(content))
				}
			}
		})
	}
}
