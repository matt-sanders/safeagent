package project

import (
	"path/filepath"
	"testing"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	tmpDir := t.TempDir()
	store := NewStore(filepath.Join(tmpDir, "projects.json"))
	return NewService(store)
}

func TestService_GetOrCreate_CreatesNewProject(t *testing.T) {
	svc := newTestService(t)

	p, created, err := svc.GetOrCreate("/home/user/myapp")
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

func TestService_GetOrCreate_ReturnsExisting(t *testing.T) {
	svc := newTestService(t)

	_, _, err := svc.GetOrCreate("/home/user/myapp")
	if err != nil {
		t.Fatalf("first GetOrCreate() error = %v", err)
	}

	p, created, err := svc.GetOrCreate("/home/user/myapp")
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

func TestService_SetProfile(t *testing.T) {
	svc := newTestService(t)

	_, _, err := svc.GetOrCreate("/home/user/myapp")
	if err != nil {
		t.Fatalf("GetOrCreate() error = %v", err)
	}

	err = svc.SetProfile("/home/user/myapp", "abc123")
	if err != nil {
		t.Fatalf("SetProfile() error = %v", err)
	}

	p, _, err := svc.GetOrCreate("/home/user/myapp")
	if err != nil {
		t.Fatalf("GetOrCreate() after SetProfile error = %v", err)
	}
	if p.ProfileID != "abc123" {
		t.Errorf("ProfileID = %q, want %q", p.ProfileID, "abc123")
	}
}

func TestService_SetProfile_ErrorsForUnknown(t *testing.T) {
	svc := newTestService(t)

	err := svc.SetProfile("/nonexistent", "abc123")
	if err == nil {
		t.Error("SetProfile() expected error for unknown project, got nil")
	}
}

func TestService_GetOrCreate_MultipleProjects(t *testing.T) {
	svc := newTestService(t)

	_, _, _ = svc.GetOrCreate("/home/user/app1")
	_, _, _ = svc.GetOrCreate("/home/user/app2")

	// Verify both exist by getting them again
	_, created1, _ := svc.GetOrCreate("/home/user/app1")
	_, created2, _ := svc.GetOrCreate("/home/user/app2")

	if created1 {
		t.Error("app1 should already exist")
	}
	if created2 {
		t.Error("app2 should already exist")
	}
}
