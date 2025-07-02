package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
)

// history file json structure
type APIJsonContent struct {
	Role  string   `json:"Role"`
	Parts []string `json:"Parts,omitempty"`
}

// readAPIHistory reads json history from an AI API history file.
func readAPIHistory(filePath string) ([]APIJsonContent, error) {
	var previousHistory []APIJsonContent
	historyBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read api history file: %v", err)
	}
	if err := json.Unmarshal(historyBytes, &previousHistory); err != nil {
		return nil, fmt.Errorf("Failed to parse ai history file: %v", err)
	}
	return previousHistory, nil
}

// aiAPIToAIContent converts a slice of APIJsonContent to a slice of
// genai.Content.
func apiToAIContent(jc []APIJsonContent) ([]*genai.Content, error) {
	contents := []*genai.Content{}
	for _, thisContent := range jc {
		c := genai.Content{}
		c.Role = thisContent.Role
		for _, p := range thisContent.Parts {
			if p == "" {
				continue
			}
			c.Parts = append(c.Parts, genai.Text(p)) // convert to a Text part type
		}
		if len(c.Parts) == 0 {
			continue
		}
		contents = append(contents, &c)
	}
	if len(contents) == 0 {
		return nil, errors.New("the provided api history file is effectively empty.")
	}
	return contents, nil
}

// HistoryAPIToAIContent parses an AI API history file from Google
// Gemini into a slice of *genai.Content, or error.
func HistoryAPIToAIContent(filePath string) ([]*genai.Content, error) {
	aiAPIExport, err := readAPIHistory(filePath)
	if err != nil {
		return nil, err
	}
	return apiToAIContent(aiAPIExport)
}
