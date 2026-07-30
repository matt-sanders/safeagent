package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// MigrateAuthLayout ensures the per-identity auth layout exists under configDir.
// It is idempotent: if <configDir>/auth already exists it does nothing.
// Otherwise it creates <configDir>/auth and either moves a pre-existing
// <configDir>/.claude into <configDir>/auth/default/.claude, or creates an
// empty default identity when there is nothing to migrate.
func MigrateAuthLayout(configDir string) error {
	authDir := filepath.Join(configDir, "auth")
	if _, err := os.Stat(authDir); err == nil {
		return nil // already migrated
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to inspect auth directory: %w", err)
	}

	defaultClaude := filepath.Join(authDir, "default", ".claude")
	legacyClaude := filepath.Join(configDir, ".claude")

	if _, err := os.Stat(legacyClaude); err == nil {
		if err := os.MkdirAll(filepath.Join(authDir, "default"), 0700); err != nil {
			return fmt.Errorf("failed to create default identity: %w", err)
		}
		if err := os.Rename(legacyClaude, defaultClaude); err != nil {
			return fmt.Errorf("failed to migrate legacy .claude: %w", err)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to inspect legacy .claude: %w", err)
	}

	if err := os.MkdirAll(defaultClaude, 0700); err != nil {
		return fmt.Errorf("failed to create default identity: %w", err)
	}
	return nil
}
