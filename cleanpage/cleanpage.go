package cleanpage

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/go-shiori/go-readability"
)

// DownloadMethod specifies the method used to download the webpage
type DownloadMethod int

const (
	// HTTPClient uses standard Go http client
	HTTPClient DownloadMethod = iota
	// HeadlessChrome uses Chrome in headless mode
	HeadlessChrome
)

// CrawlPage downloads a webpage and extracts its main content
func CrawlPage(urlStr string, method DownloadMethod) (string, error) {
	var html string
	var err error

	switch method {
	case HTTPClient:
		html, err = downloadWithHTTP(urlStr)
	case HeadlessChrome:
		html, err = downloadWithChrome(urlStr)
	}

	if err != nil {
		return "", err
	}

	// Parse URL string into *url.URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	// Parse the content using go-readability
	article, err := readability.FromReader(strings.NewReader(html), parsedURL)
	if err != nil {
		return "", err
	}

	// Normalize whitespace in the text content
	fields := strings.Fields(article.TextContent)
	return strings.Join(fields, " "), nil
}

func downloadWithHTTP(urlStr string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(urlStr)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// IsChromeInstalled checks if Chrome/Chromium is available for headless operation
func IsChromeInstalled() bool {
	var chromePath string
	switch runtime.GOOS {
	case "windows":
		chromePath = "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"
		if _, err := exec.Command("where", "chrome.exe").Output(); err == nil {
			return true
		}
	case "darwin":
		chromePath = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	case "linux":
		for _, path := range []string{"google-chrome", "chromium", "chromium-browser"} {
			if _, err := exec.LookPath(path); err == nil {
				return true
			}
		}
	}
	_, err := exec.LookPath(chromePath)
	return err == nil
}

func downloadWithChrome(urlStr string) (string, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var html string
	err := chromedp.Run(ctx,
		chromedp.Navigate(urlStr),
		chromedp.OuterHTML("html", &html),
	)
	if err != nil {
		return "", err
	}

	return html, nil
}
