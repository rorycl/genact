package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var glowPath string

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
	if len(os.Args) != 2 {
		fmt.Println("please provide a history file to parse")
		os.Exit(1)
	}

	// find "glow"
	var err error
	glowPath, err = exec.LookPath("glow")
	if err != nil {
		fmt.Printf("glow could not be found: %v", err)
	}

	// load conversation
	conversations, err := NewConversations(os.Args[1])
	if err != nil {
		fmt.Printf("could not load history file: %v", err)
		os.Exit(1)
	}

	// reverse the file
	conversations.Reverse()

	// iterate over conversations
	for c := range conversations.Iter() {
		content := fmt.Sprint(c)
		runCommand(content)
		if question() {
			conversations.Keep(c.Idx)
		}
	}

	// compact the conversations
	conversations.Compact()

	// save to file
	tf, err := os.CreateTemp("", "genact_conv_*.json")
	if err != nil {
		fmt.Printf("temporary file creation error %v\n", err)
		os.Exit(1)
	}

	output, err := conversations.Serialize()
	if err != nil {
		fmt.Printf("serialization error: %v\n", err)
		os.Exit(1)
	}
	_, err = tf.Write(output)
	if err != nil {
		fmt.Println(err)
	}

	_ = tf.Close()

}
