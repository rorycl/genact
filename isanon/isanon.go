package isanon

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dslipak/pdf"
	"github.com/fumiama/go-docx"
)

// InputLineMatch is a custom error type for reporting anonymisation matches in input.
type InputLineMatch struct {
	source  string
	line    int
	pattern string
	content string
}

func (i InputLineMatch) Error() string {
	return fmt.Sprintf("source %q line %d matched %q: %s", i.source, i.line, i.pattern, i.content)
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
				return InputLineMatch{"", i, c.String(), l}
			}
		}
	}
	return nil
}

// checkWordFileAnonymous checks if the provided Word file includes any of the
// anonymisation regexps.
func checkWordFileAnonymous(filePath string, checkers []*regexp.Regexp) error {
	fh, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("word file %q open problem %v", filePath, err)
	}
	defer func() {
		_ = fh.Close()
	}()

	fileinfo, err := fh.Stat()
	if err != nil {
		return fmt.Errorf("word file %q stat error %v", filePath, err)
	}
	size := fileinfo.Size()
	doc, err := docx.Parse(fh, size)
	if err != nil {
		return fmt.Errorf("word file %q stat error %v", filePath, err)
	}
	for i, it := range doc.Document.Body.Items {
		switch it.(type) {
		case *docx.Paragraph, *docx.Table: // printable
			content := fmt.Sprintf("%s", it)
			for _, c := range checkers {
				if c.MatchString(content) {
					return InputLineMatch{"", i, c.String(), content}
				}
			}
		}
	}
	return nil
}

// checkPDFFileAnonymous checks if the provided PDF includes any of the anonymisation
// regexps. This PDF parser is very basic and concatenates strings across lines, losing
// the newline, which will break many word boundary-based regexps.
func checkPDFFileAnonymous(filePath string, checkers []*regexp.Regexp) error {

	r, err := pdf.Open(filePath)
	if err != nil {
		return fmt.Errorf("pdf file %q open problem %v", filePath, err)
	}

	b, err := r.GetPlainText()
	if err != nil {
		return fmt.Errorf("pdf file %q get plain text error %v", filePath, err)
	}

	var buf bytes.Buffer
	_, err = buf.ReadFrom(b)
	if err != nil {
		return fmt.Errorf("pdf file %q buffer read error %v", filePath, err)
	}
	fmt.Println(buf.String())

	return checkStrAnonymous(buf.String(), checkers)

}

// routeFileTypeToChecker routes files by extension to a checker
func routeFileTypeToChecker(filePath string, checkers []*regexp.Regexp) error {

	ext := filepath.Ext(filePath)
	if ext == "" {
		return nil
	}
	switch strings.ToLower(ext) {
	case "doc", "docx":
		return checkWordFileAnonymous(filePath, checkers)
	case "pdf":
		return checkPDFFileAnonymous(filePath, checkers)
	}
	return nil
}

// IsAnonymous returns an error if the provided patterns are invalid regexps, or if
// the provided string contains any regexp matches, or if the contents of supported
// filetypes contains any regexp matches.
func IsAnonymous(source, input string, filePaths []string, patterns []string) error {

	checkers, err := compileRegexps(patterns)
	if err != nil {
		return err
	}
	if err := checkStrAnonymous(input, checkers); err != nil {
		if e, ok := errors.AsType[InputLineMatch](err); ok {
			return InputLineMatch{"prompt", e.line, e.pattern, e.content}
		}
		return err
	}
	for _, fp := range filePaths {
		if err := routeFileTypeToChecker(fp, checkers); err != nil {
			if e, ok := errors.AsType[InputLineMatch](err); ok {
				return InputLineMatch{fp, e.line, e.pattern, e.content}
			}
			return err
		}
	}
	return nil
}
