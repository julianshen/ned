package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [filename]",
	Short: "Edit a note",
	Long: `Edit a note using the editor specified in EDITOR environment variable.
The .md extension is optional and will be added automatically if not provided.
If no editor is specified, it defaults to 'vim'. You can also pipe in content
to replace the entire note.`,
	Aliases: []string{"e"},
	Args:    cobra.ExactArgs(1),
	RunE:    runEdit,
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
	filename := args[0]

	// Ensure filename has .md extension
	if !strings.HasSuffix(filename, ".md") {
		filename += ".md"
	}

	if !filepath.IsAbs(filename) {
		filename = filepath.Join(notesDir, filename)
	}

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("note not found: %s", filename)
	}

	// Check if we have content from stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Read content from stdin
		scanner := bufio.NewScanner(os.Stdin)
		var content string
		for scanner.Scan() {
			content += scanner.Text() + "\n"
		}

		// Write content to file
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write content: %w", err)
		}
		return nil
	}

	// Get editor from environment variable
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim" // Default to vim if no editor is specified
	}

	// Create command to open editor
	cmd2 := exec.Command(editor, filename)
	cmd2.Stdin = os.Stdin
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr

	// Run editor
	if err := cmd2.Run(); err != nil {
		return fmt.Errorf("failed to run editor: %w", err)
	}

	return nil
}
