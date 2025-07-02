package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type ApiResponse struct {
	tokenCount     int32
	latestResponse string
	fullHistory    string
}

// local debugging
var logging bool = true

func startChat(ctx context.Context, settings map[string]string) (*genai.Client, *genai.ChatSession, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(settings["apiKey"]))
	if err != nil {
		return nil, nil, fmt.Errorf("could not create client: %v", err)
	}
	model := client.GenerativeModel(settings["modelName"])
	chat := model.StartChat()
	return client, chat, nil
}

func endChat(client *genai.Client) {
	client.Close()
}

// runAPI runs the api given a *genai.ChatSession, history (if any) and
// a prompt string.
func runAPI(ctx context.Context, chat *genai.ChatSession, history []*genai.Content, prompt string) (*genai.GenerateContentResponse, error) {

	logline := func(s string) {
		if !logging {
			return
		}
		log.Println(s)
	}

	chat.History = history

	logline("Sending prompt to Gemini API...")
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
	logline(fmt.Sprintf("Response ok, total token count %d", resp.UsageMetadata.PromptTokenCount))

	return resp, nil
}

// parseResponse takes a *genai.ChatSession and a
// *genai.GenerateContentResponse to parse a responseinto a local
// ApiResponse struct for easier handling. Presently only the first
// Candidate is considered (resp.Candidates[0]).
//
// Todo: deal with "isThinking" responses from the AI API, which may not
// be useful to put into history.
func parseResponse(chat *genai.ChatSession, resp *genai.GenerateContentResponse) (*ApiResponse, error) {

	thisResponse := ApiResponse{
		tokenCount: resp.UsageMetadata.PromptTokenCount,
	}

	// expecting only 1 candidate in this code
	if len(resp.Candidates) != 1 {
		return nil, fmt.Errorf("expected only 1 candidate, got %d", len(resp.Candidates))
	}

	var latestResponse strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			latestResponse.WriteString(string(txt))
		}
	}
	thisResponse.latestResponse = latestResponse.String()
	if thisResponse.latestResponse == "" {
		return nil, errors.New("latest response had no text content")
	}

	fullHistory := chat.History
	historyJSON, err := json.MarshalIndent(fullHistory, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal new history: %v", err)
	}
	thisResponse.fullHistory = historyJSON
	return &thisResponse, nil
}

// APIGetResponse creates a genai client and chat session, runs the api
// to receive a response, and then puts the response into a local
// ApiResponse struct for convenient processing.
func APIGetResponse(settings map[string]string, history []*genai.Content, prompt string) (*ApiResponse, error) {
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
