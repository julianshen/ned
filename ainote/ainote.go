package ainote

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const claudeAPIEndpoint = "https://api.anthropic.com/v1/messages"

// ErrEmptyContent is returned when the provided content is empty
var ErrEmptyContent = fmt.Errorf("content cannot be empty")

// AINote represents a note-taking service using Claude AI
type AINote struct {
	apiKey string
	model  string
}

// NewAINote creates a new AINote instance with the specified model.
// If apiKeyPath is empty, it defaults to "API_KEY" file.
func NewAINote(model string, apiKey string) (*AINote, error) {
	if model == "" {
		model = "claude-3-opus-20240229"
	}

	if apiKey == "" {
		return nil, fmt.Errorf("no API key provided")
	}

	return &AINote{
		apiKey: strings.TrimSpace(string(apiKey)),
		model:  model,
	}, nil
}

// SummarizeArticle takes an article content and returns a markdown-formatted summary.
// It uses Claude AI to analyze the content and generate a structured markdown note.
// The summary will include a title and relevant sections based on the content.
func (a *AINote) SummarizeArticle(content string) (string, error) {
	// Validate input
	if strings.TrimSpace(content) == "" {
		return "", ErrEmptyContent
	}

	// Create HTTP client
	client := &http.Client{}

	// Format the prompt
	prompt := fmt.Sprintf("<document>%s</document>take a note of this document with markdown output, start from title", content)

	// Create request body
	requestBody := map[string]interface{}{
		"model": a.model,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": prompt,
					},
				},
			},
		},
		"max_tokens": 4096,
	}

	// Marshal request body to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", claudeAPIEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Return the markdown summary
	if len(response.Content) > 0 && len(response.Content[0].Text) > 0 {
		return response.Content[0].Text, nil
	}
	return "", fmt.Errorf("no content in response")
}
