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
			err:      InputLineMatch{"", 0, 2, `w.*fail`, "will fail"},
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

func TestFileIsAnon(t *testing.T) {

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
			err:      InputLineMatch{"testdata/test.docx", 0, 2, `w.*some.*data`, "will fail"},
		},
		{
			source:   "testdata/test.pdf",
			patterns: []*regexp.Regexp{regexp.MustCompile(`w.*some.*data`)},
			err:      InputLineMatch{"testdata/test.docx", 0, 2, `w.*some.*data`, "will fail"},
		},
		{
			source:   "testdata/test.pptx",
			patterns: []*regexp.Regexp{regexp.MustCompile(`A test presentation -- not`)},
			err:      nil,
		},
		{
			source:   "testdata/image.heic",
			patterns: []*regexp.Regexp{regexp.MustCompile(`some content`)},
			err:      errorFileTypeNotSupported,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {

			err := checkFileAnonymous(tt.source, tt.patterns)
			if err == nil && tt.err != nil {
				t.Fatalf("expected err: %v", tt.err)
			}
			if err != nil && tt.err == nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if err != nil {

				if errors.Is(err, errorFileTypeNotSupported) && errors.Is(tt.err, errorFileTypeNotSupported) {
					return
				}
				_, got := errors.AsType[InputLineMatch](err)
				_, want := errors.AsType[InputLineMatch](tt.err)
				if got != want {
					t.Errorf("got error type %T expected %T", err, tt.err)
				}
				// t.Logf("%T %#v\n", err, err)
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
			err:      InputLineMatch{"prompt", 0, 0, `h.*there`, "hi there"},
		},
		{
			prompt:   "hi there",
			patterns: []string{"no match"},
			err:      nil,
		},
		{
			prompt:   "hi there",
			patterns: []string{"no match"},
			files:    []string{"testdata/test.docx", "testdata/test.pdf"},
			err:      nil,
		},
		{
			prompt:   "hi there",
			patterns: []string{`(?i)a test presentation`},
			files:    []string{"testdata/test.docx", "testdata/test.pdf", "testdata/test.pptx"},
			err:      InputLineMatch{"testdata/test.pptx", 0, 0, `(?i)a test presentation`, "A test presentation"},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {

			err := IsAnonymous(tt.prompt, tt.files, tt.patterns)
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
				// t.Logf("%T %#v", err, err)
			}
		})
	}
}
