package genact

import (
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"slices"
	"strings"
)

// conversation is a record of single user/agent ("model") conversation
// interaction.
type conversation struct {
	User     string
	Model    []string
	Idx      int
	agentLen int
}

// String is a string representation of a conversation, suitable for
// piping to a markdown reader such as "glow".
func (co conversation) String() string {
	output := "\n`user`:\n\n"
	output += co.User
	output += "\n\n---\n\n`agent`:\n\n"
	output += strings.Join(co.Model, "\n\n---\n\n")
	return output
}

// Conversations is a collection of user/agent (or model) conversation
// interactions recorded in time order.
//
// Conversations are the components of a genact history file, normally
// the user conversing with the agent in user/model pairs. Sometimes
// there might be more than one agent response, when agent's responses
// includes a "thinking" record.
//
// Conversations are a slice of Conversation arranging these
// conversation turns into a sequence to allow for `thinning` to exclude
// one or more instances of a conversation from the sequence of
// conversations to reduce token size and processing cost.
//
// The resulting Conversations slice can be serialized to a json
// "history" file for providing back to the genesis api for further
// conversation rounds.
type Conversations struct {
	conversations []conversation
	keep          []int        // indexes to keep
	compactRun    bool         // has index thinning occurred?
	reversed      bool         // has conversations been reversed?
	itemsToReview map[int]bool // only review these items
	itemsToKeep   map[int]bool // preset items to keep
}

// Reverse reverses the order of the conversations slice.
func (c *Conversations) Reverse() {
	slices.Reverse(c.conversations)
	c.reversed = !c.reversed
}

// Len returns the length of conversations.
func (c *Conversations) Len() int {
	return len(c.conversations)
}

// ReviewItems sets the list of conversations to review by idx. Items
// with a negative index are replaced with the relevant item from the
// end of the slice of conversation. ReviewItems are only relevant to
// the `Iter` main review iterator which sets a keep index, used for
// compacting the conversations using the `Compact` method.
func (c *Conversations) ReviewItems(ri []int) error {
	var err error
	c.itemsToReview, err = c.makeMapOfItems(ri)
	return err
}

// KeepItems sets the list of conversations to keep by idx,
// short-circuiting the main review process provided by `Iter`. Notes
// otherwise are as for ReviewItems.
func (c *Conversations) KeepItems(ki []int) error {
	var err error
	c.itemsToKeep, err = c.makeMapOfItems(ki)
	return err
}

// makeMapOfItems makes a map of valid conversation items.
func (c *Conversations) makeMapOfItems(is []int) (map[int]bool, error) {
	thisMap := map[int]bool{}
	for _, idx := range is {
		if idx < 0 {
			i := len(c.conversations) + idx
			if i < 0 {
				return nil, fmt.Errorf("item %d out of range len %d", idx, len(c.conversations))
			}
			thisMap[i] = true
			continue
		}
		if idx > len(c.conversations)-1 {
			return nil, fmt.Errorf("item %d out of range len %d", idx, len(c.conversations))
		}
		thisMap[idx] = true
	}
	return thisMap, nil
}

// Iter iterates over the sequence of conversation, altered by the
// itemsToReview and itemsToKeep maps if set, to allow the user to
// review each conversation in turn to decide if it is to be included
// (using `Keep`) or discarded from the resulting history file.

// KeepItems simply sets the items wanted for compaction without
// yielding any conversations to the user.
//
// Either the full sequence of conversation or the KeepItems of
// conversation used in conjunction with ReviewItems will only yield the
// ReviewItems.
//
// Iter is normally applied to a reversed conversation (using Reverse)
// so that older material can be removed with reference to the latest
// information, rather than the other way around, at the user's
// preference.
func (c *Conversations) Iter() iter.Seq[conversation] {

	reviewOK := func(idx int) bool {
		if _, ok := c.itemsToReview[idx]; ok {
			return true
		}
		return false
	}
	keepOK := func(idx int) bool {
		if _, ok := c.itemsToKeep[idx]; ok {
			return true
		}
		return false
	}
	return func(yield func(conversation) bool) {
		for _, conv := range c.conversations {

			switch {
			// Return each item if itemsToKeep and itemsToReview are
			// empty.
			case len(c.itemsToKeep) == 0 && len(c.itemsToReview) == 0:
				// Show everything.

			// If there are no items to keep and some to review, keep
			// everything in conversations but only show those to
			// review.
			case len(c.itemsToKeep) == 0 && len(c.itemsToReview) > 0:
				if !reviewOK(conv.Idx) {
					_ = c.Keep(conv.Idx)
					continue
				}
				// Show only items to review.

			// If there are items to keep and none to review, keep the
			// former but don't show any.
			case len(c.itemsToKeep) > 0 && len(c.itemsToReview) == 0:
				if keepOK(conv.Idx) {
					_ = c.Keep(conv.Idx)
					continue
				}
				continue
				// Don't show any items.

			// If there are both items to keep and items to review, keep
			// those in the former but always show those to review.
			case len(c.itemsToKeep) > 0 && len(c.itemsToReview) > 0:
				if !reviewOK(conv.Idx) && !keepOK(conv.Idx) {
					continue
				}
				if !reviewOK(conv.Idx) && keepOK(conv.Idx) {
					_ = c.Keep(conv.Idx)
					continue
				}
				// Show only review items.

			}
			// Show the conversation.
			if !yield(conv) {
				return
			}
		}
	}
}

// Get gets a conversation from the conversations slice.
func (c *Conversations) Get(idx int) conversation {
	if idx > len(c.conversations)-1 || idx < 0 {
		panic(fmt.Sprintf("conversation len %d, invalid index %d", len(c.conversations), idx))
	}
	if !c.reversed {
		return c.conversations[idx]
	}
	return c.conversations[len(c.conversations)-1-idx]
}

// Keep stores a conversation turn for keeping on compaction.
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

// Compact rewrites the conversations to only include those conversation
// turns selected by the user by using Keep (potentially after
// interactive review) or KeepItems.
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

// Serialize serializes conversations to json after conversion to a
// slice of APIJsonContent.
func (c *Conversations) Serialize() ([]byte, error) {
	ajc := []APIJsonContent{}
	addContent := func(role string, parts []string) {
		ajc = append(ajc, APIJsonContent{role, parts})
	}
	for _, conv := range c.conversations {
		addContent("user", []string{conv.User})
		addContent("model", conv.Model)
	}
	return json.MarshalIndent(ajc, "", "  ")
}

// NewConversations reads a json history API file from disk and converts
// it into a slice of Conversation.
func NewConversations(filePath string) (*Conversations, error) {
	history, err := ReadAPIHistory(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}
	conversations := Conversations{}
	idx := 0
	conv := conversation{Idx: idx}
	for _, h := range history {
		if h.Role == "user" {
			if conv.User != "" {
				conversations.conversations = append(conversations.conversations, conv)
				idx++
				conv = conversation{Idx: idx}
			}
			conv.User = strings.Join(h.Parts, "\n\n--\n\n")
		} else { // agent|model
			conv.Model = h.Parts
			conv.agentLen = len(h.Parts)
		}
	}
	if conv.User != "" {
		conversations.conversations = append(conversations.conversations, conv)
	}
	return &conversations, nil
}
