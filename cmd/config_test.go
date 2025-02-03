package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigSetCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		errMsg   string
		simulate func() // For simulating user input
	}{
		{
			name:    "set new key",
			args:    []string{"test_key", "test_value"},
			wantErr: false,
		},
		{
			name:    "missing value",
			args:    []string{"test_key"},
			wantErr: true,
			errMsg:  "accepts 2 arg(s)",
		},
		{
			name:    "missing key and value",
			args:    []string{},
			wantErr: true,
			errMsg:  "accepts 2 arg(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.simulate != nil {
				tt.simulate()
			}

			err := configSetCmd.RunE(configSetCmd, tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				// Verify the value was set
				config, err := loadConfig()
				assert.NoError(t, err)
				if len(tt.args) == 2 {
					assert.Equal(t, tt.args[1], config.Values[tt.args[0]])
				}
			}
		})
	}
}

func TestConfigShowCmd(t *testing.T) {
	// Create a temporary test config file before each test
	oldConfig, err := loadConfig()
	assert.NoError(t, err)
	tmpConfig := &Config{}
	err = saveConfig(tmpConfig)
	assert.NoError(t, err)
	defer saveConfig(oldConfig)

	tests := []struct {
		name     string
		setup    func() error
		expected string
	}{
		{
			name: "show empty config",
			setup: func() error {
				return nil
			},
			expected: "No configuration values set\n",
		},
		{
			name: "show config with values",
			setup: func() error {
				config := &Config{
					Values: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				}
				return saveConfig(config)
			},
			expected: "key1: value1\nkey2: value2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup()
			assert.NoError(t, err)

			// Capture stdout
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			assert.NoError(t, err)
			os.Stdout = w

			err = configShowCmd.RunE(configShowCmd, []string{})
			assert.NoError(t, err)
			w.Close()

			// Read captured output
			var buf bytes.Buffer
			_, err = io.Copy(&buf, r)
			assert.NoError(t, err)
			os.Stdout = oldStdout

			// Sort the lines for consistent comparison
			actualLines := strings.Split(buf.String(), "\n")
			expectedLines := strings.Split(tt.expected, "\n")
			if len(actualLines) > 0 && actualLines[len(actualLines)-1] == "" {
				actualLines = actualLines[:len(actualLines)-1]
			}
			if len(expectedLines) > 0 && expectedLines[len(expectedLines)-1] == "" {
				expectedLines = expectedLines[:len(expectedLines)-1]
			}

			// Sort both slices
			if len(actualLines) > 1 {
				sort.Strings(actualLines)
			}
			if len(expectedLines) > 1 {
				sort.Strings(expectedLines)
			}

			assert.Equal(t, expectedLines, actualLines)
		})
	}
}

func TestConfigHelpers(t *testing.T) {
	// Create a temporary test config file before each test
	oldConfig, err := loadConfig()
	assert.NoError(t, err)
	tmpConfig := &Config{}
	err = saveConfig(tmpConfig)
	assert.NoError(t, err)
	defer saveConfig(oldConfig)

	// Test config path creation
	configPath, err := getConfigPath()
	assert.NoError(t, err)
	home, err := os.UserHomeDir()
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".config", "ned", "config.toml"), configPath)

	// Test initial config load
	config, err := loadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Empty(t, config.Values)

	// Test saving and loading config
	testConfig := &Config{
		Values: map[string]string{
			"test_key": "test_value",
		},
	}
	err = saveConfig(testConfig)
	assert.NoError(t, err)

	loadedConfig, err := loadConfig()
	assert.NoError(t, err)
	assert.Equal(t, testConfig.Values, loadedConfig.Values)
}
