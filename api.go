// Package genact provides a way for interacting with the Google Genesis
// AI API for text only prompts using the text chat interface. This
// module provides a simple mechanism for sending prompts and recorded
// history (if applicable) and receiving a response and the new history,
// suitable for iterative interactions with Genesis Pro 2.5 taking
// advantage of large Genesis token windows.
package genact

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type ApiResponse struct {
	TokenCount     int32
	LatestResponse string
	FullHistory    string
}

var logger *log.Logger

// initalise logging
func newLogger(enabled bool) {
	writer := io.Discard
	if enabled {
		writer = os.Stdout
	}
	logger = log.New(writer, "", log.LstdFlags)
}

// startChat starts a client/model/chat.
func startChat(ctx context.Context, settings map[string]string) (*genai.Client, *genai.ChatSession, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(settings["apiKey"]))
	if err != nil {
		return nil, nil, fmt.Errorf("could not create client: %v", err)
	}
	model := client.GenerativeModel(settings["modelName"])
	chat := model.StartChat()
	return client, chat, nil
}

// endChat closes a client session.
func endChat(client *genai.Client) {
	client.Close()
}

// runAPI runs the api given a *genai.ChatSession, history (if any) and
// a prompt string.
func runAPI(ctx context.Context, chat *genai.ChatSession, history []*genai.Content, prompt string) (*genai.GenerateContentResponse, error) {

	chat.History = history

	logger.Println("Sending prompt to Gemini API...")
	resp, err := chat.SendMessage(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("Failed to send message: %v", err)
	}
	if resp.PromptFeedback != nil && resp.PromptFeedback.BlockReason > 0 {
		return nil, fmt.Errorf("Received BlockReason: %d", resp.PromptFeedback.BlockReason)
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("Received an empty response from the API.")
	}
	logger.Println("Response ok")

	return resp, nil
}

// parseResponse takes a *genai.ChatSession and a
// *genai.GenerateContentResponse to parse a responseinto a local
// ApiResponse struct for easier handling. Presently only the first
// Candidate is considered (resp.Candidates[0]).
//
// Todo: deal with "isThinking"/"Thinking" responses from the AI API,
// which may not be useful to put into history.
func parseResponse(chat *genai.ChatSession, resp *genai.GenerateContentResponse) (*ApiResponse, error) {

	thisResponse := ApiResponse{
		TokenCount: resp.UsageMetadata.PromptTokenCount,
	}

	// expecting only 1 candidate in this code
	if len(resp.Candidates) != 1 {
		return nil, fmt.Errorf("expected only 1 candidate, got %d", len(resp.Candidates))
	}

	var LatestResponse strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			LatestResponse.WriteString(string(txt))
		}
	}
	thisResponse.LatestResponse = LatestResponse.String()
	if thisResponse.LatestResponse == "" {
		return nil, errors.New("latest response had no text content")
	}

	FullHistory := chat.History
	historyJSON, err := json.MarshalIndent(FullHistory, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal new history: %v", err)
	}
	thisResponse.FullHistory = string(historyJSON)
	return &thisResponse, nil
}

// APIGetResponse creates a genai client and chat session, runs the api
// to receive a response, and then puts the response into a local
// ApiResponse struct for convenient processing.
func APIGetResponse(settings map[string]string, history []*genai.Content, prompt string) (*ApiResponse, error) {

	if settings == nil {
		return nil, errors.New("settings not provided")
	}
	logging := !(settings["logging"] == "false")
	newLogger(logging)

	ctx := context.Background()
	client, chat, err := startChat(ctx, settings)
	defer endChat(client)
	if err != nil {
		return nil, fmt.Errorf("could not start chat: %w", err)
	}
	response, err := runAPI(ctx, chat, history, prompt)
	if err != nil {
		return nil, fmt.Errorf("chat response error: %w", err)
	}
	return parseResponse(chat, response)
}
