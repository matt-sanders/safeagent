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
	if p.ID == "" {
		t.Error("ID is empty, want non-empty")
	}
	if len(p.ID) != 8 {
		t.Errorf("ID length = %d, want 8", len(p.ID))
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

func TestService_AddExclusion(t *testing.T) {
	svc := newTestService(t)
	svc.GetOrCreate("/home/user/myapp")

	err := svc.AddExclusion("/home/user/myapp", "node_modules")
	if err != nil {
		t.Fatalf("AddExclusion() error = %v", err)
	}

	exclusions, err := svc.GetExclusions("/home/user/myapp")
	if err != nil {
		t.Fatalf("GetExclusions() error = %v", err)
	}
	if len(exclusions) != 1 {
		t.Fatalf("expected 1 exclusion, got %d", len(exclusions))
	}
	if exclusions[0] != "node_modules" {
		t.Errorf("exclusion = %q, want %q", exclusions[0], "node_modules")
	}
}

func TestService_AddExclusion_RejectsDuplicate(t *testing.T) {
	svc := newTestService(t)
	svc.GetOrCreate("/home/user/myapp")

	svc.AddExclusion("/home/user/myapp", "node_modules")
	err := svc.AddExclusion("/home/user/myapp", "node_modules")
	if err == nil {
		t.Error("AddExclusion() expected error for duplicate, got nil")
	}
}

func TestService_RemoveExclusion(t *testing.T) {
	svc := newTestService(t)
	svc.GetOrCreate("/home/user/myapp")

	svc.AddExclusion("/home/user/myapp", "node_modules")
	svc.AddExclusion("/home/user/myapp", ".env")

	err := svc.RemoveExclusion("/home/user/myapp", "node_modules")
	if err != nil {
		t.Fatalf("RemoveExclusion() error = %v", err)
	}

	exclusions, _ := svc.GetExclusions("/home/user/myapp")
	if len(exclusions) != 1 {
		t.Fatalf("expected 1 exclusion, got %d", len(exclusions))
	}
	if exclusions[0] != ".env" {
		t.Errorf("remaining exclusion = %q, want %q", exclusions[0], ".env")
	}
}

func TestService_RemoveExclusion_ErrorsForMissing(t *testing.T) {
	svc := newTestService(t)
	svc.GetOrCreate("/home/user/myapp")

	err := svc.RemoveExclusion("/home/user/myapp", "nonexistent")
	if err == nil {
		t.Error("RemoveExclusion() expected error for missing exclusion, got nil")
	}
}

func TestService_GetExclusions_EmptyByDefault(t *testing.T) {
	svc := newTestService(t)
	svc.GetOrCreate("/home/user/myapp")

	exclusions, err := svc.GetExclusions("/home/user/myapp")
	if err != nil {
		t.Fatalf("GetExclusions() error = %v", err)
	}
	if len(exclusions) != 0 {
		t.Errorf("expected 0 exclusions, got %d", len(exclusions))
	}
}

func TestService_SetContainerID(t *testing.T) {
	svc := newTestService(t)

	_, _, err := svc.GetOrCreate("/home/user/myapp")
	if err != nil {
		t.Fatalf("GetOrCreate() error = %v", err)
	}

	err = svc.SetContainerID("/home/user/myapp", "safe-claude-abc123")
	if err != nil {
		t.Fatalf("SetContainerID() error = %v", err)
	}

	p, _, err := svc.GetOrCreate("/home/user/myapp")
	if err != nil {
		t.Fatalf("GetOrCreate() after SetContainerID error = %v", err)
	}
	if p.ContainerID != "safe-claude-abc123" {
		t.Errorf("ContainerID = %q, want %q", p.ContainerID, "safe-claude-abc123")
	}
}

func TestService_SetContainerID_ErrorsForUnknown(t *testing.T) {
	svc := newTestService(t)

	err := svc.SetContainerID("/nonexistent", "safe-claude-abc123")
	if err == nil {
		t.Error("SetContainerID() expected error for unknown project, got nil")
	}
}
