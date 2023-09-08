package distribution

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Application struct {
	ID          string `yaml:"id"`
	LongID      string `yaml:"long_id"`
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Contact     string `yaml:"contact"`
	MainDir     string `yaml:"main"`
	IconFile    string `yaml:"icon"`
	Copyright   string `yaml:"copyright"`
}

func LoadAppDescriptor(fileLoc string) (*Application, error) {
	file, err := os.Open(fileLoc)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	var app Application
	if err := yaml.NewDecoder(file).Decode(&app); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	return &app, nil
}
