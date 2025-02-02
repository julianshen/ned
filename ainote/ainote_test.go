package ainote

import (
	"os"
	"strings"
	"testing"
)

func TestNewAINote(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: ANTHROPIC_API_KEY environment variable not set")
	}

	tests := []struct {
		name      string
		model     string
		key       string
		wantError bool
	}{
		{
			name:      "Default values",
			model:     "",
			key:       apiKey,
			wantError: false,
		},
		{
			name:      "Custom model",
			model:     "claude-3-opus-20240229",
			key:       apiKey,
			wantError: false,
		},
		{
			name:      "Invalid key",
			model:     "",
			key:       "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			note, err := NewAINote(tt.model, tt.key)
			if (err != nil) != tt.wantError {
				t.Errorf("NewAINote() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && note == nil {
				t.Error("NewAINote() returned nil but no error")
			}
		})
	}
}

func TestSummarizeArticle(t *testing.T) {
	// Skip if no API key is provided in environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: ANTHROPIC_API_KEY environment variable not set")
	}

	// Create AINote instance
	note, err := NewAINote("", apiKey)
	if err != nil {
		t.Fatalf("Failed to create AINote instance: %v", err)
	}

	testCases := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "Basic article",
			content: `The Impact of Artificial Intelligence on Modern Society

Artificial Intelligence (AI) has become an integral part of our daily lives. From 
smartphones to smart homes, AI technology is transforming how we live and work. 
This technology has shown remarkable progress in various fields including 
healthcare, transportation, and education.

In healthcare, AI systems are helping doctors diagnose diseases more accurately 
and develop personalized treatment plans. In transportation, self-driving cars 
powered by AI are promising to make our roads safer. The education sector is 
seeing AI-powered adaptive learning systems that customize content based on 
student performance.

However, the rise of AI also raises important ethical considerations. Questions 
about privacy, job displacement, and algorithmic bias need careful consideration 
as we move forward with AI development.`,
			wantErr: false,
		},
		{
			name:    "Empty content",
			content: "",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			summary, err := note.SummarizeArticle(tc.content)
			t.Log(summary)
			if (err != nil) != tc.wantErr {
				t.Errorf("SummarizeArticle() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr {
				if !strings.Contains(summary, "#") {
					t.Error("Summary should contain markdown headings")
				}
				if len(summary) == 0 {
					t.Error("Summary should not be empty")
				}
			}
		})
	}
}
