package genact

import (
	"testing"
)

// TestHistoryAPI tests the parsing of a history file downloaded through
// the API (normally aggregated from previous API calls).
func TestHistoryAPI(t *testing.T) {
	file := "testdata/api-history-tennis.json"
	h, err := HistoryAPIToAIContent(file)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(h), 8; got != want {
		t.Errorf("got %d want %d contents", got, want)
	}
}
