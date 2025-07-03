package main

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestConversations(t *testing.T) {

	file := "../../testdata/api-history-tennis.json"
	conversations, err := NewConversations(file)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := conversations.Len(), 4; got != want {
		fmt.Errorf("got %d want %d length conversation", got, want)
	}

	first := conversations.Get(0)
	conversations.Reverse()
	last := conversations.Get(0) // retrieved by original index

	if got, want := first.idx, last.idx; got != want {
		t.Errorf("got %d should equal %d after reverse", got, want)
	}
	if got, want := conversations.reversed, true; got != want {
		t.Errorf("got %t want %t for reversed", got, want)
	}

	// keep only two items from
	// natural index 0 1 2 3
	//         idx   3 2 1 0
	err = conversations.Keep(2)
	if err != nil {
		t.Fatal(err)
	}
	err = conversations.Keep(4) // past end
	if err == nil {
		t.Fatal("expected Keep(4) to error")
	}
	err = conversations.Keep(1) // item 2
	if err != nil {
		t.Fatal(err)
	}

	// compact and check
	conversations.Compact()

	if got, want := conversations.Len(), 2; got != want {
		fmt.Errorf("got %d want %d length conversation", got, want)
	}

	indexes := []int{}
	for _, c := range conversations.conversations {
		indexes = append(indexes, c.idx)
	}

	if diff := cmp.Diff(indexes, []int{1, 2}); diff != "" {
		t.Errorf("indexes mismatch (-want +got):\n%s", diff)
	}

}
