package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"genact/internal/filechk"
)

// App represents the main gathering point of the application, bringing
// together application logic which can be injected into other
// components, such as the cli.
type App struct{}

// NewApp returns a new App.
func NewApp() *App {
	return &App{}
}

// ConversationOptions are the options to pass to Converse.
type ConverseOptions struct {
	ConversationName string
	PromptPath       string
	SettingsPath     string
	HistoryPath      string
	ThinkingLevel    string
	IsNew            bool
	IsLegacyHistory  bool
	Attachments      []string
}

// Validate validates the options.
func (c ConverseOptions) Validate() error {

	if c.PromptPath == "" {
		return errors.New("prompt file not specified")
	}
	if !filechk.IsFile(c.PromptPath) {
		return fmt.Errorf("prompt file %q not found", c.PromptPath)
	}

	if c.ConversationName == "" {
		return errors.New("conversation name `-c` cannot be empty")
	}

	if !filechk.IfNotEmptyAndIsFile(c.HistoryPath) {
		return fmt.Errorf("history file %q not found", c.HistoryPath)
	}

	for _, a := range c.Attachments {
		if filechk.IsFile(a) == false {
			return fmt.Errorf("attachment %q not found", a)
		}
	}
	return nil
}

// Converse runs a conversation turn with the Gemini API.
func (a *App) Converse(ctx context.Context, cfg ConverseOptions) error {

	if err := cfg.Validate(); err != nil {
		return err
	}

	settings, err := LoadYaml(cfg.SettingsPath)
	if err != nil {
		return err
	}

	// Override thinkingLevel in settings based on config.
	settings.ThinkingLevel = cfg.ThinkingLevel

	paths, err := NewFilePaths("", cfg.ConversationName, cfg.IsNew, false)
	if err != nil {
		return err
	}

	// Determine history file management. Note that if forking, we
	// define the PARENT as the ID of the loaded file.
	// The history is defaulted to a "new", history-less HistoryPath.
	var forkReason string
	var parentID string
	var history *HistoryFile

	initHistory := func() *HistoryFile {
		return &HistoryFile{
			ID:            "",
			ParentID:      "",
			ForkReason:    "new",
			OriginalModel: settings.ModelName,
			Turns:         []Turn{},
		}
	}

	switch {
	case cfg.IsNew:
		history = initHistory()

	case cfg.HistoryPath != "" && cfg.IsLegacyHistory:
		history, err = MigrateLegacy(cfg.HistoryPath)
		if err != nil {
			return err
		}
		forkReason = "legacy_fork"
		parentID = history.ID

	case cfg.HistoryPath != "":
		history, err = LoadHistory(cfg.HistoryPath)
		if err != nil {
			return err
		}
		forkReason = "branch"
		parentID = history.ID

	default:
		latest, err := FindLatestHistory(paths.ChatDir)
		if err != nil {
			return err
		}
		if latest == "" {
			history = initHistory()
		}
		history, err = LoadHistory(latest)
		if err != nil {
			return fmt.Errorf("failed to load previous history: %w", err)
		}
		forkReason = "reply"
		parentID = history.ID
	}

	// Create a new ID for this session
	sessionID := paths.Timestamp

	// Prepare User Turn
	promptBytes, err := os.ReadFile(cfg.PromptPath)
	if err != nil {
		return err
	}
	userTurn := Turn{
		Role:      "user",
		TextParts: []string{string(promptBytes)},
		Timestamp: time.Now(),
	}

	// Load attachments
	for _, ap := range cfg.Attachments {
		att, err := LoadAttachment(ap)
		if err != nil {
			return err
		}
		userTurn.Attachments = append(userTurn.Attachments, att)
	}

	// API Call
	log.Printf("Generating response using %s...\n", settings.ModelName)
	modelTurn, tokenCount, err := GenerateResponse(ctx, settings, history, userTurn)
	if err != nil {
		return err
	}

	// Update History and Save, creating a new history
	// file with:
	// [Ancestors] + [UserTurn] + [ModelTurn]
	newHistory := history.DeepCopy(sessionID, forkReason)
	newHistory.ParentID = parentID
	newHistory.OriginalModel = settings.ModelName
	newHistory.TotalTokens = tokenCount
	newHistory.Turns = append(newHistory.Turns, userTurn, *modelTurn)

	historyBytes, _ := newHistory.Serialize()
	if err := os.WriteFile(paths.HistoryFile, historyBytes, 0644); err != nil {
		return err
	}

	// Write Outputs
	fullOutput := strings.Join(modelTurn.TextParts, "\n\n")
	if modelTurn.Thought != "" {
		// consider using an anonymous struct for this to look cleaner.
		fullOutput = fmt.Sprintf("<details><summary>Thinking</summary>\n\n%s\n\n</details>\n\n%s", modelTurn.Thought, fullOutput)
	}

	os.WriteFile(paths.ResponseFile, []byte(fullOutput), 0644)
	os.WriteFile(paths.LocalResponseFile, []byte(fullOutput), 0644) // save snapshot of output
	os.WriteFile(paths.PromptFile, promptBytes, 0644)               // save snapshot of prompt

	log.Printf("Done. Saved to %s\n", paths.HistoryFile)

	return nil
}

// Regenerate regenerates a history file to remove its thought
// signatures.
func (a *App) Regenerate(inputPath string) error {

	history, err := LoadHistory(inputPath)
	if err != nil {
		return err
	}

	// Perform stripping
	StripSignatures(history)

	// Update Metadata
	ts := time.Now().Format(timeFormat)
	history.ParentID = history.ID
	history.ID = ts
	history.ForkReason = "strip_signatures"

	// Save to same dir with new timestamp
	dir := filepath.Dir(inputPath)
	newPath := filepath.Join(dir, fmt.Sprintf("%s_history.json", ts))

	bytes, _ := history.Serialize()
	if err := os.WriteFile(newPath, bytes, 0644); err != nil {
		return err
	}

	fmt.Printf("Stripped signatures. New file: %s\n", newPath)
	return nil
}

// ParseFiles does a one-shot analysis of the attachments for OCR-type
// extraction using Gemini, using the instructions in the prompt file.
func (a *App) ParseFiles(ctx context.Context, settingsPath, promptPath string, attachments []string) error {

	settings, err := LoadYaml(settingsPath)
	if err != nil {
		return err
	}

	// Setup "files" directory structure
	paths, err := NewFilePaths("", "parsed", true, true) // "parsed" is a generic bucket name in files/
	if err != nil {
		return err
	}

	// Prepare User Turn (No History)
	promptBytes, err := os.ReadFile(promptPath)
	if err != nil {
		return err
	}

	userTurn := Turn{
		Role:      "user",
		TextParts: []string{string(promptBytes)},
		Timestamp: time.Now(),
	}

	for _, f := range attachments {
		att, err := LoadAttachment(f)
		if err != nil {
			return err
		}
		userTurn.Attachments = append(userTurn.Attachments, att)
	}

	fmt.Println("Parsing file(s)...")
	// Pass nil history
	modelTurn, _, err := GenerateResponse(ctx, settings, nil, userTurn)
	if err != nil {
		return err
	}

	// Write Output
	fullOutput := strings.Join(modelTurn.TextParts, "\n\n")
	if err := os.WriteFile(paths.ResponseFile, []byte(fullOutput), 0644); err != nil {
		return err
	}

	// Do not save history.json for parsefile, only save response
	// markdown.
	fmt.Printf("Output saved to %s\n", paths.ResponseFile)
	return nil

}

// Lineage reports on the lineage of a conversation.
func (a *App) Lineage(conversation string) error {
	paths, err := NewFilePaths("", conversation, false, false)
	if err != nil {
		return err
	}

	files, err := ScanForLineage(paths.ChatDir)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		fmt.Println("No history found.")
		return nil
	}

	// Simple output
	fmt.Printf("Lineage for %s:\n", conversation)
	fmt.Println("ID                  | Parent              | Reason             | Model          | Tokens")
	fmt.Println("---------------------------------------------------------------------------------------")
	for _, f := range files {
		parent := f.ParentID
		if len(parent) > 15 {
			parent = "..." + parent[len(parent)-12:]
		}
		if parent == "" {
			parent = "root"
		}

		fmt.Printf("%-19s | %-19s | %-18s | %-14s | %d\n",
			f.ID, parent, f.ForkReason, f.OriginalModel, f.TotalTokens)
	}

	return nil
}
