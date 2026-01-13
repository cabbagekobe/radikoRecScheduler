package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// Config holds the application's configuration.
type Config struct {
	RadigoCommandPath string `json:"radigo_command_path"`
}

// DefaultConfig provides default values for the configuration.
var DefaultConfig = Config{
	RadigoCommandPath: "radigo", // Default to "radigo" assuming it's in PATH
}

var AppConfig Config

// LoadConfig loads the configuration from the given file path or uses defaults.
func LoadConfig(configFilePath string) {
	file, err := os.ReadFile(configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			AppConfig = DefaultConfig
			log.Printf("Config file '%s' not found. Using default configuration and creating it.", configFilePath)
			if err := SaveConfig(configFilePath, AppConfig); err != nil {
				log.Printf("Warning: Could not save default config file: %v\n", err)
			}
			return
		}
		log.Fatalf("Error reading config file '%s': %v", configFilePath, err)
	}

	if err := json.Unmarshal(file, &AppConfig); err != nil {
		log.Fatalf("Error parsing JSON from config file '%s': %v", configFilePath, err)
	}
	log.Printf("Configuration loaded from '%s'.\n", configFilePath)
}

// SaveConfig writes the current configuration to the specified file path.
func SaveConfig(filePath string, config Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling config to JSON: %w", err)
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("error writing config file '%s': %w", filePath, err)
	}
	return nil
}
