package genact

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestConversations(t *testing.T) {

	file := "testdata/api-history-tennis.json"
	conversations, err := NewConversations(file)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := conversations.Len(), 4; got != want {
		t.Errorf("got %d want %d length conversation", got, want)
	}

	first := conversations.Get(0)
	conversations.Reverse()
	last := conversations.Get(0) // retrieved by original index

	strOutput := fmt.Sprint(last)
	if !strings.Contains(strOutput, "`agent`:") {
		t.Error("last item string output did not include `agent`")
	}

	if got, want := first.Idx, last.Idx; got != want {
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
		t.Errorf("got %d want %d length conversation", got, want)
	}

	indexes := []int{}
	for _, c := range conversations.conversations {
		indexes = append(indexes, c.Idx)
	}

	if diff := cmp.Diff(indexes, []int{1, 2}); diff != "" {
		t.Errorf("indexes mismatch (-got +want):\n%s", diff)
	}

	indexes = []int{}
	for c := range conversations.Iter() {
		indexes = append(indexes, c.Idx)
	}

	if diff := cmp.Diff(indexes, []int{1, 2}); diff != "" {
		t.Errorf("indexes mismatch (-got +want):\n%s", diff)
	}

	json, err := conversations.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(json))

}

func TestConversationsReview(t *testing.T) {

	file := "testdata/api-history-tennis.json"
	conversations, err := NewConversations(file)
	if err != nil {
		t.Fatal(err)
	}

	err = conversations.ReviewItems([]int{5})
	if err == nil {
		t.Fatal("unexpected nil error for ReviewItems(5)")
	}
	err = conversations.ReviewItems([]int{-5})
	if err == nil {
		t.Fatal("unexpected nil error for ReviewItems(-5)")
	}
	err = conversations.ReviewItems([]int{0, -1})
	if err != nil {
		t.Fatalf("unexpected error for ReviewItems(0, -1) %v", err)
	}

	if diff := cmp.Diff(conversations.itemsToReview, map[int]bool{0: true, 3: true}); diff != "" {
		t.Errorf("itemsToReview mismatch (-want +got):\n%s", diff)
	}

	counter := 0
	for _ = range conversations.Iter() {
		counter++
	}
	if got, want := counter, 2; got != want {
		t.Errorf("got %d want %d iter items", got, want)
	}
	conversations.Compact()
	if got, want := conversations.Len(), 2; got != want {
		t.Errorf("got %d want %d len items", got, want)
	}

}

func TestConversationsKeep(t *testing.T) {

	file := "testdata/api-history-tennis.json"
	conversations, err := NewConversations(file)
	if err != nil {
		t.Fatal(err)
	}

	err = conversations.KeepItems([]int{0, 2})
	if err != nil {
		t.Fatalf("unexpected error for KeepItems(0, 2) %v", err)
	}

	counter := 0
	for _ = range conversations.Iter() {
		counter++
	}
	if got, want := counter, 0; got != want {
		t.Errorf("got %d want %d iter items", got, want)
	}
	conversations.Compact()
	if got, want := conversations.Len(), 2; got != want {
		t.Errorf("got %d want %d len items", got, want)
	}

}

func TestConversationsKeepAndReview(t *testing.T) {

	file := "testdata/api-history-tennis.json"
	conversations, err := NewConversations(file)
	if err != nil {
		t.Fatal(err)
	}

	err = conversations.ReviewItems([]int{-1})
	if err != nil {
		t.Fatalf("unexpected error for ReviewItems(-1) %v", err)
	}

	err = conversations.KeepItems([]int{0})
	if err != nil {
		t.Fatalf("unexpected error for KeepItems(0) %v", err)
	}

	counter := 0
	for c := range conversations.Iter() {
		_ = conversations.Keep(c.Idx) // keep the review item
		counter++
	}
	if got, want := counter, 1; got != want {
		t.Errorf("got %d want %d iter items", got, want)
	}

	conversations.Compact()

	if got, want := conversations.Len(), 2; got != want {
		t.Errorf("got %d want %d len items", got, want)
	}
	if diff := cmp.Diff(conversations.keep, []int{0, 3}); diff != "" {
		t.Errorf("conversations.keep mismatch (-got +want):\n%s", diff)
	}

}
