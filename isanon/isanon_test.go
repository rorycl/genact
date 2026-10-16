package isanon

import (
	"errors"
	"fmt"
	"regexp"
	"testing"
)

func TestStrIsAnon(t *testing.T) {

	tests := []struct {
		source   string
		input    string
		patterns []string
		err      error
	}{
		{
			input: `this is
				a chance this
				might be ok`,
			patterns: []string{`nomatch'`},
			err:      nil,
		},
		{
			input: `this is
				a chance this
				will fail`,
			patterns: []string{`w.*fail`},
			err:      InputLineMatch{"", 2, `w.*fail`, "will fail"},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {

			checkers, err := compileRegexps(tt.patterns)
			if err != nil {
				t.Error(err)
				return
			}

			err = checkStrAnonymous(tt.input, checkers)
			if err == nil && tt.err != nil {
				t.Fatalf("expected err: %v", tt.err)
			}
			if err != nil && tt.err == nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if err != nil {
				_, got := errors.AsType[InputLineMatch](err)
				_, want := errors.AsType[InputLineMatch](tt.err)
				if got != want {
					t.Errorf("got error type %T expected %T", err, tt.err)
				}
				// fmt.Printf("%T %#v", err, err)
			}
		})
	}
}

func TestIsAnonDoc(t *testing.T) {

	tests := []struct {
		source   string
		patterns []*regexp.Regexp
		err      error
	}{
		{
			source:   "testdata/test.docx",
			patterns: []*regexp.Regexp{regexp.MustCompile(`nomatch`)},
			err:      nil,
		},
		{
			source:   "testdata/test.docx",
			patterns: []*regexp.Regexp{regexp.MustCompile(`w.*some.*data`)},
			err:      InputLineMatch{"testdata/test.docx", 2, `w.*some.*data`, "will fail"},
		},
	}

	// fmt.Println(checkWordFileAnonymous("testdata/test.docx", nil))

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {

			err := checkWordFileAnonymous(tt.source, tt.patterns)
			if err == nil && tt.err != nil {
				t.Fatalf("expected err: %v", tt.err)
			}
			if err != nil && tt.err == nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if err != nil {
				_, got := errors.AsType[InputLineMatch](err)
				_, want := errors.AsType[InputLineMatch](tt.err)
				if got != want {
					t.Errorf("got error type %T expected %T", err, tt.err)
				}
				fmt.Printf("%T %#v", err, err)
			}
		})
	}
}

func TestIsAnonPDF(t *testing.T) {

	tests := []struct {
		source   string
		patterns []*regexp.Regexp
		err      error
	}{
		{
			source:   "testdata/test.pdf",
			patterns: []*regexp.Regexp{regexp.MustCompile(`nomatch`)},
			err:      nil,
		},
		{
			source:   "testdata/test.pdf",
			patterns: []*regexp.Regexp{regexp.MustCompile(`w.*some.*data`)},
			err:      InputLineMatch{"testdata/test.pdf", 2, `w.*some.*data`, "will fail"},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {

			err := checkPDFFileAnonymous(tt.source, tt.patterns)
			if err == nil && tt.err != nil {
				t.Fatalf("expected err: %v", tt.err)
			}
			if err != nil && tt.err == nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if err != nil {
				_, got := errors.AsType[InputLineMatch](err)
				_, want := errors.AsType[InputLineMatch](tt.err)
				if got != want {
					t.Errorf("got error type %T expected %T", err, tt.err)
				}
				fmt.Printf("%T %#v", err, err)
			}
		})
	}
}

func TestIsAnon(t *testing.T) {

	tests := []struct {
		prompt   string
		patterns []string
		files    []string
		err      error
	}{
		{
			prompt:   "hi there",
			patterns: []string{"no match"},
			err:      nil,
		},
		{
			prompt:   "hi there",
			patterns: []string{"h.*there"},
			err:      InputLineMatch{"prompt", 0, `h.*there`, "hi there"},
		},
		{
			prompt:   "hi there",
			patterns: []string{"no match"},
			err:      nil,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {

			err := IsAnonymous("prompt", tt.prompt, tt.files, tt.patterns)
			if err == nil && tt.err != nil {
				t.Fatalf("expected err: %v", tt.err)
			}
			if err != nil && tt.err == nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if err != nil {
				_, got := errors.AsType[InputLineMatch](err)
				_, want := errors.AsType[InputLineMatch](tt.err)
				if got != want {
					t.Errorf("got error type %T expected %T", err, tt.err)
				}
				fmt.Printf("%T %#v", err, err)
			}
		})
	}
}
