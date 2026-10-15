package app

import (
	"fmt"
	"testing"
)

func TestSettings(t *testing.T) {

	filer := "../settings.example.yaml"

	s, err := LoadYaml(filer)
	if err != nil {
		t.Fatalf("could not load file %s: %v", filer, err)
	}

	if got, want := len(s.AnonymisationChecks), 2; got != want {
		fmt.Printf("%#v\n", s)
		t.Errorf("expected %d anonymisation checks, got %d", got, want)
	}

}
