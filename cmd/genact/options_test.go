package main

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

func TestOptions(t *testing.T) {

	tmpDir, err := os.MkdirTemp("", "genai_tmpdir_*")
	if err != nil {
		t.Fatalf("could not make temporary directory: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	t.Chdir("../..")

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get current working directory: %s", err)
	}

	tests := []struct {
		desc              string
		args              []string
		isErr             bool
		dir               string
		chatDirPathExists bool
	}{
		{
			desc:              "simple invocation no error",
			args:              []string{"prog", "-c", "chat1", "-d", "testdata/optionsdir3", "testdata/optionsdir3/prompt.txt"},
			isErr:             false,
			dir:               "testdata/optionsdir3",
			chatDirPathExists: true,
		},
		{
			desc:              "simple invocation error with no chat",
			args:              []string{"prog", "-d", "testdata/optionsdir3", "testdata/optionsdir3/prompt.txt"},
			isErr:             true,
			chatDirPathExists: false, // not really needed
		},
		{
			desc:              "simple invocation error with no prompt",
			args:              []string{"prog", "-c", "chat1", "-d", "testdata/optionsdir3"},
			isErr:             true,
			chatDirPathExists: false,
		},
		{
			desc:              "simple invocation error with unknown directory",
			args:              []string{"prog", "-c", "chat1", "-d", "testdata/unknown", "testdata/optionsdir3/prompt.txt"},
			isErr:             true,
			chatDirPathExists: false,
		},
		{
			desc:              "invocation with history api history file given",
			args:              []string{"prog", "-c", "chat9", "-a", "testdata/optionsdir3/apihistory.json", "-d", "testdata/optionsdir3", "testdata/optionsdir3/prompt.txt"},
			isErr:             false,
			dir:               "testdata/optionsdir3",
			chatDirPathExists: false,
		},
		{
			desc:              "invocation with history studio history file given",
			args:              []string{"prog", "-c", "chat1", "-s", "testdata/optionsdir3/studiohistory.json", "-d", "testdata/optionsdir3", "testdata/optionsdir3/prompt.txt"},
			isErr:             false,
			dir:               "testdata/optionsdir3",
			chatDirPathExists: true,
		},
		{
			desc:              "invocation with history error with clashing history files",
			args:              []string{"prog", "-c", "chat1", "-a", "testdata/optionsdir3/apihistory.json", "-s", "testdata/optionsdir3/studiohistory.json", "-d", "testdata/optionsdir3", "testdata/optionsdir3/prompt.txt"},
			isErr:             true,
			chatDirPathExists: false,
		},
		{
			desc:              "invocation with no conversation dir",
			args:              []string{"prog", "-c", "chat1", "-d", tmpDir, "testdata/optionsdir3/prompt.txt"},
			isErr:             false,
			dir:               tmpDir,
			chatDirPathExists: false,
		},
		{
			desc:              "invocation with current working directory",
			args:              []string{"prog", "-c", "chat1", "testdata/optionsdir3/prompt.txt"},
			isErr:             false,
			dir:               cwd,
			chatDirPathExists: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			os.Args = tt.args
			cmdOptions, err := ParseOptions()
			if got, want := (err != nil), tt.isErr; got != want {
				if err != nil {
					t.Fatalf("got unexpected error %s", err)
				}
				if err == nil {
					t.Fatalf("expected error")
				}
			}
			// For verbose test runs show the error.
			// Native flag errors are printed to stdout; don't print those twice.
			if tt.isErr {
				var pe ParserError
				if !errors.As(err, &pe) {
					fmt.Println(err)
				}
				return
			}
			if got, want := cmdOptions.Directory, tt.dir; got != want {
				t.Errorf("directory path got %s want %s", got, want)
			}
			if got, want := cmdOptions.chatDirPathExists, tt.chatDirPathExists; got != want {
				t.Errorf("chatDirPathExists got %t want %t", got, want)
			}
			// fmt.Printf("%#v\n", cmdOptions)
		})
	}
}
