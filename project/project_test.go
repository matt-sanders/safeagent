package project

import (
	"path/filepath"
	"testing"
)

func TestGetOrCreate_CreatesNewProject(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "projects.json")

	p, created, err := GetOrCreate(filePath, "/home/user/myapp")
	if err != nil {
		t.Fatalf("GetOrCreate() error = %v", err)
	}
	if !created {
		t.Error("expected created = true for new project")
	}
	if p.Path != "/home/user/myapp" {
		t.Errorf("Path = %q, want %q", p.Path, "/home/user/myapp")
	}
	if p.ProfileID != "" {
		t.Errorf("ProfileID = %q, want empty", p.ProfileID)
	}
}

func TestGetOrCreate_ReturnsExistingProject(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "projects.json")

	_, _, err := GetOrCreate(filePath, "/home/user/myapp")
	if err != nil {
		t.Fatalf("first GetOrCreate() error = %v", err)
	}

	p, created, err := GetOrCreate(filePath, "/home/user/myapp")
	if err != nil {
		t.Fatalf("second GetOrCreate() error = %v", err)
	}
	if created {
		t.Error("expected created = false for existing project")
	}
	if p.Path != "/home/user/myapp" {
		t.Errorf("Path = %q, want %q", p.Path, "/home/user/myapp")
	}
}

func TestSetProfile_UpdatesProject(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "projects.json")

	_, _, err := GetOrCreate(filePath, "/home/user/myapp")
	if err != nil {
		t.Fatalf("GetOrCreate() error = %v", err)
	}

	err = SetProfile(filePath, "/home/user/myapp", "abc123")
	if err != nil {
		t.Fatalf("SetProfile() error = %v", err)
	}

	p, _, err := GetOrCreate(filePath, "/home/user/myapp")
	if err != nil {
		t.Fatalf("GetOrCreate() after SetProfile error = %v", err)
	}
	if p.ProfileID != "abc123" {
		t.Errorf("ProfileID = %q, want %q", p.ProfileID, "abc123")
	}
}

func TestSetProfile_ErrorsForUnknownProject(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "projects.json")

	err := SetProfile(filePath, "/nonexistent", "abc123")
	if err == nil {
		t.Error("SetProfile() expected error for unknown project, got nil")
	}
}

func TestGetOrCreate_MultipleProjects(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "projects.json")

	_, _, err := GetOrCreate(filePath, "/home/user/app1")
	if err != nil {
		t.Fatalf("first GetOrCreate() error = %v", err)
	}
	_, _, err = GetOrCreate(filePath, "/home/user/app2")
	if err != nil {
		t.Fatalf("second GetOrCreate() error = %v", err)
	}

	projects, err := Load(filePath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
}
