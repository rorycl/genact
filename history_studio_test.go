package genact

import (
	"testing"
)

// TestHistoryStudio tests the parsing of a history file downloaded from
// Google AI Studio.
func TestHistoryStudio(t *testing.T) {
	file := "testdata/studio-history-tennis.json"
	h, err := HistoryStudioToAIContent(file)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(h), 6; got != want { // 3 additional are thoughts
		t.Errorf("got %d want %d contents", got, want)
	}
}
