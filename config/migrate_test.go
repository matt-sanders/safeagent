package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateAuthLayout_MovesLegacyClaude(t *testing.T) {
	configDir := t.TempDir()
	legacy := filepath.Join(configDir, ".claude")
	if err := os.MkdirAll(legacy, 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	// A marker file proves the contents were moved, not recreated.
	if err := os.WriteFile(filepath.Join(legacy, "marker"), []byte("x"), 0600); err != nil {
		t.Fatalf("setup marker: %v", err)
	}

	if err := MigrateAuthLayout(configDir); err != nil {
		t.Fatalf("MigrateAuthLayout() error = %v", err)
	}

	moved := filepath.Join(configDir, "auth", "default", ".claude", "marker")
	if _, err := os.Stat(moved); err != nil {
		t.Fatalf("legacy contents not moved to default identity: %v", err)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Errorf("legacy .claude still exists after migration (err=%v)", err)
	}
}

func TestMigrateAuthLayout_FreshInstallCreatesDefault(t *testing.T) {
	configDir := t.TempDir()

	if err := MigrateAuthLayout(configDir); err != nil {
		t.Fatalf("MigrateAuthLayout() error = %v", err)
	}

	defaultClaude := filepath.Join(configDir, "auth", "default", ".claude")
	info, err := os.Stat(defaultClaude)
	if err != nil {
		t.Fatalf("default identity not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("default .claude is not a directory")
	}
}

func TestMigrateAuthLayout_IdempotentWhenAuthExists(t *testing.T) {
	configDir := t.TempDir()
	// Pre-existing auth layout with a marker; a legacy .claude also present.
	existing := filepath.Join(configDir, "auth", "default", ".claude")
	if err := os.MkdirAll(existing, 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(existing, "keep"), []byte("x"), 0600); err != nil {
		t.Fatalf("setup keep: %v", err)
	}
	legacy := filepath.Join(configDir, ".claude")
	if err := os.MkdirAll(legacy, 0700); err != nil {
		t.Fatalf("setup legacy: %v", err)
	}

	if err := MigrateAuthLayout(configDir); err != nil {
		t.Fatalf("MigrateAuthLayout() error = %v", err)
	}

	// Existing default identity is untouched...
	if _, err := os.Stat(filepath.Join(existing, "keep")); err != nil {
		t.Errorf("existing default identity was disturbed: %v", err)
	}
	// ...and the legacy .claude is left alone (not moved) since auth/ existed.
	if _, err := os.Stat(legacy); err != nil {
		t.Errorf("legacy .claude should be left untouched when auth/ exists: %v", err)
	}
}
