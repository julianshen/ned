package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	title string
)

var newCmd = &cobra.Command{
	Use:   "new [filename]",
	Short: "Create a new note",
	Long: `Create a new note with optional filename and title.
If filename is not provided, an auto-generated name will be used.
The .md extension is optional and will be added automatically if not provided.
You can specify subdirectories in the filename.`,
	Aliases: []string{"n"},
	RunE:    runNew,
}

func init() {
	newCmd.Flags().StringVarP(&title, "title", "t", "", "Title of the note")
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	var filename string
	if len(args) > 0 {
		filename = args[0]
	} else {
		// Generate filename based on current timestamp
		filename = time.Now().Format("2006-01-02-150405") + ".md"
	}

	// Ensure filename has .md extension
	if !strings.HasSuffix(filename, ".md") {
		filename += ".md"
	}

	// Clean and validate the path
	cleanPath := filepath.Clean(filename)
	if filepath.IsAbs(cleanPath) {
		return fmt.Errorf("absolute paths are not allowed")
	}

	// Create full path and verify it's within notes directory
	absNotesDir, err := filepath.Abs(notesDir)
	if err != nil {
		return fmt.Errorf("failed to resolve notes directory path: %w", err)
	}

	fullPath := filepath.Join(notesDir, cleanPath)
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Ensure the target path is within the notes directory
	if !strings.HasPrefix(absPath, absNotesDir) {
		return fmt.Errorf("path must be within notes directory")
	}

	// Create subdirectories if they don't exist
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Check if we have content from stdin
	stat, _ := os.Stdin.Stat()
	var content string
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		var builder strings.Builder
		for scanner.Scan() {
			builder.WriteString(scanner.Text() + "\n")
		}
		content = builder.String()
	}

	// Create file with title if specified
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create note file: %w", err)
	}
	defer f.Close()

	// Set default title if not specified and filename is provided (without extension)
	if title == "" && filename != "" {
		title = strings.TrimSuffix(filepath.Base(filename), ".md")
	}

	// Write title if specified
	if title != "" {
		if _, err := f.WriteString(fmt.Sprintf("# %s\n\n", title)); err != nil {
			return fmt.Errorf("failed to write title: %w", err)
		}
	}

	// Write content from stdin if available
	if content != "" {
		if _, err := f.WriteString(content); err != nil {
			return fmt.Errorf("failed to write content: %w", err)
		}
	}

	fmt.Printf("Created new note: %s\n", filename)

	if !testMode {
		// Prompt user to edit the new note
		var editNote string
		fmt.Print("Do you want to edit the new note? (y/n): ")
		fmt.Scanln(&editNote)

		if strings.ToLower(editNote) == "y" {
			err := runEdit(editCmd, []string{filename})
			if err != nil {
				return fmt.Errorf("failed to edit note: %w", err)
			}
		}
	}

	return nil
}
