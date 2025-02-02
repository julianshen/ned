package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

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

// transformImagePaths modifies markdown image paths to point to the correct ._images_ directory
func transformImagePaths(content string, noteDir string) string {
	// Regular expression to match markdown image syntax: ![alt](path)
	re := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

	return re.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the path from ![alt](path)
		parts := re.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}

		alt := parts[1]
		imgPath := parts[2]

		// If path contains folders, treat it as from root
		if strings.Contains(imgPath, "/") || strings.Contains(imgPath, "\\") {
			// Split into directory and filename
			dir, file := filepath.Split(imgPath)
			// Insert ._images_ before the filename
			newPath := filepath.Join(notesDir, dir, "._images_", file)
			return fmt.Sprintf("![%s](%s)", alt, strings.ReplaceAll(newPath, "\\", "\\\\"))
		}

		// For simple filenames, use note's directory
		newPath := filepath.Join(noteDir, "._images_", imgPath)
		return fmt.Sprintf("![%s](%s)", alt, strings.ReplaceAll(newPath, "\\", "\\\\"))
	})
}

// For testing purposes
var testMode bool
var testFile string

func runView(cobraCmd *cobra.Command, args []string) error {
	noteName := args[0]
	notePath := filepath.Join(notesDir, noteName+".md")

	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		return fmt.Errorf("note '%s' not found", noteName)
	}

	content, err := os.ReadFile(notePath)
	if err != nil {
		return fmt.Errorf("could not read note: %w", err)
	}

	// Get the directory containing the note for relative image paths
	noteDir := filepath.Dir(notePath)

	// Process markdown content
	mdContent := string(content)

	// Transform image paths
	mdContent = transformImagePaths(mdContent, noteDir)
	fmt.Printf(mdContent)

	// Replace Mermaid code blocks with div elements
	mdContent = strings.ReplaceAll(mdContent, "```mermaid", "<div class=\"mermaid\">")
	mdContent = strings.ReplaceAll(mdContent, "```", "</div>")

	htmlContent := markdown.ToHTML([]byte(mdContent), nil, nil)

	// Add HTML wrapper with Mermaid support
	finalHTML := []byte(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
    <script>
        mermaid.initialize({ startOnLoad: true });
    </script>
</head>
<body>
` + string(htmlContent) + `
</body>
</html>`)

	tmpDir := os.TempDir()                               // Get temp dir
	tmpFile, err := os.CreateTemp(tmpDir, "note-*.html") // Create temp file in temp dir
	if err != nil {
		return fmt.Errorf("could not create temp file: %w", err)
	}

	if _, err := tmpFile.Write(finalHTML); err != nil {
		return fmt.Errorf("could not write HTML to temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("could not close temp file: %w", err)
	}

	if !testMode {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "linux":
			cmd = exec.Command("xdg-open", tmpFile.Name())
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", tmpFile.Name())
		case "darwin":
			cmd = exec.Command("open", tmpFile.Name())
		default:
			return fmt.Errorf("unsupported platform")
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to open browser: %w", err)
		}
		fmt.Printf("Viewing note: %s in browser\n", noteName)
		fmt.Print("Press Enter to continue...")
		fmt.Scanln() // Wait for user to press Enter
		defer os.Remove(tmpFile.Name())
	} else {
		testFile = tmpFile.Name()
	}

	return nil
}
