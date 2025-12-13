package app

import (
	"encoding/json"
	"time"
)

// Attachment represents a file embedded in the history.
type Attachment struct {
	MIMEType string `json:"mime_type"`
	FileName string `json:"file_name,omitempty"`
	Data     []byte `json:"data"` // Base64 encoded content automatically by encoding/json
}

// Turn represents a single exchange in the conversation.
type Turn struct {
	Role        string       `json:"role"` // "user" or "model"
	TextParts   []string     `json:"text_parts,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`

	// Gemini 3.0 Specifics
	Thought          string    `json:"thought,omitempty"`           // The visible reasoning text
	ThoughtSignature []byte    `json:"thought_signature,omitempty"` // The opaque, encrypted state blob
	Timestamp        time.Time `json:"timestamp"`
}

// HistoryFile is the top-level structure for a saved conversation.
type HistoryFile struct {
	ID            string `json:"id"`             // Creation timestamp string (YYYYMMDDTHHMMSS)
	ParentID      string `json:"parent_id"`      // The ID of the history this was branched from
	OriginalModel string `json:"original_model"` // The model used for the last turn
	ForkReason    string `json:"fork_reason"`    // e.g. "new", "branch", "strip_signature", "legacy_import"
	Turns         []Turn `json:"turns"`
	TotalTokens   int    `json:"total_tokens"`
}

// DeepCopy creates a fork of the history.
func (h *HistoryFile) DeepCopy(newID, forkReason string) *HistoryFile {
	newH := &HistoryFile{
		ID:            newID,
		ParentID:      h.ID,
		OriginalModel: h.OriginalModel,
		ForkReason:    forkReason,
		TotalTokens:   h.TotalTokens,
		Turns:         make([]Turn, len(h.Turns)),
	}
	copy(newH.Turns, h.Turns)
	return newH
}

// Serialize returns the JSON representation.
func (h *HistoryFile) Serialize() ([]byte, error) {
	return json.MarshalIndent(h, "", "  ")
}

// LegacyConversation for migration (v0.0.4)
type LegacyConversation struct {
	Role  string   `json:"Role"`
	Parts []string `json:"Parts,omitempty"`
}
