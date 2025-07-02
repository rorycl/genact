package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/generative-ai-go/genai"
)

func main() {

	start := time.Now()

	// settings
	f, err := os.ReadFile("settings_example.yaml")
	if err != nil {
		log.Fatal(err)
	}
	settings, err := LoadYaml(f)
	if err != nil {
		log.Fatal(err)
	}

	// options
	options, err := ParseOptions()
	if err != nil {
		log.Fatal(err)
	}

	// load history if required
	history := []*genai.Content{}
	if options.APIHistory != "" {
		history, err := HistoryAPIToAIContent(options.APIHistory)
		if err != nil {
			log.Fatal(err)
		}
	} else if options.StudioHistory != "" {
		history, err := HistoryStudioToAIContent(options.StudioHistory)
		if err != nil {
			log.Fatal(err)
		}
	}

	// load prompt
	prompt, err := os.ReadFile(options.promptFile)
	if err != nil {
		log.Fatal(err)
	}

	// run api
	response, err := APIGetResponse(settings, history, string(prompt))
	if err != nil {
		log.Fatal(err)
	}

	// setup files
	files, err := NewFiles(options.Directory, options.Chat)
	if err != nil {
		log.Fatal(err)
	}
	err = files.WritePrompt(prompt)
	if err != nil {
		log.Fatal(err)
	}
	err = files.WriteOutput([]byte(response.latestResponse))
	if err != nil {
		log.Fatal(err)
	}
	err = files.WriteHistory([]byte(response.fullHistory))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("finished in %s, token count %d\n", t.Now().Sub(start), response.tokenCount)

}
