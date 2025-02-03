package cleanpage

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/go-shiori/go-readability"
)

// CrawlPage downloads a webpage and extracts its main content
func CrawlPage(urlStr string) (string, error) {
	// Try Chrome first
	html, err := downloadWithChrome(urlStr)
	if err != nil {
		// Fall back to HTTP client
		html, err = downloadWithHTTP(urlStr)
		if err != nil {
			return "", err
		}
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

func downloadWithChrome(urlStr string) (string, error) {
	// Create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Create a timeout
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var html string
	err := chromedp.Run(ctx,
		// Navigate to the page
		chromedp.Navigate(urlStr),
		// Wait for the page to load
		chromedp.WaitReady("body"),
		// Extract the HTML
		chromedp.OuterHTML("html", &html),
	)

	if err != nil {
		return "", err
	}

	return html, nil
}
