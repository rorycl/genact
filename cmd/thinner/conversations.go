package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/rorycl/genact"
)

// conversations are the components of a history file, normally the user
// conversing with the agent in user/agent pairs. Sometimes there might
// be more than one agent response, when one of the agent pairs is a
// "thinking" record.
//
// Conversations are a slice of Conversation arranging these
// conversation turns into a sequence to allow for `thinning` to exclude
// one ore more instance of a conversation from the sequence of
// conversations.
//
// The resulting Conversations slice can be serialized to a json
// "history" file for providing back to the genesis api for further
// conversation rounds.

// conversation is a single user/agent conversation.
type conversation struct {
	User     string
	Agent    []string
	idx      int
	agentLen int
}

// Conversations is a collection of user/agent conversation interactions
// in time order.
type Conversations struct {
	conversations []conversation
	keep          []int // indexes to keep
	compactRun    bool  // has index thinning occurred?
	reversed      bool  // has conversations been reversed?
}

// Reverse reverses the order of the conversations slice.
func (c *Conversations) Reverse() {
	slices.Reverse(c.conversations)
	c.reversed = !c.reversed
}

// Len returns the length of conversations
func (c *Conversations) Len() int {
	return len(c.conversations)
}

// Get gets a conversation by idx
func (c *Conversations) Get(idx int) conversation {
	if idx > len(c.conversations)-1 || idx < 0 {
		panic(fmt.Sprintf("conversation len %d, invalid index %d", len(c.conversations), idx))
	}
	if !c.reversed {
		return c.conversations[idx]
	}
	return c.conversations[len(c.conversations)-1-idx]
}

func (c *Conversations) Keep(idx int) error {
	if c.compactRun {
		return errors.New("compaction already run")
	}
	if idx > len(c.conversations)-1 {
		return fmt.Errorf(
			"index %d was larger than the conversation length %d",
			idx,
			len(c.conversations),
		)
	}
	for _, ix := range c.keep {
		if idx == ix {
			return nil // don't re-add any indexes
		}
	}
	c.keep = append(c.keep, idx)
	return nil
}

// Compact rewrites the conversations to only those that are in the keep
// index slice.
func (c *Conversations) Compact() {
	if c.compactRun {
		panic("compaction already run")
	}
	if c.reversed {
		c.Reverse()
		c.reversed = false
	}
	newConversations := []conversation{}
	slices.Sort(c.keep)
	for _, ix := range c.keep {
		newConversations = append(newConversations, c.conversations[ix])
	}
	c.conversations = newConversations
	c.compactRun = true
}

// Serialize serializes a conversations to json after conversion to a
// slice of APIJsonContent.
//
//	type APIJsonContent struct {
//		Role  string   `json:"Role"`
//		Parts []string `json:"Parts,omitempty"`
//	}
func (c *Conversations) Serialize() ([]byte, error) {
	ajc := []genact.APIJsonContent{}
	addContent := func(role string, parts []string) {
		ajc = append(ajc, APIJsonContent{role, parts})
	}
	for _, conv := range c.conversations {
		addContent("user", []string{conv.User})
		addContent("agent", conv.Agent)
	}

}

// NewConversations reads a json history API file from disk and converts
// it into a slice of Conversation.
func NewConversations(filePath string) (*Conversations, error) {
	history, err := genact.ReadAPIHistory(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}
	conversations := Conversations{}
	idx := 0
	conv := conversation{idx: idx}
	for _, h := range history {
		if h.Role == "user" {
			if conv.User != "" {
				conversations.conversations = append(conversations.conversations, conv)
				idx++
				conv = conversation{idx: idx}
			}
			conv.User = strings.Join(h.Parts, "\n\n--\n\n")
		} else { // agent
			conv.Agent = h.Parts
			conv.agentLen = len(h.Parts)
		}
	}
	if conv.User != "" {
		conversations.conversations = append(conversations.conversations, conv)
	}
	return &conversations, nil
}
