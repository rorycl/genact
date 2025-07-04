package main

import (
	"fmt"
	"strings"
)

// question asks the user for a yes/no answer.
func question() bool {
	var response string
	for {
		fmt.Print("keep selection [Y/n]: ")
		_, err := fmt.Scanln(&response)
		if err == nil {
			break
		}
	}

	response = strings.ToLower(strings.TrimSpace(response))
	switch {
	case strings.ContainsAny(response, "yY"):
		return true
	/*
		case strings.ContainsAny(response, "qQ"):
				return -1
	*/
	default:
		return false
	}
}
