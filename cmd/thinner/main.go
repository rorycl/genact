package main

import (
	"fmt"
	"os"

	"github.com/rorycl/genact"
)

func question() int {
	var response string
    _, err := fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	switch {
		case strings.Contains(response, "y"):
		return 0
		case strings.Contains(response, "l"):
		return 1
		default:
		return -1
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("please provide a history file to parse")
		os.Exit(1)
	}
	history, err := genact.ReadAPIHistory(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	SavedHistory := []APIJsonContent

	var builder strings.Builder
	indexes := []int{}
	for i, entry := range history {
		if entry.Role == "user" {
			if builder.Len() > 0 {
				// show
			}

		fmt.Println("------------")
		fmt.Println(i, entry.Role)
		fmt.Println(entry.Parts)
	}

}
