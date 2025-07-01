package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const (
	markdownOutputFile = "output.txt"
	modelName          = "gemini-2.5-pro"
	apiKey             = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
)

// AIStudioChunk represents a single turn in the conversation from the export file.
type AIStudioChunk struct {
	Text string `json:"text"`
	Role string `json:"role"`
	// ignore other fields.
}

// AIStudioExport represents the top-level structure of the JSON file
// exported from Google AI Studio.
type AIStudioExport struct {
	ChunkedPrompt struct {
		Chunks []AIStudioChunk `json:"chunks"`
	} `json:"chunkedPrompt"`
	// We can ignore other top-level fields like runSettings, systemInstruction, etc.
}

// history file json structure
type jsonContent struct {
	Role  string   `json:"Role"`
	Parts []string `json:"Parts,omitempty"`
}

// readHistory reads json history from a file
func readHistory(file string) ([]jsonContent, error) {
	var previousHistory []jsonContent
	historyBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Failed to read history file: %v", err)
	}
	if err := json.Unmarshal(historyBytes, &previousHistory); err != nil {
		return nil, fmt.Errorf("Failed to parse history.json: %v", err)
	}
	return previousHistory, nil
}

// toAIContent converts json history to a slice of genai.Content.
func historyToAIContent(jc []jsonContent) []*genai.Content {
	contents := []*genai.Content{}
	for _, thisContent := range jc {
		c := genai.Content{}
		c.Role = thisContent.Role
		for _, p := range thisContent.Parts {
			c.Parts = append(c.Parts, genai.Text(p)) // convert to a Text part type
		}
		contents = append(contents, &c)
	}
	return contents
}

// readStudioHistory reads json history from an AI Studio history file.
func readStudioHistory(file string) (AIStudioExport, error) {
	var studioExport AIStudioExport
	historyBytes, err := os.ReadFile(historyInputFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read studio history file: %v", err)
	}
	if err := json.Unmarshal(historyBytes, &studioExport); err != nil {
		return nil, fmt.Errorf("Failed to parse studio history.json: %v", err)
	}
	return studioExport, nil
}

// aiStudioToAIContent converts a slice of AIStudioChunk to a slice of
// genai.Content.
func aiStudioToAIContent(studioChunks []AIStudioChunk) []*genai.Content {
	contents := []*genai.Content{}
	for _, thisContent := range studioChunks {
		c := genai.Content{}
		c.Role = thisContent.Role
		c.Parts = append(c.Parts, genai.Text(thisContent.Text)) // convert to a Text part type
		contents = append(contents, &c)
	}
	return contents
}

// generateTimeString makes a string from the current time.
func generateTimeString() string {
	return time.Now().Format("20060102T150405")
}

// copyFile copies a file src to dst
func copyFile(src, dst string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", src, err)
	}

	err = ioutil.WriteFile(dst, input, 0644)
	if err != nil {
		return fmt.Errorf("error creating %s, %w", dst, err)
	}
	return nil
}

func main() {
	// 1. Get API Key from environment variable
	/*
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			log.Fatal("GEMINI_API_KEY environment variable not set.")
		}
	*/

	ctx := context.Background()

	// check arguments
	if len(os.Args) != 3 {
		fmt.Println("need 3 arguments: 1. history 2. prompt")
		os.Exit(1)
	}

	historyInputFile := os.Args[1]
	promptInputFile := os.Args[2]

	// 2. Read the new prompt from prompt.txt
	log.Printf("Reading new prompt from %s...", promptInputFile)
	promptBytes, err := os.ReadFile(promptInputFile)
	if err != nil {
		log.Fatalf("Failed to read prompt file: %v", err)
	}
	newPrompt := string(promptBytes)

	// 3. Read and parse the existing history.json
	// var chatHistory []*genai.Content
	log.Printf("Reading existing history from %s...", historyInputFile)

	var previousHistory []jsonContent
	historyBytes, err := os.ReadFile(historyInputFile)
	if err != nil {
		log.Fatalf("Failed to read history file: %v", err)
	}

	if err := json.Unmarshal(historyBytes, &previousHistory); err != nil {
		log.Fatalf("Failed to parse history.json: %v", err)
	}
	chatHistory := historyToAIContent(previousHistory)

	// 4. Initialize the Gemini client
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel(modelName)
	log.Printf("Initialized model: %s", modelName)

	// 5. Start a chat session and load the history
	chat := model.StartChat()
	chat.History = chatHistory

	log.Println("Sending prompt to Gemini API...")
	resp, err := chat.SendMessage(ctx, genai.Text(newPrompt))
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	// 6. Process the response and extract the latest message
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		log.Fatal("Received an empty response from the API.")
	}

	var latestResponse strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			latestResponse.WriteString(string(txt))
		}
	}
	log.Println("Successfully received response.")

	// get a string timestamp
	timestamp := generateTimeString()

	// 7. Write the markdown output file
	err = os.WriteFile(markdownOutputFile, []byte(latestResponse.String()), 0644)
	if err != nil {
		log.Fatalf("Failed to write output.txt: %v", err)
	}
	log.Printf("Latest response saved to %s", markdownOutputFile)

	// also write to conversations
	newOutputFilename := fmt.Sprintf("conversations/%s-output.md", timestamp)
	err = os.WriteFile(newOutputFilename, []byte(latestResponse.String()), 0644)
	if err != nil {
		log.Fatalf("Failed to write conversations output.txt: %v", err)
	}

	// 8. Copy prompt to conversations
	newPromptFilename := fmt.Sprintf("conversations/%s-prompt.txt", timestamp)
	err = copyFile(promptInputFile, newPromptFilename)

	// 9. Write the new, full history to a timestamped JSON file
	fullHistory := chat.History
	historyJSON, err := json.MarshalIndent(fullHistory, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal new history: %v", err)
	}

	// record stuff
	newHistoryFilename := fmt.Sprintf("conversations/%s-history.json", timestamp)

	err = os.WriteFile(newHistoryFilename, historyJSON, 0644)
	if err != nil {
		log.Fatalf("Failed to write new history file: %v", err)
	}
	log.Printf("Full conversation history saved to %s", newHistoryFilename)
}
