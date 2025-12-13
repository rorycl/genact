package app

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v3"
)

// Settings controls the API interactions.
type Settings struct {
	OutputFile    string `yaml:"outputFile"`    // Default output file name
	ModelName     string `yaml:"modelName"`     // e.g. gemini-2.0-flash-exp, gemini-3.0-pro
	APIKey        string `yaml:"apiKey"`        // Google Cloud API Key
	Logging       bool   `yaml:"logging"`       // Enable verbose logging
	ThinkingLevel string `yaml:"thinkingLevel"` // "low" or "high"
}

// DefaultSettings returns safe defaults.
func DefaultSettings() Settings {
	return Settings{
		OutputFile:    "output.md",
		ModelName:     "gemini-2.5-pro",
		Logging:       true,
		ThinkingLevel: "high",
	}
}

// LoadYaml reads settings from a YAML file.
func LoadYaml(path string) (Settings, error) {
	settings := DefaultSettings()

	f, err := os.ReadFile(path)
	if err != nil {
		// If file doesn't exist, we rely on env vars or defaults
		if os.IsNotExist(err) {
			if key := os.Getenv("GOOGLE_API_KEY"); key != "" {
				settings.APIKey = key
				return settings, nil
			}
		}
		return settings, fmt.Errorf("could not read settings file %s: %w", path, err)
	}

	err = yaml.Unmarshal(f, &settings)
	if err != nil {
		return settings, fmt.Errorf("could not parse settings yaml: %w", err)
	}

	if settings.APIKey == "" {
		settings.APIKey = os.Getenv("GOOGLE_API_KEY")
		if settings.APIKey == "" {
			return settings, fmt.Errorf("apiKey not found in settings file or GOOGLE_API_KEY env var")
		}
	}

	return settings, nil
}
