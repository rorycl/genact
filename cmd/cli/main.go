// cli is a programme for working interactively with the Genesis 2.5 api
// to make the most of the large token window provided by Genesis.
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/rorycl/genact"
)

func main() {

	start := time.Now()

	// options
	options, err := ParseOptions()
	if err != nil {
		log.Fatal(err)
	}

	// settings
	f, err := os.ReadFile(options.YamlFile)
	if err != nil {
		log.Fatal(err)
	}
	settings, err := LoadYaml(f)
	if err != nil {
		log.Fatal(err)
	}

	// load history if required
	history := []*genai.Content{}
	if options.APIHistory != "" {
		history, err = genact.HistoryAPIToAIContent(options.APIHistory)
		if err != nil {
			log.Fatal(err)
		}
	} else if options.StudioHistory != "" {
		history, err = genact.HistoryStudioToAIContent(options.StudioHistory)
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
	response, err := genact.APIGetResponse(settings, history, string(prompt))
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
	err = files.WriteOutput([]byte(response.LatestResponse))
	if err != nil {
		log.Fatal(err)
	}
	err = files.WriteHistory([]byte(response.FullHistory))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("finished in %s, token count %d\n", time.Since(start), response.TokenCount)

}
