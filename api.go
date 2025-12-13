package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"google.golang.org/genai"
)

// GenerateResponse calls the API.
func GenerateResponse(ctx context.Context, settings Settings, history *HistoryFile, newTurn Turn) (*Turn, int, error) {

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  settings.APIKey,
		Backend: genai.BackendGeminiAPI, // Use standard Gemini API backend
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create client: %w", err)
	}

	// Initialise the content history.
	var contents []*genai.Content

	// Count the conversation turns.
	var turns int = 0

	// Add any previous history
	if history != nil {
		for _, turn := range history.Turns {
			c := &genai.Content{
				Role:  turn.Role,
				Parts: make([]*genai.Part, 0),
			}

			// Add Text
			for _, text := range turn.TextParts {
				c.Parts = append(c.Parts, &genai.Part{Text: text})
			}

			// Add Attachments
			for _, att := range turn.Attachments {
				c.Parts = append(c.Parts, &genai.Part{
					InlineData: &genai.Blob{
						MIMEType: att.MIMEType,
						Data:     att.Data,
					},
				})
			}

			// Add Thought Signature (if present and model role)
			if turn.Role == "model" && len(turn.ThoughtSignature) > 0 {
				c.Parts = append(c.Parts, &genai.Part{
					ThoughtSignature: turn.ThoughtSignature,
				})
			}

			// Count only User turns.
			if turn.Role == "user" {
				turns++
			}

			contents = append(contents, c)
		}
	}

	// 2. Add Current User Turn
	userContent := &genai.Content{
		Role:  "user",
		Parts: make([]*genai.Part, 0),
	}
	for _, text := range newTurn.TextParts {
		userContent.Parts = append(userContent.Parts, &genai.Part{Text: text})
	}
	for _, att := range newTurn.Attachments {
		userContent.Parts = append(userContent.Parts, &genai.Part{
			InlineData: &genai.Blob{
				MIMEType: att.MIMEType,
				Data:     att.Data,
			},
		})
	}
	contents = append(contents, userContent)

	// Configure Thinking
	// Check if model supports thinking (simple heuristic or explicit settings)
	// Assuming Gemini 3.0 or 2.0-flash-thinking supports it.
	var thinkingConfig *genai.ThinkingConfig
	if strings.Contains(settings.ModelName, "thinking") || strings.Contains(settings.ModelName, "gemini-3") {
		level := genai.ThinkingLevelHigh
		if settings.ThinkingLevel == "low" {
			level = genai.ThinkingLevelLow
		}
		thinkingConfig = &genai.ThinkingConfig{
			IncludeThoughts: true,
			ThinkingLevel:   level,
		}
	}

	// Call the API.
	if settings.Logging {
		log.Printf("Sending request to %s with %d history turns...", settings.ModelName, turns)
	}

	resp, err := client.Models.GenerateContent(ctx, settings.ModelName, contents, &genai.GenerateContentConfig{
		ThinkingConfig: thinkingConfig,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("api error: %w", err)
	}

	// Parse the API response.
	if len(resp.Candidates) == 0 {
		return nil, 0, fmt.Errorf("no candidates returned")
	}

	cand := resp.Candidates[0]
	modelTurn := &Turn{
		Role:      "model",
		Timestamp: resp.CreateTime,
	}

	// Extract content
	for _, part := range cand.Content.Parts {
		if part.Text != "" {
			if part.Thought {
				modelTurn.Thought += part.Text + "\n"
			} else {
				modelTurn.TextParts = append(modelTurn.TextParts, part.Text)
			}
		}
		if len(part.ThoughtSignature) > 0 {
			modelTurn.ThoughtSignature = part.ThoughtSignature
		}
	}

	// Clean up whitespace
	modelTurn.Thought = strings.TrimSpace(modelTurn.Thought)

	// Tokens
	tokenCount := 0
	if resp.UsageMetadata != nil {
		tokenCount = int(resp.UsageMetadata.TotalTokenCount)
	}

	return modelTurn, tokenCount, nil
}
