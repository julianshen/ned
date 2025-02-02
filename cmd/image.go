package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

type EmptyError struct {
	Message string
}

func (e EmptyError) Error() string {
	return e.Message
}

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Manage images in notes",
	Long: `Manage images in the notes system.
List images in a folder or display a specific image.`,
}

var imageListCmd = &cobra.Command{
	Use:   "list [folder]",
	Short: "List images in a folder",
	Long: `List all images in a folder's ._images_ directory.
If no folder is specified, lists images in the root ._images_ directory.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runImageList,
}

var imageShowCmd = &cobra.Command{
	Use:   "show [image]",
	Short: "Show an image",
	Long: `Show an image using the system's default image viewer.
The image path can be just the filename for root images, or folder/filename for images in subfolders.
Examples:
  ned image show image1.jpg           # Shows ._images_/image1.jpg
  ned image show folder1/image2.png   # Shows folder1/._images_/image2.png`,
	Args: cobra.ExactArgs(1),
	RunE: runImageShow,
}

func init() {
	imageCmd.AddCommand(imageListCmd)
	imageCmd.AddCommand(imageShowCmd)
	rootCmd.AddCommand(imageCmd)
}

func runImageList(cmd *cobra.Command, args []string) error {
	// Determine target folder
	folder := ""
	if len(args) > 0 {
		folder = args[0]
	}

	// Get absolute path of notes directory
	absNotesDir, err := filepath.Abs(notesDir)
	if err != nil {
		return fmt.Errorf("failed to resolve notes directory path: %w", err)
	}

	var cleanPath string
	if folder != "" {
		// Clean and validate the folder path
		cleanPath = filepath.Clean(folder)
		if filepath.IsAbs(cleanPath) {
			return fmt.Errorf("absolute paths are not allowed")
		}

		// Verify the folder path is within notes directory
		fullPath := filepath.Join(absNotesDir, cleanPath)
		absPath, err := filepath.Abs(fullPath)
		if err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}
		if !strings.HasPrefix(absPath, absNotesDir) {
			return fmt.Errorf("path must be within notes directory")
		}
	}
	// Check for invalid folder names
	if folder == "." || folder == ".." {
		return fmt.Errorf("invalid folder name: %s", folder)
	}
	// Create full path for the images directory
	imagesDir := filepath.Join(absNotesDir, cleanPath, "._images_")
	// Check if images directory exists
	if _, err := os.Stat(imagesDir); os.IsNotExist(err) {
		return EmptyError{fmt.Sprintf("No images found in %s", filepath.Join(cleanPath, "._images_"))}
	}

	// Check if images directory exists
	if _, err := os.Stat(imagesDir); os.IsNotExist(err) {
		if folder == "" {
			fmt.Println("No images found in root ._images_ directory")
		} else {
			fmt.Printf("No images found in %s\n", filepath.Join(cleanPath, "._images_"))
		}
		return nil
	}

	// Read directory contents
	entries, err := os.ReadDir(imagesDir)
	if err != nil {
		return fmt.Errorf("failed to read images directory: %w", err)
	}

	// Filter and display image files
	var images []string
	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			// Basic image file extension check
			ext := strings.ToLower(filepath.Ext(name))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".bmp" || ext == ".webp" {
				images = append(images, name)
			}
		}
	}

	if len(images) == 0 {
		if folder == "" {
			fmt.Println("No images found in root ._images_ directory")
		} else {
			fmt.Printf("No images found in %s\n", filepath.Join(cleanPath, "._images_"))
		}
		return nil
	}

	// Display images
	if folder == "" {
		fmt.Println("Images in root ._images_ directory:")
	} else {
		fmt.Printf("Images in %s:\n", filepath.Join(cleanPath, "._images_"))
	}
	for _, name := range images {
		fmt.Printf("  %s\n", name)
	}

	return nil
}

func runImageShow(cmd *cobra.Command, args []string) error {
	imagePath := args[0]

	// Clean and validate the path
	cleanPath := filepath.Clean(imagePath)
	if filepath.IsAbs(cleanPath) {
		return fmt.Errorf("absolute paths are not allowed")
	}

	// Get absolute path of notes directory
	absNotesDir, err := filepath.Abs(notesDir)
	if err != nil {
		return fmt.Errorf("failed to resolve notes directory path: %w", err)
	}

	// Split path into directory and filename
	dir, file := filepath.Split(cleanPath)
	if dir == "" {
		// If no directory specified, use root
		dir = "."
	} else {
		// Remove trailing separator if present
		dir = strings.TrimRight(dir, string(filepath.Separator))
	}

	// Create path with ._images_ directory
	imagePath = filepath.Join(dir, "._images_", file)
	fullPath := filepath.Join(absNotesDir, imagePath)
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Ensure the path is within the notes directory
	if !strings.HasPrefix(absPath, absNotesDir) {
		return fmt.Errorf("path must be within notes directory")
	}

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("image not found: %s", imagePath)
	}

	// Open the image with the system's default viewer
	var cmd2 *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd2 = exec.Command("cmd", "/C", "start", "", fullPath)
	case "darwin":
		cmd2 = exec.Command("open", fullPath)
	default: // Linux and others
		cmd2 = exec.Command("xdg-open", fullPath)
	}

	if err := cmd2.Run(); err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}

	return nil
}
