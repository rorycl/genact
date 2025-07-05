package main

import (
	"errors"
	"fmt"
	"os"

	flags "github.com/jessevdk/go-flags"
)

const (
	Version string = "0.0.1" // program version
)

var usage string = fmt.Sprintf(`[-o outputFile] [-r 1, -r 3...] historyFile.json

version %s

Interactively "thin" a gemini history file saved with genact by choosing
which conversations from the history to output to a new history json file.

Note that the conversations are replayed in reverse order, but
recompiled in the original order.

This uses bubbletea's "glow" markdown pager programme, which needs to be
on your PATH.

Using the -r/--review flag only reviews the (0-indexed) conversations
numbered, the other indexed items are kept. Negative indexing can be
used to refer to items from the end of the list of conversations, so -1
means the last item.
`, Version)

// CmdOptions are flag options which consume os.Args input.
type CmdOptions struct {
	OutputFile string `short:"o" long:"outputFile" required:"true" description:"file path to save output"`
	Review     []int  `short:"r" long:"review" description:"list of specific conversation pairs to review"`
	output     *os.File
	inputFile  string
	Args       struct {
		InputFile string `description:"input history json file"`
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

	if !checkFileExists(options.Args.InputFile) {
		return nil, fmt.Errorf("history file %s does not exist", options.Args.InputFile)
	}
	options.inputFile = options.Args.InputFile

	var err error
	options.output, err = os.Create(options.OutputFile)
	if err != nil {
		return nil, fmt.Errorf("could not create output file %s: %v ", options.OutputFile, err)
	}

	return &options, nil
}
