package main

import "encoding/json"

// API response/history file json structure part.
type Part struct {
	Text    string `json:"text"`
	Thought bool   `json:"thought,omitempty"`
}

// API response/history file json structure.
type ApiJsonContent struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts,omitempty"`
}

// MarshalJSON marshals an ApiJsonContent object to json.
func (a *ApiJsonContent) MarshalJSON() ([]byte, error) {
	return json.MarshalIndent(a, "", "  ")
}
