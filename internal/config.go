package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetScheduleConfigPath returns the XDG compliant path for schedule.json.
// It creates the necessary directory structure if it doesn't exist.
func GetScheduleConfigPath() (string, error) {
	var configHome string

	// 1. Check XDG_CONFIG_HOME environment variable
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		configHome = xdgConfigHome
	} else {
		// 2. Fallback to ~/.config
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		configHome = filepath.Join(homeDir, ".config")
	}

	appConfigDir := filepath.Join(configHome, "radikoRecScheduler")
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create application config directory '%s': %w", appConfigDir, err)
	}

	return filepath.Join(appConfigDir, "schedule.json"), nil
}
