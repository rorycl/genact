package genact

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
)

// APIConversation represents each conversation "turn" in an API history
// file. This struct is used for both reading and writing history files.
type APIConversation struct {
	Role  string   `json:"Role"`
	Parts []string `json:"Parts,omitempty"`
}

// ReadAPIHistory reads json history from an API history file, returning
// a slice of APIConversation.
func ReadAPIHistory(filePath string) ([]APIConversation, error) {
	var previousHistory []APIConversation
	historyBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read api history file: %v", err)
	}
	if err := json.Unmarshal(historyBytes, &previousHistory); err != nil {
		return nil, fmt.Errorf("failed to parse ai history file: %v", err)
	}
	return previousHistory, nil
}

// aiAPIToAIContent converts a slice of APIConversation to a slice of
// genai.Content.
func apiToAIContent(jc []APIConversation) ([]*genai.Content, error) {
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
		return nil, errors.New("the provided api history file is effectively empty")
	}
	return contents, nil
}

// HistoryAPIToAIContent parses an API history file from Google Gemini
// into a slice of *genai.Content, or error.
func HistoryAPIToAIContent(filePath string) ([]*genai.Content, error) {
	aiAPIExport, err := ReadAPIHistory(filePath)
	if err != nil {
		return nil, err
	}
	return apiToAIContent(aiAPIExport)
}
