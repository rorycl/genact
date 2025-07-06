package main

import (
	yaml "gopkg.in/yaml.v3"
)

type Settings map[string]string

// LoadYaml reads some settings into a Settings map.
func LoadYaml(yamlByte []byte) (Settings, error) {

	var settings Settings
	err := yaml.Unmarshal(yamlByte, &settings)
	return settings, err
}
