package main

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
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
		desc   string
		args   []string
		review []int
		isErr  bool
	}{
		{
			desc:   "missing output",
			args:   []string{"prog", "../../testdata/api-history-tennis.json"},
			review: nil,
			isErr:  true,
		},
		{
			desc:   "missing input",
			args:   []string{"prog", "-o", outputFileName, testFileName},
			review: nil,
			isErr:  true,
		},
		{
			desc:   "ok",
			args:   []string{"prog", "-o", outputFileName, "../../testdata/api-history-tennis.json"},
			review: nil,
			isErr:  false,
		},
		{
			desc:   "ok with review",
			args:   []string{"prog", "-o", outputFileName, "-r", "3", "-r", "4", "../../testdata/api-history-tennis.json"},
			review: []int{3, 4},
			isErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			os.Args = tt.args
			po, err := ParseOptions()
			if got, want := (err != nil), tt.isErr; got != want {
				if err != nil {
					t.Fatalf("got unexpected error %s", err)
				}
				if err == nil {
					t.Fatalf("expected error")
				}
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(po.Review, tt.review); diff != "" {
				t.Errorf("review mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
