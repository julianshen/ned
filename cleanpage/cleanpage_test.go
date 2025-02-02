package cleanpage

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsChromeInstalled(t *testing.T) {
	// Test the Chrome installation check
	result := IsChromeInstalled()
	t.Logf("Chrome installation status: %v", result)
}

func TestCrawlPage(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Test Page</title>
		</head>
		<body>
			<header>
				<nav>Menu items</nav>
			</header>
			<main>
				<article>
					<h1>Main Article</h1>
					<p>This is the main content that should be extracted.</p>
					<p>Additional important content here.</p>
				</article>
			</main>
			<footer>
				<p>Footer content that should be ignored</p>
			</footer>
		</body>
		</html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}))
	defer ts.Close()

	tests := []struct {
		name    string
		url     string
		method  DownloadMethod
		want    string
		wantErr bool
	}{
		{
			name:    "HTTP Client Download",
			url:     ts.URL,
			method:  HTTPClient,
			want:    "Main Article This is the main content that should be extracted. Additional important content here.",
			wantErr: false,
		},
		{
			name:    "Invalid URL",
			url:     "http://invalid.url.that.does.not.exist",
			method:  HTTPClient,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CrawlPage(tt.url, tt.method)
			if (err != nil) != tt.wantErr {
				t.Errorf("CrawlPage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !strings.Contains(strings.ReplaceAll(got, "\n", " "), tt.want) {
				t.Errorf("CrawlPage() = %v, want %v", strings.ReplaceAll(got, "\n", " "), tt.want)
			}
		})
	}
}
