package main

import (
	"os"
	"testing"
)

// TestYaml reads settings from a settings file
func TestYaml(t *testing.T) {
	f, err := os.ReadFile("settings_example.yaml")
	if err != nil {
		t.Fatal(err)
	}
	y, err := LoadYaml(f)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := y["modelName"], "gemini-2.5-pro"; got != want {
		t.Errorf("got %s != want %s", got, want)
	}

	// fmt.Printf("%#v\n", y)
}
