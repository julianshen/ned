package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all notes in tree structure",
	Long:    `List all notes in a tree structure showing directories and files`,
	Aliases: []string{"l"},
	RunE:    runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	// Check if directory has any .md files
	entries, err := os.ReadDir(notesDir)
	if err != nil {
		return err
	}

	hasMdFiles := false
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
			hasMdFiles = true
			break
		}
	}

	if !hasMdFiles {
		fmt.Println("Empty")
		return nil
	}

	fmt.Println("Notes structure:")
	return filepath.Walk(notesDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself and non-.md files
		if path == notesDir {
			return nil
		}

		// Only process directories and .md files
		if !info.IsDir() && filepath.Ext(path) != ".md" {
			return nil
		}

		// Get relative path from notes directory
		relPath, err := filepath.Rel(notesDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Calculate depth for indentation
		depth := len(strings.Split(relPath, string(os.PathSeparator))) - 1
		indent := strings.Repeat("  ", depth)

		// Determine if the current entry is the last visible one in its directory
		isLast := false
		parentDir := filepath.Dir(path)
		entries, err := os.ReadDir(parentDir)
		if err == nil {
			// Find the last visible entry (directory or .md file)
			var lastVisibleEntry string
			for i := len(entries) - 1; i >= 0; i-- {
				entry := entries[i]
				if entry.IsDir() || filepath.Ext(entry.Name()) == ".md" {
					lastVisibleEntry = entry.Name()
					break
				}
			}
			if lastVisibleEntry == filepath.Base(path) {
				isLast = true
			}
		}

		// Add different prefix for files and directories
		prefix := "├──"
		if isLast {
			prefix = "└──"
		}
		name := filepath.Base(path)
		if filepath.Ext(name) == ".md" {
			name = strings.TrimSuffix(name, ".md")
		}

		fmt.Printf("%s%s %s\n", indent, prefix, name)
		return nil
	})
}
