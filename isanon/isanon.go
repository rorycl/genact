package isanon

import (
	"fmt"
	"regexp"
	"strings"
)

// InputLineMatch is a custom error type for reporting anonymisation matches in input.
type InputLineMatch struct {
	line    int
	pattern string
	content string
}

func (i InputLineMatch) Error() string {
	return fmt.Sprintf("line %d matched %q: %s", i.line, i.pattern, i.content)
}

// compileRegexps compiles regular expressions from the provided patterns.
func compileRegexps(patterns []string) ([]*regexp.Regexp, error) {
	rgs := []*regexp.Regexp{}
	for _, p := range patterns {
		r, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("pattern %q could not compile: %s", p, err)
		}
		rgs = append(rgs, r)
	}
	return rgs, nil
}

// checkStrAnonymous checks if the provided string data matches any of the anonymisation
// regexps. The function splits the input by line.
func checkStrAnonymous(input string, checkers []*regexp.Regexp) error {
	lines := strings.Split(input, "\n")
	for i, l := range lines {
		for _, c := range checkers {
			if c.MatchString(l) {
				return InputLineMatch{i, c.String(), l}
			}
		}
	}
	return nil
}

// IsStrAnonymous returns an error if the provided patterns are invalid regexps, or if
// the provided string contains
func IsStrAnonymous(input string, patterns []string) error {

	checkers, err := compileRegexps(patterns)
	if err != nil {
		return err
	}
	return checkStrAnonymous(input, checkers)
}
