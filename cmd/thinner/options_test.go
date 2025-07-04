package main

import (
	"os"
	"testing"
)

func TestOptions(t *testing.T) {

	testFile, err := os.CreateTemp("", "thin_cmd_*")
	if err != nil {
		t.Fatal(err)
	}
	testFileName := testFile.Name()
	_ = os.Remove(testFileName)

	outputFile, err := os.CreateTemp("", "thin_cmd2_*")
	if err != nil {
		t.Fatal(err)
	}
	outputFileName := outputFile.Name()
	defer func() {
		_ = os.Remove(outputFileName)
	}()

	tests := []struct {
		desc  string
		args  []string
		isErr bool
	}{
		{
			desc:  "missing output",
			args:  []string{"prog", "../../testdata/api-history-tennis.json"},
			isErr: true,
		},
		{
			desc:  "missing input",
			args:  []string{"prog", "-o", outputFileName, testFileName},
			isErr: true,
		},
		{
			desc:  "ok",
			args:  []string{"prog", "-o", outputFileName, "../../testdata/api-history-tennis.json"},
			isErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			os.Args = tt.args
			_, err := ParseOptions()
			if got, want := (err != nil), tt.isErr; got != want {
				if err != nil {
					t.Fatalf("got unexpected error %s", err)
				}
				if err == nil {
					t.Fatalf("expected error")
				}
			}
		})
	}
}
