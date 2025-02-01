package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ned",
	Short: "A CLI note-taking application",
	Long: `ned is a command line note-taking application that allows you to
create, list, and edit notes in markdown format. All notes are stored in
$HOME/.mynotes directory.`,
	Aliases: []string{"e", "n", "l", "d", "v", "h"},
}

// notesDir is the directory where all notes are stored
var notesDir string

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	notesDir = filepath.Join(homeDir, ".mynotes")

	// Create notes directory if it doesn't exist
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		panic(err)
	}
}
