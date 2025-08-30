package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"
)

const (
	promptFileBaseName  = "prompt.txt"
	outputFileBaseName  = "output.md"
	historyFileBaseName = "history.json"
	conversationDir     = "conversations"
	timeFormat          = "20060102T150405"
)

// files are the file and directory paths for the programme output.
type files struct {
	workingDir      string
	conversationDir string
	chatDir         string
	outputFile      string
	chatPromptFile  string
	chatHistoryFile string
	chatOutputFile  string
	timestamp       string
}

// makeDirs simply tries to make the chatDir and parents in workingDir.
func (f *files) makeDirs() error {
	return os.MkdirAll(f.chatDir, 0755)
}

// writePrompt writes the chat prompt file.
func (f *files) WritePrompt(b []byte) error {
	err := os.WriteFile(f.chatPromptFile, b, 0644)
	if err != nil {
		return fmt.Errorf("could not write chat prompt file %s: %s", f.chatPromptFile, err)
	}
	return nil
}

// writeOutput writes the output and chat output files.
func (f *files) WriteOutput(b []byte) error {
	err := os.WriteFile(f.outputFile, b, 0644)
	if err != nil {
		return fmt.Errorf("could not write output file %s: %s", f.outputFile, err)
	}
	err = os.WriteFile(f.chatOutputFile, b, 0644)
	if err != nil {
		return fmt.Errorf("could not write chat output file %s: %s", f.chatOutputFile, err)
	}
	return nil
}

// writeHistory writes the chat history file.
func (f *files) WriteHistory(b []byte) error {
	err := os.WriteFile(f.chatHistoryFile, b, 0644)
	if err != nil {
		return fmt.Errorf("could not write chat history file %s: %s", f.chatHistoryFile, err)
	}
	return f.makeDirs()
}

// LatestHistoryFile finds the latest history file, if any. This is a
// package function. This returns an empty string if no history file is
// found. An example history file name is `20250829T220640_history.json`
// and the creation time is extracted from the filename.
func LatestHistoryFile(path string) string {
	type cf struct {
		creation time.Time
		name     string
	}
	historyFiles := []cf{}
	getCF := func(s string) {
		t, err := time.Parse(timeFormat+"_history.json", s)
		if err != nil {
			return
		}
		historyFiles = append(historyFiles, cf{t, s})
	}
	files, err := os.ReadDir(path)
	if err != nil {
		return "" // path may not have been made yet
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		getCF(f.Name())
	}
	if len(historyFiles) < 1 {
		return ""
	}
	slices.SortFunc(historyFiles, func(a, b cf) int {
		if a.creation.After(b.creation) {
			return -1
		}
		return 1
	})
	return filepath.Join(path, historyFiles[0].name)
}

func NewFiles(workingDir, chat string) (*files, error) {
	ts := time.Now().Format(timeFormat)
	joinTS := func(s string) string {
		return fmt.Sprintf("%s_%s", ts, s)
	}
	if workingDir == "" || chat == "" {
		return nil, fmt.Errorf("workingDir %s or chat %s empty", workingDir, chat)
	}
	f := files{
		workingDir:      workingDir,
		conversationDir: filepath.Join(workingDir, conversationDir),
		chatDir:         filepath.Join(workingDir, conversationDir, chat),
		outputFile:      filepath.Join(workingDir, outputFileBaseName),
		chatPromptFile:  filepath.Join(workingDir, conversationDir, chat, joinTS(promptFileBaseName)),
		chatHistoryFile: filepath.Join(workingDir, conversationDir, chat, joinTS(historyFileBaseName)),
		chatOutputFile:  filepath.Join(workingDir, conversationDir, chat, joinTS(outputFileBaseName)),
		timestamp:       ts,
	}
	err := f.makeDirs()
	return &f, err
}
