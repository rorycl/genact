package main

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

func chkFileExists(path string) bool {
	p, err := os.Stat(path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false
	}
	if p.IsDir() {
		return false
	}
	return true
}

func TestFiles(t *testing.T) {

	var err error
	var files *files

	tmpDir, err := os.MkdirTemp("", "genai_tmpfilesdir_*")
	if err != nil {
		t.Fatalf("could not make temporary directory: %s", err)
	}
	defer func() {
		if err == nil {
			return
		}
		_ = os.RemoveAll(tmpDir)
	}()

	files, err = NewFiles(tmpDir, "chat1")

	b := []byte("hi there")

	err = files.WritePrompt(b)
	if err != nil {
		t.Fatalf("writePrompt error %s", err)
	}

	err = files.WriteOutput(b)
	if err != nil {
		t.Fatalf("writePrompt error %s", err)
	}

	err = files.WriteHistory(b)
	if err != nil {
		t.Fatalf("writePrompt error %s", err)
	}

	for _, f := range []string{
		files.outputFile,
		files.chatPromptFile,
		files.chatHistoryFile,
		files.chatOutputFile,
	} {
		fmt.Println(f)
		if !chkFileExists(f) {
			// stop removal of tmpDir
			err = fmt.Errorf("file %s could not be found", f)
			t.Error(err)
		}
	}
}
