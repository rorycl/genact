package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// LoadHistory loads a history file from disk.
func LoadHistory(path string) (*HistoryFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var h HistoryFile
	if err := json.Unmarshal(data, &h); err != nil {
		return nil, err
	}
	return &h, nil
}

// MigrateLegacy loads a v0.0.4 file and upgrades it.
func MigrateLegacy(path string) (*HistoryFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var oldData []LegacyConversation
	if err := json.Unmarshal(data, &oldData); err != nil {
		return nil, fmt.Errorf("invalid legacy format: %w", err)
	}

	ts := time.Now()
	id := ts.Format("20060102T150405")

	h := &HistoryFile{
		ID:         id,
		ParentID:   "legacy_v0.0.4",
		ForkReason: "legacy_import",
		Turns:      make([]Turn, 0, len(oldData)),
	}

	for _, old := range oldData {
		h.Turns = append(h.Turns, Turn{
			Role:      old.Role,
			TextParts: old.Parts,
			Timestamp: ts, // Approximation
		})
	}

	return h, nil
}

// StripSignatures removes ThoughtSignatures from a history object.
func StripSignatures(h *HistoryFile) {
	for i := range h.Turns {
		h.Turns[i].ThoughtSignature = nil
		// We keep the 'Thought' text string for human reference, 
		// but removing the signature forces the model to re-digest context.
	}
}
