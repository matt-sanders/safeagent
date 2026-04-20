package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureConfigDir_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".safeagent")

	got, err := EnsureConfigDir(tmpDir)
	if err != nil {
		t.Fatalf("EnsureConfigDir() error = %v", err)
	}

	if got != configDir {
		t.Errorf("EnsureConfigDir() = %q, want %q", got, configDir)
	}

	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("config dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("config path is not a directory")
	}
	if info.Mode().Perm() != 0700 {
		t.Errorf("config dir permissions = %o, want 0700", info.Mode().Perm())
	}
}

func TestEnsureConfigDir_ExistingDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".safeagent")

	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}

	got, err := EnsureConfigDir(tmpDir)
	if err != nil {
		t.Fatalf("EnsureConfigDir() error = %v", err)
	}
	if got != configDir {
		t.Errorf("EnsureConfigDir() = %q, want %q", got, configDir)
	}
}

func TestEnsureConfigDir_FailsOnUnwritablePath(t *testing.T) {
	_, err := EnsureConfigDir("/proc/nonexistent")
	if err == nil {
		t.Error("EnsureConfigDir() expected error for unwritable path, got nil")
	}
}
