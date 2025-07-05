package main

import (
	"fmt"
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

	options, err := ParseOptions()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// find "glow"
	glowPath, err = exec.LookPath("glow")
	if err != nil {
		fmt.Printf("glow could not be found: %v", err)
		os.Exit(1)
	}

	// load conversation
	conversations, err := genact.NewConversations(options.inputFile)
	if err != nil {
		fmt.Printf("could not load history file: %v", err)
		os.Exit(1)
	}

	// reverse the file
	conversations.Reverse()

	// initialise review items, if any
	conversations.ReviewItems(options.Review)

	// iterate over conversations
	for c := range conversations.Iter() {
		content := fmt.Sprint(c) // convert to string representation
		runCommand(content)
		if question() {
			conversations.Keep(c.Idx)
		}
	}

	// compact the conversations
	conversations.Compact()

	// serialize output to json
	output, err := conversations.Serialize()
	if err != nil {
		fmt.Printf("serialization error: %v\n", err)
		os.Exit(1)
	}

	// write to file
	_, err = options.output.Write(output)
	if err != nil {
		fmt.Println(err)
	}
	_ = options.output.Close()
}
