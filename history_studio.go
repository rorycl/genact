package genact

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
)

// AIStudioChunk represents a single turn in the conversation from a
// Google AI Studio export file.
type AIStudioChunk struct {
	Text       string `json:"text"`
	Role       string `json:"role"`
	IsThought  bool   `json:"isThought"`
	TokenCount int32  `json:"tokenCount"`
	// ignore other fields.
}

// AIStudioExport represents the top-level structure of the JSON file
// exported from Google AI Studio containing a slice of AIStudioChunk.
type AIStudioExport struct {
	ChunkedPrompt struct {
		Chunks []AIStudioChunk `json:"chunks"`
	} `json:"chunkedPrompt"`
	// ignore other settings etc.
}

// readStudioHistory reads json history from an AI Studio history file.
func readStudioHistory(filePath string) (*AIStudioExport, error) {
	var studioExport AIStudioExport
	historyBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read studio history file: %v", err)
	}
	if err := json.Unmarshal(historyBytes, &studioExport); err != nil {
		return nil, fmt.Errorf("failed to parse studio history file: %v", err)
	}
	if len(studioExport.ChunkedPrompt.Chunks) == 0 {
		return nil, fmt.Errorf("no history found in %s", filePath)
	}
	return &studioExport, nil
}

// aiStudioToAIContent converts a slice of AIStudioChunk to a slice of
// genai.Content.
func aiStudioToAIContent(studioChunks []AIStudioChunk) ([]*genai.Content, error) {
	contents := []*genai.Content{}
	for _, thisContent := range studioChunks {
		if thisContent.Text == "" {
			continue
		}
		if thisContent.IsThought { // don't load thoughts
			continue
		}
		c := genai.Content{}
		c.Role = thisContent.Role
		c.Parts = append(c.Parts, genai.Text(thisContent.Text)) // convert to a Text part type
		contents = append(contents, &c)
	}
	if len(contents) == 0 {
		return nil, errors.New("the provided studio history file is effectively empty")
	}
	return contents, nil
}

// HistoryStudioToAIContent parses a history file saved from Google
// Gemini AI Studio into a slice of *genai.Content, or error.
func HistoryStudioToAIContent(filePath string) ([]*genai.Content, error) {
	aiStudioExport, err := readStudioHistory(filePath)
	if err != nil {
		return nil, err
	}
	return aiStudioToAIContent(aiStudioExport.ChunkedPrompt.Chunks)
}
