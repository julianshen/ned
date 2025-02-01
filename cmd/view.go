package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/gomarkdown/markdown"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:     "view [note name]",
	Short:   "View a note in the browser",
	Long:    `Renders a note from markdown to HTML and opens it in the default browser.`,
	Aliases: []string{"v"},
	Args:    cobra.ExactArgs(1),
	RunE:    runView,
}

func init() {
	rootCmd.AddCommand(viewCmd)
}

// For testing purposes
var testMode bool
var testFile string

func runView(cmd *cobra.Command, args []string) error {
	noteName := args[0]
	notePath := filepath.Join(notesDir, noteName+".md")

	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		return fmt.Errorf("note '%s' not found", noteName)
	}

	content, err := os.ReadFile(notePath)
	if err != nil {
		return fmt.Errorf("could not read note: %w", err)
	}

	htmlContent := markdown.ToHTML(content, nil, nil)

	tmpDir := os.TempDir()                               // Get temp dir
	tmpFile, err := os.CreateTemp(tmpDir, "note-*.html") // Create temp file in temp dir
	if err != nil {
		return fmt.Errorf("could not create temp file: %w", err)
	}

	if _, err := tmpFile.Write(htmlContent); err != nil {
		return fmt.Errorf("could not write HTML to temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("could not close temp file: %w", err)
	}

	var browserCmd string
	switch runtime.GOOS {
	case "linux":
		browserCmd = "xdg-open"
	case "windows":
		browserCmd = "cmd /c start"
	case "darwin":
		browserCmd = "open"
	default:
		return fmt.Errorf("unsupported platform")
	}

	browserPath, err := exec.LookPath(browserCmd)
	if err != nil {
		return fmt.Errorf("browser command not found: %w", err)
	}

	if !testMode {
		cmdBrowser := exec.Command(browserPath, tmpFile.Name())
		cmdBrowser.Stdout = os.Stdout
		cmdBrowser.Stderr = os.Stderr
		if err := cmdBrowser.Run(); err != nil {
			return fmt.Errorf("failed to open browser: %w", err)
		}
	} else {
		testFile = tmpFile.Name()
	}

	fmt.Printf("Viewing note: %s in browser\n", noteName)
	return nil
}
