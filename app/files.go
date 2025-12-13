package app

import (
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const (
	promptFileBaseName  = "prompt.txt"
	outputFileBaseName  = "output.md"
	historyFileBaseName = "history.json"
	conversationDir     = "conversations"
	filesDir            = "files"
	timeFormat          = "20060102T150405"
)

// LoadAttachment reads a file from disk and returns an Attachment struct.
func LoadAttachment(path string) (Attachment, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Attachment{}, fmt.Errorf("failed to read attachment %s: %w", path, err)
	}

	mimeType := mime.TypeByExtension(filepath.Ext(path))
	if mimeType == "" {
		// Fallback for common types if OS mime db is missing
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".md", ".txt", ".go", ".json", ".yaml":
			mimeType = "text/plain"
		case ".pdf":
			mimeType = "application/pdf"
		case ".png":
			mimeType = "image/png"
		case ".jpg", ".jpeg":
			mimeType = "image/jpeg"
		default:
			mimeType = "application/octet-stream"
		}
	}

	return Attachment{
		FileName: filepath.Base(path),
		MIMEType: mimeType,
		Data:     data,
	}, nil
}

// FilePaths manages the directory structure.
type FilePaths struct {
	BaseDir           string
	ChatDir           string
	Timestamp         string
	PromptFile        string
	ResponseFile      string
	LocalResponseFile string
	HistoryFile       string
	HistoryLinkFile   string // symlink to latest
}

func NewFilePaths(baseDir, chatName string, isNewChat, isParseFile bool) (*FilePaths, error) {
	ts := time.Now().Format(timeFormat)

	rootDir := conversationDir
	if isParseFile {
		rootDir = filesDir
	}

	// Clean paths
	if baseDir == "" {
		var err error
		baseDir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	// Sanitize chat name
	chatName = strings.ReplaceAll(strings.ToLower(strings.TrimSpace(chatName)), " ", "_")

	// /conversations/<chat>/
	fullChatDir := filepath.Join(baseDir, rootDir, chatName)

	// Check the status of the target directory. For new conversations,
	// the directory should not exist. For continuing conversations
	// (without the isNewChat flag), the directory should exist.
	stat, err := os.Stat(fullChatDir)
	var pe fs.PathError
	switch {
	case isNewChat && !errors.As(err, &pe):
		return nil, fmt.Errorf("directory %q already exists", fullChatDir)
	case !isNewChat && errors.Is(err, &pe):
		return nil, fmt.Errorf("chat %q not initialised; used 'new'", chatName)
	case !isNewChat && !stat.IsDir():
		return nil, fmt.Errorf("%q is not a directory", fullChatDir)
	}

	if err := os.MkdirAll(fullChatDir, 0755); err != nil {
		return nil, err
	}

	return &FilePaths{
		BaseDir:           baseDir,
		ChatDir:           fullChatDir,
		Timestamp:         ts,
		PromptFile:        filepath.Join(fullChatDir, fmt.Sprintf("%s_%s", ts, promptFileBaseName)),
		ResponseFile:      filepath.Join(fullChatDir, fmt.Sprintf("%s_%s", ts, outputFileBaseName)),
		LocalResponseFile: filepath.Join(baseDir, "output.md"),
		HistoryFile:       filepath.Join(fullChatDir, fmt.Sprintf("%s_%s", ts, historyFileBaseName)),
	}, nil
}

// FindLatestHistory scans the chat directory for the most recent history file.
func FindLatestHistory(chatDir string) (string, error) {
	entries, err := os.ReadDir(chatDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	var candidates []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), "_history.json") {
			candidates = append(candidates, filepath.Join(chatDir, e.Name()))
		}
	}

	if len(candidates) == 0 {
		return "", nil
	}

	// Sort by filename (which includes timestamp) descending
	slices.Sort(candidates)
	slices.Reverse(candidates)

	return candidates[0], nil
}

// ScanForLineage finds all history files in a chat directory.
func ScanForLineage(chatDir string) ([]*HistoryFile, error) {
	entries, err := os.ReadDir(chatDir)
	if err != nil {
		return nil, err
	}

	var files []*HistoryFile
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), "_history.json") {
			path := filepath.Join(chatDir, e.Name())
			hf, err := LoadHistory(path)
			if err != nil {
				continue // skip corrupt files
			}
			files = append(files, hf)
		}
	}
	return files, nil
}
