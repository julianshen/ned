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
	// Check if directory is empty
	entries, err := os.ReadDir(notesDir)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		fmt.Println("Empty")
		return nil
	}

	fmt.Println("Notes structure:")
	return filepath.Walk(notesDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == notesDir {
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

		// Determine if the current entry is the last one in its directory
		isLast := false
		parentDir := filepath.Dir(path)
		entries, err := os.ReadDir(parentDir)
		if err == nil {
			if len(entries) > 0 && entries[len(entries)-1].Name() == filepath.Base(path) {
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
