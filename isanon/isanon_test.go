package isanon

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsAnon(t *testing.T) {

	tests := []struct {
		source   string
		patterns []string
		err      error
	}{
		{
			source: `this is
				a chance this
				might be ok`,
			patterns: []string{`nomatch'`},
			err:      nil,
		},
		{
			source: `this is
				a chance this
				will fail`,
			patterns: []string{`w.*fail`},
			err:      InputLineMatch{2, `w.*fail`, "will fail"},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {

			err := IsStrAnonymous(tt.source, tt.patterns)
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
