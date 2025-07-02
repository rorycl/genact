package main

import (
	"testing"
)

// TestHistoryStudio tests the parsing of a history file downloaded from
// Google AI Studio.
func TestHistoryStudio(t *testing.T) {
	file := "testdata/studio-history.json"
	h, err := HistoryStudioToAIContent(file)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(h), 23; got != want { // 12 additional are thoughts
		t.Errorf("got %d want %d contents", got, want)
	}
}
