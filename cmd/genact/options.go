package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	flags "github.com/jessevdk/go-flags"
	"github.com/rorycl/genact"
)

const historyDir string = "conversations"

var usage string = fmt.Sprintf(`version %s

Have a conversation with gemini AI with the provided prompt file using
the settings file (by default at settings.yaml) and, optionally, either
a history file saved from previous AI discussions or downloaded from
Google AI studio. By default the last-generated history file will be
used, if it exists.

Within "Directory" (by default the current working directory) a
"conversations" directory will be made if it does not exist. For each
chat a further directory will be made, the "chat" directory. Within each
relevant "chat" directory a timestamped prompt and AI response (output)
will be saved, together with the previous history, if any, with the
prompt and response appended. An output.md file is also written from the
response to the current working directory for easy reference.

As a result, a call to this program for the first time along the
following lines:

	./genai -c hi prompt.txt

Will create something like the following output:

	./
	├── conversations
	│   └─── hi
	│       ├── 20250630T204842-history.json
	│       ├── 20250630T204842-output.md
	│       └── 20250630T204842-prompt.txt
	└── output.md

The 20250630T204842-history.json file can be used for the next call to
the api to "continue" the conversation, which is what will happen by
default if no apiHistory or studioHistory is specified.

./genact [-a apiHistory] [-s studioHistory] -c "chat name" \
         [-d directory] [-y yaml] `, genact.Version)

// CmdOptions are flag options which consume os.Args input.
type CmdOptions struct {
	APIHistory     string `short:"a" long:"apiHistory" description:"path to api history json file"`
	StudioHistory  string `short:"s" long:"studioHistory" description:"path to studio history json file"`
	Chat           string `short:"c" long:"chatName" description:"name of this conversation" required:"true"`
	Directory      string `short:"d" long:"directory" description:"directory" default:"current working directory"`
	YamlFile       string `short:"y" long:"yamlFile" description:"settings yaml file" default:"settings.yaml"`
	withoutHistory bool

	// paths
	conversationDirPath       string // path under Directory for conversations
	conversationDirPathExists bool
	chatDirPath               string // path under Directory/conversations for this chat
	chatDirPathExists         bool

	// prompt
	promptFile string // options.Args.Prompt is copied here
	// output
	Args struct {
		Prompt string `description:"prompt text file"`
	} `positional-args:"yes" required:"yes"`
}

// checkFileExists checks if a file exists
func checkFileExists(path string) bool {
	p, err := os.Stat(path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false
	}
	if p.IsDir() {
		return false
	}
	return true
}

// checkDirExists checks if a directory exists
func checkDirExists(path string) bool {
	p, err := os.Stat(path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false
	}
	if !p.IsDir() {
		return false
	}
	return true
}

// ParserError indicates a parser error
type ParserError struct {
	err error
}

func (p ParserError) Error() string {
	return fmt.Sprintf("%v", p.err)
}

// ParseOptions parses the command line options and returns a pointer to
// a ProgramOptions struct.
func ParseOptions() (*CmdOptions, error) {

	var options CmdOptions
	var parser = flags.NewParser(&options, flags.Default)
	parser.Usage = usage

	if _, err := parser.Parse(); err != nil {
		return nil, ParserError{err}
	}

	// history file checks
	if options.APIHistory != "" && options.StudioHistory != "" {
		return nil, errors.New("this program cannot accept both studio and api history files")
	}
	if options.APIHistory == "" && options.StudioHistory == "" {
		options.withoutHistory = true // set convenience flag
	}

	if options.APIHistory != "" && !checkFileExists(options.APIHistory) {
		return nil, fmt.Errorf("api history file %s could not be found", options.APIHistory)
	}
	if options.StudioHistory != "" && !checkFileExists(options.StudioHistory) {
		return nil, fmt.Errorf("api history file %s could not be found", options.StudioHistory)
	}

	// directory check
	if options.Directory == "" || options.Directory == "current working directory" {
		var err error
		options.Directory, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("could not get current working directory: %s", err)
		}
	} else {
		if !checkDirExists(options.Directory) {
			return nil, fmt.Errorf("could not find directory %s", options.Directory)
		}
	}

	// chat
	if options.Chat == "" {
		return nil, errors.New("chat name must be specified")
	}
	options.Chat = strings.ReplaceAll(strings.ToLower(strings.TrimSpace(options.Chat)), " ", "_")
	if strings.ContainsRune(options.Chat, filepath.Separator) {
		return nil, fmt.Errorf("chat name %s cannot contain path separator %q", options.Chat, filepath.Separator)
	}

	// prompt
	if options.Args.Prompt == "" || !checkFileExists(options.Args.Prompt) {
		return nil, fmt.Errorf("file '%s' could not be found", options.Args.Prompt)
	}
	options.promptFile = options.Args.Prompt

	// construct and check paths
	options.conversationDirPath = filepath.Clean(filepath.Join(options.Directory, historyDir))
	options.conversationDirPathExists = checkDirExists(options.conversationDirPath)
	options.chatDirPath = filepath.Clean(filepath.Join(options.conversationDirPath, options.Chat))
	options.chatDirPathExists = checkDirExists(options.chatDirPath)

	return &options, nil

}
