package runner

import (
	"fmt"
	"os"

	"github.com/e1l1ya/findtarget/pkg/types"
	"gopkg.in/yaml.v3"
)

// LoadTemplate loads the YAML configuration from the user-specified file path.
func LoadTemplate(templatePath string) (*types.Config, error) {
	// If no template path is provided, return an error
	if templatePath == "" {
		return nil, fmt.Errorf("no template path provided")
	}

	// Check if the provided template file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", templatePath)
	}

	// Read and parse the YAML file
	configData, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config types.Config
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	return &config, nil
}
