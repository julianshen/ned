package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ned/ainote"
	"ned/cleanpage"

	"github.com/spf13/cobra"
)

var clipCmd = &cobra.Command{
	Use:   "clip [note] [url]",
	Short: "Clip a webpage to a note",
	Long: `Clip a webpage to a note. If ANTHROPIC_API_KEY is set in config, 
it will download and summarize the content. Otherwise, it will just save the URL.

Example:
  ned clip mynote https://example.com`,
	Args: cobra.ExactArgs(2),
	RunE: runClip,
}

func init() {
	rootCmd.AddCommand(clipCmd)
}

func runClip(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("accepts 2 arg(s), received %d", len(args))
	}
	noteName := args[0]
	url := args[1]

	// Add .md extension if not present
	if !strings.HasSuffix(noteName, ".md") {
		noteName += ".md"
	}

	// Get absolute path of notes directory
	absNotesDir, err := filepath.Abs(notesDir)
	if err != nil {
		return fmt.Errorf("failed to resolve notes directory path: %w", err)
	}

	// Create full path for the note
	notePath := filepath.Join(absNotesDir, noteName)

	// Check if the note's directory exists, create if not
	noteDir := filepath.Dir(notePath)
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Load config to check for API key
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Clipping content to note...")
	var content string
	if apiKey, exists := config.Values["ANTHROPIC_API_KEY"]; exists && apiKey != "" {
		// Download and clean the webpage content
		pageContent, err := cleanpage.CrawlPage(url, cleanpage.HTTPClient)
		if err != nil {
			return fmt.Errorf("failed to download webpage: %w", err)
		}

		// Create AI note instance
		ai, err := ainote.NewAINote("", apiKey)
		if err != nil {
			return fmt.Errorf("failed to create AI note: %w", err)
		}

		// Summarize the content
		summary, err := ai.SummarizeArticle(pageContent)
		if err != nil {
			return fmt.Errorf("failed to summarize article: %w", err)
		}

		content = summary
	} else {
		content = "# " + strings.TrimSuffix(noteName, ".md") + "\n"
	}

	fmt.Println(content)
	// Append the URL at the end
	content += "\n\nSource: [" + url + "](" + url + ")\n"

	// Write to file
	if err := os.WriteFile(notePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write note: %w", err)
	}

	fmt.Printf("Created note: %s\n", noteName)
	return nil
}
