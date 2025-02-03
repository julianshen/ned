package cleanpage

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-shiori/go-readability"
)

// CrawlPage downloads a webpage and extracts its main content
func CrawlPage(urlStr string) (string, error) {
	html, err := downloadWithHTTP(urlStr)
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
