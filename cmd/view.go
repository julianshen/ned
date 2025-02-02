package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"bytes"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
)

var viewCmd = &cobra.Command{
	Use:   "view [note name]",
	Short: "View a note or the welcome page in the browser",
	Long: `Renders a note from markdown to HTML and opens it in the default browser.
If no note name is provided, opens the welcome page showing all available notes.`,
	Aliases: []string{"v"},
	Args:    cobra.MaximumNArgs(1),
	RunE:    runView,
}

func init() {
	rootCmd.AddCommand(viewCmd)
}

// For testing purposes
var testMode bool

const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
    <script>
        mermaid.initialize({ startOnLoad: true });
    </script>
    <style>
        body {
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            line-height: 1.6;
        }
        img {
            max-width: 100%%;
            height: auto;
        }
    </style>
</head>
<body>
%s
</body>
</html>`

func transformImagePaths(content string, notePath string) string {
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

		// Convert image path to standardized URL path
		if strings.Contains(imgPath, "/") || strings.Contains(imgPath, "\\") {
			// If path contains folders, keep the structure
			dir, file := filepath.Split(imgPath)
			// Remove trailing slash and clean the directory path
			dir = strings.TrimRight(dir, "/\\")
			if dir == "" {
				return fmt.Sprintf("![%s](/images/%s)", alt, file)
			}
			return fmt.Sprintf("![%s](/images/%s/%s)", alt, dir, file)
		}

		// For simple filenames, use the note's parent folder
		noteDir := filepath.Dir(notePath)
		// Get relative path from notes directory
		relNoteDir, err := filepath.Rel(notesDir, noteDir)
		if err != nil || relNoteDir == "." {
			return fmt.Sprintf("![%s](/images/%s)", alt, imgPath)
		}
		return fmt.Sprintf("![%s](/images/%s/%s)", alt, relNoteDir, imgPath)
	})
}

func setupServer(noteName string) (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Serve welcome page at root
	r.GET("/", func(c *gin.Context) {
		// Get all markdown files in notes directory
		var notes []string
		err := filepath.Walk(notesDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
				// Get relative path from notes directory
				relPath, err := filepath.Rel(notesDir, path)
				if err != nil {
					return err
				}
				// Remove .md extension
				noteName := strings.TrimSuffix(relPath, ".md")
				notes = append(notes, noteName)
			}
			return nil
		})
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to list notes")
			return
		}

		// Sort notes alphabetically
		sort.Strings(notes)

		// Create welcome page HTML
		welcomeHTML := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Notes</title>
    <style>
        body {
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            line-height: 1.6;
        }
        h1 {
            border-bottom: 2px solid #eee;
            padding-bottom: 10px;
        }
        ul {
            list-style-type: none;
            padding: 0;
        }
        li {
            margin: 10px 0;
            padding: 10px;
            background: #f5f5f5;
            border-radius: 4px;
        }
        a {
            color: #0366d6;
            text-decoration: none;
        }
        a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <h1>Notes</h1>
    <ul>
`
		for _, note := range notes {
			welcomeHTML += fmt.Sprintf("        <li><a href=\"/notes/%s\">%s</a></li>\n", note, note)
		}
		welcomeHTML += `    </ul>
</body>
</html>`

		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, welcomeHTML)
	})

	// Serve notes
	r.GET("/notes/*path", func(c *gin.Context) {
		path := c.Param("path")
		notePath := filepath.Join(notesDir, path+".md")

		content, err := os.ReadFile(notePath)
		if err != nil {
			c.String(http.StatusNotFound, "Note not found")
			return
		}

		// Transform content
		mdContent := transformImagePaths(string(content), notePath)

		// Replace Mermaid code blocks
		var inMermaid bool
		lines := strings.Split(mdContent, "\n")
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "```mermaid" {
				lines[i] = "<div class=\"mermaid\">"
				inMermaid = true
			} else if inMermaid && trimmed == "```" {
				lines[i] = "</div>"
				inMermaid = false
			}
		}
		mdContent = strings.Join(lines, "\n")

		// Convert to HTML using goldmark
		var buf bytes.Buffer
		if err := goldmark.Convert([]byte(mdContent), &buf); err != nil {
			c.String(http.StatusInternalServerError, "Failed to convert markdown")
			return
		}
		finalHTML := fmt.Sprintf(htmlTemplate, buf.String())

		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, finalHTML)
	})

	// Serve images from ._images_ directories under the /images path
	r.GET("/images/*path", func(c *gin.Context) {
		// Get the requested image path
		imgPath := c.Param("path")
		if imgPath == "" {
			c.String(http.StatusNotFound, "Image not found")
			return
		}
		// Remove leading slash
		imgPath = strings.TrimPrefix(imgPath, "/")

		// Split into directory and filename
		dir, file := filepath.Split(imgPath)
		dir = strings.TrimRight(dir, "/\\")

		// Construct the physical path
		var physicalPath string
		if dir == "" {
			// Root level image
			physicalPath = filepath.Join(filepath.Dir(filepath.Join(notesDir, noteName+".md")), "._images_", file)
		} else {
			// Nested image
			physicalPath = filepath.Join(notesDir, dir, "._images_", file)
		}

		// Check if file exists
		if _, err := os.Stat(physicalPath); os.IsNotExist(err) {
			c.String(http.StatusNotFound, "Image not found")
			return
		}

		// Serve the file
		c.File(physicalPath)
	})

	return r, nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Run()
}

func runView(cmd *cobra.Command, args []string) error {
	var noteName string
	if len(args) > 0 {
		noteName = args[0]
		// Check if note exists when a specific note is requested
		notePath := filepath.Join(notesDir, noteName+".md")
		if _, err := os.Stat(notePath); os.IsNotExist(err) {
			return fmt.Errorf("note '%s' not found", noteName)
		}
	}

	// Setup and start server
	r, err := setupServer(noteName)
	if err != nil {
		return fmt.Errorf("failed to setup server: %w", err)
	}

	if !testMode {
		// Start server in a goroutine
		go func() {
			if err := r.Run(":3000"); err != nil {
				fmt.Printf("Server error: %v\n", err)
			}
		}()

		// Open browser to either welcome page or specific note
		url := "http://localhost:3000"
		if noteName != "" {
			url = fmt.Sprintf("http://localhost:3000/notes/%s", noteName)
		}
		if err := openBrowser(url); err != nil {
			return fmt.Errorf("failed to open browser: %w", err)
		}

		fmt.Printf("Viewing note: %s\nPress Enter to stop the server...\n", noteName)
		fmt.Scanln()
	} else {
		// In test mode, just return without starting the server
		return nil
	}

	return nil
}
