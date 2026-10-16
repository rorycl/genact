// package isanon determines if content in the provided string or in the attached files
// (if parseable) contain the provided regular expressions describing non-anonymised
// text.
package isanon

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/gen2brain/go-fitz"
)

// InputLineMatch is a custom error type for reporting anonymisation matches in input.
type InputLineMatch struct {
	source  string
	page    int
	line    int
	pattern string
	content string
}

func (i InputLineMatch) Error() string {
	return fmt.Sprintf("source %q page %d line %d matched %q: %s", i.source, i.page, i.line, i.pattern, i.content)
}

// errorFileTypeNotSupported reports if a provided file is supported by the underlying
// parser library (currently go-fitz).
var errorFileTypeNotSupported error = errors.New("file type not supported")

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
				return InputLineMatch{
					line:    i,
					pattern: c.String(),
					content: l,
				}
			}
		}
	}
	return nil
}

// checkFileAnonymous checks if the provided file is openable by the fitz/mupdf parsers
// and if it is, if the text therein includes any of the anonymisation regexes.
func checkFileAnonymous(filePath string, checkers []*regexp.Regexp) error {

	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%q does not exist", filePath)
	}

	doc, err := fitz.New(filePath)
	if err != nil {
		return errorFileTypeNotSupported
	}

	for page := 0; page < doc.NumPage(); page++ {
		text, err := doc.Text(page)
		if err != nil {
			return fmt.Errorf("fitz module file %q: page %d could not extract text: %v", filePath, page, err)
		}
		if err := checkStrAnonymous(text, checkers); err != nil {
			if e, ok := errors.AsType[InputLineMatch](err); ok {
				return InputLineMatch{
					source:  filePath,
					page:    page,
					line:    e.line,
					pattern: e.pattern,
					content: e.content,
				}
			}
			return err
		}
	}
	return nil
}

// IsAnonymous returns an error if the provided patterns are invalid regexps, or if
// the provided string contains any regexp matches, or if the contents of supported
// filetypes contains any regexp matches.
func IsAnonymous(input string, filePaths []string, patterns []string) error {

	checkers, err := compileRegexps(patterns)
	if err != nil {
		return err
	}
	if err := checkStrAnonymous(input, checkers); err != nil {
		if e, ok := errors.AsType[InputLineMatch](err); ok {
			return InputLineMatch{
				source:  "prompt", // default
				page:    0,
				line:    e.line,
				pattern: e.pattern,
				content: e.content,
			}
		}
		return err
	}
	for _, fp := range filePaths {
		err := checkFileAnonymous(fp, checkers)
		if errors.Is(errorFileTypeNotSupported, err) {
			// consider reporting to console
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}
