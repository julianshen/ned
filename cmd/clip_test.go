package cmd

import (
	"net/http"
	"net/http/httptest"
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

	// Set up test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body>
<article><p>Test content for clipping.</p></article>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}))
	defer ts.Close()

	// Create test config
	config, err := loadConfig()
	assert.NoError(t, err)

	if config.Values["ANTHROPIC_API_KEY"] == "" {
		t.Skip("Skipping test because ANTHROPIC_API_KEY is not set")
		return
	}

	// Save old config key
	oldAPIKey := config.Values["ANTHROPIC_API_KEY"]
	config.Values["ANTHROPIC_API_KEY"] = ""
	err = saveConfig(config)
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
			args: []string{"test", ts.URL},
			validate: func(t *testing.T, notePath string) {
				content, err := os.ReadFile(notePath)
				assert.NoError(t, err)
				assert.Contains(t, string(content), "# test\n")
				assert.Contains(t, string(content), "Source: ["+ts.URL+"]("+ts.URL+")")
			},
		},
		{
			name: "clip with API key",
			args: []string{"test2", ts.URL},
			setup: func() error {
				config.Values["ANTHROPIC_API_KEY"] = oldAPIKey
				return saveConfig(config)
			},
			validate: func(t *testing.T, notePath string) {
				content, err := os.ReadFile(notePath)
				assert.NoError(t, err)
				assert.Contains(t, string(content), "Source: ["+ts.URL+"]("+ts.URL+")")
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
			args:    []string{"test", ts.URL, "extra"},
			wantErr: true,
			errMsg:  "accepts 2 arg(s)",
		},
		{
			name:    "invalid URL",
			args:    []string{"test", "http://invalid.url"},
			wantErr: true,
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
