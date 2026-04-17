package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAdd_CreatesFileAndProfile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "profiles.json")

	p := Profile{Name: "default", NodeVersion: "22"}
	err := Add(filePath, p)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	profiles, err := Load(filePath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].Name != "default" {
		t.Errorf("Name = %q, want %q", profiles[0].Name, "default")
	}
	if profiles[0].NodeVersion != "22" {
		t.Errorf("NodeVersion = %q, want %q", profiles[0].NodeVersion, "22")
	}
	if profiles[0].ID == "" {
		t.Error("ID is empty, want non-empty")
	}
	if len(profiles[0].ID) != 8 {
		t.Errorf("ID length = %d, want 8", len(profiles[0].ID))
	}
}

func TestAdd_AppendsToExisting(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "profiles.json")

	err := Add(filePath, Profile{Name: "first", NodeVersion: "20"})
	if err != nil {
		t.Fatalf("Add() first error = %v", err)
	}
	err = Add(filePath, Profile{Name: "second", NodeVersion: "22"})
	if err != nil {
		t.Fatalf("Add() second error = %v", err)
	}

	profiles, err := Load(filePath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}
	if profiles[0].ID == profiles[1].ID {
		t.Errorf("expected different IDs, both are %q", profiles[0].ID)
	}
}

func TestAdd_GeneratesUniqueIDs(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "profiles.json")

	names := []string{"a", "b", "c", "d", "e"}
	for _, name := range names {
		if err := Add(filePath, Profile{Name: name, NodeVersion: "22"}); err != nil {
			t.Fatalf("Add() error for %q: %v", name, err)
		}
	}

	profiles, err := Load(filePath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	seen := make(map[string]bool)
	for _, p := range profiles {
		if seen[p.ID] {
			t.Errorf("duplicate ID %q found", p.ID)
		}
		seen[p.ID] = true
	}
}

func TestAdd_RejectsDuplicateName(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "profiles.json")

	err := Add(filePath, Profile{Name: "dupe", NodeVersion: "20"})
	if err != nil {
		t.Fatalf("Add() first error = %v", err)
	}
	err = Add(filePath, Profile{Name: "dupe", NodeVersion: "22"})
	if err == nil {
		t.Error("Add() expected error for duplicate name, got nil")
	}
}

func TestLoad_ReturnsEmptyForMissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "profiles.json")

	profiles, err := Load(filePath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(profiles))
	}
}

func TestLoad_ReturnsErrorForCorruptFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "profiles.json")

	os.WriteFile(filePath, []byte("not json"), 0600)

	_, err := Load(filePath)
	if err == nil {
		t.Error("Load() expected error for corrupt file, got nil")
	}
}
