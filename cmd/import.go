package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import [image_source] [folder]",
	Short: "Import an image from file or URL",
	Long: `Import an image from a local file or URL into the notes system.
The image will be stored in the ._images_ subdirectory under the specified folder
or root folder if no folder is specified.

Examples:
  ned import image.jpg            # Import local image to root ._images_ folder
  ned import http://example.com/image.jpg  # Import remote image to root ._images_ folder
  ned import image.jpg projects   # Import local image to projects/._images_ folder`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runImport,
}

func init() {
	rootCmd.AddCommand(importCmd)
}

func runImport(cmd *cobra.Command, args []string) error {
	source := args[0]

	// Determine target folder
	targetFolder := ""
	if len(args) > 1 {
		targetFolder = args[1]
	}

	// Clean and validate the folder path
	cleanPath := filepath.Clean(targetFolder)
	if filepath.IsAbs(cleanPath) {
		return fmt.Errorf("absolute paths are not allowed")
	}

	// Convert to absolute paths for consistent handling
	absNotesDir, err := filepath.Abs(notesDir)
	if err != nil {
		return fmt.Errorf("failed to resolve notes directory path: %w", err)
	}

	// Create full path for the images directory
	imagesDir := filepath.Join(absNotesDir, cleanPath, "._images_")

	// Create the images directory if it doesn't exist
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create images directory: %w", err)
	}

	// Get the filename from source
	filename := filepath.Base(source)
	if filename == "" || filename == "." {
		return fmt.Errorf("invalid source filename")
	}

	// Full path for the target image
	targetPath := filepath.Join(imagesDir, filename)

	// Check if it's a URL or local file
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		// Download from URL
		resp, err := http.Get(source)
		if err != nil {
			return fmt.Errorf("failed to download image: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to download image: HTTP status %d", resp.StatusCode)
		}

		// Create the target file
		out, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("failed to create image file: %w", err)
		}
		defer out.Close()

		// Copy the content
		if _, err := io.Copy(out, resp.Body); err != nil {
			return fmt.Errorf("failed to save image: %w", err)
		}
	} else {
		// Handle local file
		// Clean and validate the source path
		cleanSource := filepath.Clean(source)
		if filepath.IsAbs(cleanSource) {
			return fmt.Errorf("absolute paths are not allowed for source file")
		}

		// Check for . and .. in source path components
		sourceParts := strings.Split(filepath.Dir(cleanSource), string(filepath.Separator))
		for _, part := range sourceParts {
			if part == "." || part == ".." {
				return fmt.Errorf("source path cannot contain '.' or '..'")
			}
			if part == "._images_" {
				return fmt.Errorf("cannot import from ._images_ directory")
			}
		}

		// Create full source path within notes directory
		fullSourcePath := filepath.Join(absNotesDir, cleanSource)
		absSourcePath, err := filepath.Abs(fullSourcePath)
		if err != nil {
			return fmt.Errorf("invalid source path: %w", err)
		}

		// Ensure the source path is within the notes directory and not in any ._images_ directory
		if !strings.HasPrefix(absSourcePath, absNotesDir) {
			return fmt.Errorf("source path must be within notes directory")
		}
		if strings.Contains(absSourcePath, string(filepath.Separator)+"._images_"+string(filepath.Separator)) {
			return fmt.Errorf("cannot import from ._images_ directory")
		}

		// Read the source file
		in, err := os.Open(fullSourcePath)
		if err != nil {
			return fmt.Errorf("failed to open source image: %w", err)
		}
		defer in.Close()

		// Create the target file
		out, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("failed to create image file: %w", err)
		}
		defer out.Close()

		// Copy the content
		if _, err := io.Copy(out, in); err != nil {
			return fmt.Errorf("failed to copy image: %w", err)
		}
	}

	// Generate relative path for output
	relPath := filepath.Join(targetFolder, "._images_", filename)
	fmt.Printf("Imported image to: %s\n", relPath)
	return nil
}
