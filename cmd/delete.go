package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	force  bool
	silent bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete [filename]",
	Short: "Delete a note or empty directory",
	Long: `Delete a note or directory. The .md extension is optional for note files.
Empty directories can be deleted normally. Use --force to delete non-empty directories.
Use --silent to skip confirmation prompt.`,
	Aliases: []string{"d"},
	Args:    cobra.ExactArgs(1),
	RunE:    runDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&force, "force", "f", false, "Force delete non-empty directories")
	deleteCmd.Flags().BoolVarP(&silent, "silent", "s", false, "Delete without confirmation")
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	path := args[0]

	// Create full path
	fullPath := filepath.Join(notesDir, path)

	var info os.FileInfo
	var err error

	// Check if path exists
	checkInfo, err := os.Stat(fullPath)
	if err == nil && !checkInfo.IsDir() && !strings.HasSuffix(path, ".md") {
		// If it's a file and doesn't have .md extension, try with .md
		fullPath = fullPath + ".md"
	}

	// Try both with and without .md extension for files
	pathToShow := path
	info, err = os.Stat(fullPath)
	if os.IsNotExist(err) && !strings.HasSuffix(fullPath, ".md") {
		// Try with .md extension
		info, err = os.Stat(fullPath + ".md")
		if err == nil {
			fullPath = fullPath + ".md"
			pathToShow = path + ".md"
		}
	}

	if os.IsNotExist(err) {
		return fmt.Errorf("note or directory not found: %s", path)
	} else if err != nil {
		return fmt.Errorf("error accessing path: %w", err)
	}

	// Handle directory deletion
	if info.IsDir() {
		isEmpty, dirErr := isDirEmpty(fullPath)
		if dirErr != nil {
			return fmt.Errorf("error checking directory: %w", dirErr)
		}

		if !isEmpty && !force {
			return fmt.Errorf("directory not empty, use --force to delete recursively")
		}
	}

	// Confirm deletion unless silent flag is set
	if !silent {
		fmt.Printf("Are you sure you want to delete '%s'? [y/N]: ", path)
		reader := bufio.NewReader(os.Stdin)
		response, confirmErr := reader.ReadString('\n')
		if confirmErr != nil {
			return fmt.Errorf("error reading response: %w", confirmErr)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	// Perform deletion
	if info.IsDir() && force {
		err = os.RemoveAll(fullPath)
	} else {
		err = os.Remove(fullPath)
	}

	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	fmt.Printf("Deleted: %s\n", pathToShow)
	return nil
}

// isDirEmpty returns true if the directory is empty
func isDirEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}
