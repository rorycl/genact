package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/rorycl/genact"
)

// glowPath is the path to the "glow" binary.
// See https://github.com/charmbracelet/glow
var glowPath string

// run glow with pager with a display width of 180 chars
func runCommand(s string) {
	cmd := exec.Command(glowPath, "-w", "180", "-p")
	cmd.Stdin = strings.NewReader(s)
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Println("command error: ", err)
	}
}

func main() {

	var (
		originalConversationLen  int
		compactedConversationLen int
	)

	// Parse the command line options.
	options, err := ParseOptions()
	if err != nil {
		var pe ParserError
		if !errors.As(err, &pe) {
			log.Fatal(err)
		}
		os.Exit(1)
	}

	// Find the "glow" binary.
	glowPath, err = exec.LookPath("glow")
	if err != nil {
		fmt.Printf("glow could not be found: %v'n", err)
		os.Exit(1)
	}

	// Load the conversation from history.
	conversations, err := genact.NewConversations(options.inputFile)
	if err != nil {
		fmt.Printf("could not load history file: %v\n", err)
		os.Exit(1)
	}
	originalConversationLen = conversations.Len()
	if originalConversationLen == 0 {
		fmt.Println("no conversations found")
		os.Exit(0)
	}

	// Reverse the file.
	conversations.Reverse()

	// Initialise review items, if any.
	err = conversations.ReviewItems(options.Review)
	if err != nil {
		fmt.Printf("review items failed: %v\n", err)
		os.Exit(1)
	}

	// Initialise keep items, if any.
	err = conversations.KeepItems(options.Keep)
	if err != nil {
		fmt.Printf("keep items failed: %v\n", err)
		os.Exit(1)
	}

	// Iterate over conversations.
	for c := range conversations.Iter() {
		content := fmt.Sprint(c) // convert to string representation
		runCommand(content)
		if question() {
			err := conversations.Keep(c.Idx)
			if err != nil {
				fmt.Printf("conversation index %d could not be kept: %v\n", c.Idx, err)
			}
		}
	}

	// Compact the conversations to remove unwanted items.
	conversations.Compact()
	compactedConversationLen = conversations.Len()
	if compactedConversationLen == 0 {
		fmt.Println("No conversations left after compaction")
		os.Exit(0)
	}
	if compactedConversationLen == originalConversationLen {
		fmt.Println("Compacted and original conversation length are the same. Aborting.")
		os.Exit(0)
	}

	// Serialize the compacted conversation to json.
	output, err := conversations.Serialize()
	if err != nil {
		fmt.Printf("serialization error: %v\n", err)
		os.Exit(1)
	}

	// Write the serialized information to file.
	_, err = options.output.Write(output)
	if err != nil {
		fmt.Println(err)
	}
	_ = options.output.Close()

	// Inform the user.
	fmt.Printf(
		"Original conversations of %d items successfully compacted to %d items.\n",
		originalConversationLen,
		compactedConversationLen,
	)
}
