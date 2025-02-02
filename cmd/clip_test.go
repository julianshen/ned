package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClipCmd(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	originalNotesDir := notesDir
	notesDir = tmpDir
	defer func() {
		notesDir = originalNotesDir
	}()

	// Create test config
	config := &Config{
		Values: map[string]string{},
	}
	err := saveConfig(config)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		args     []string
		setup    func() error
		validate func(t *testing.T, notePath string)
		wantErr  bool
		errMsg   string
	}{
		{
			name: "clip without API key",
			args: []string{"test", "https://example.com"},
			validate: func(t *testing.T, notePath string) {
				content, err := os.ReadFile(notePath)
				assert.NoError(t, err)
				assert.Contains(t, string(content), "# test\n")
				assert.Contains(t, string(content), "Source: https://example.com")
			},
		},
		{
			name: "clip with API key",
			args: []string{"test2", "https://example.com"},
			setup: func() error {
				config.Values["ANTHROPIC_API_KEY"] = "test-key"
				return saveConfig(config)
			},
			validate: func(t *testing.T, notePath string) {
				content, err := os.ReadFile(notePath)
				assert.NoError(t, err)
				assert.Contains(t, string(content), "Source: https://example.com")
			},
		},
		{
			name:    "missing URL",
			args:    []string{"test"},
			wantErr: true,
			errMsg:  "accepts 2 arg(s)",
		},
		{
			name:    "too many arguments",
			args:    []string{"test", "https://example.com", "extra"},
			wantErr: true,
			errMsg:  "accepts 2 arg(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				err := tt.setup()
				assert.NoError(t, err)
			}

			err := clipCmd.RunE(clipCmd, tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				noteName := tt.args[0]
				if !strings.HasSuffix(noteName, ".md") {
					noteName += ".md"
				}
				notePath := filepath.Join(tmpDir, noteName)
				tt.validate(t, notePath)
			}
		})
	}
}
