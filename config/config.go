package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const configDirName = ".safe-claude"

// EnsureConfigDir ensures ~/.safe-claude exists inside the given home directory.
// It returns the full path to the config directory.
func EnsureConfigDir(homeDir string) (string, error) {
	configDir := filepath.Join(homeDir, configDirName)

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}

	return configDir, nil
}
