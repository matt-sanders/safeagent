package project

import (
	"os"
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

func TestService_ListByProfileID(t *testing.T) {
	svc := newTestService(t)

	svc.GetOrCreate("/home/user/app1")
	svc.GetOrCreate("/home/user/app2")
	svc.GetOrCreate("/home/user/app3")

	svc.SetProfile("/home/user/app1", "profile-a")
	svc.SetProfile("/home/user/app2", "profile-a")
	svc.SetProfile("/home/user/app3", "profile-b")

	projects, err := svc.ListByProfileID("profile-a")
	if err != nil {
		t.Fatalf("ListByProfileID() error = %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}

	projects, err = svc.ListByProfileID("nonexistent")
	if err != nil {
		t.Fatalf("ListByProfileID() error = %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestStore_Load_IgnoresLegacyContainerID(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "projects.json")

	// A legacy projects.json from before container_id was removed.
	legacy := `[{"id":"abc12345","path":"/home/user/myapp","profile_id":"prof1","container_id":"safeagent-abc12345"}]`
	if err := os.WriteFile(filePath, []byte(legacy), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	store := NewStore(filePath)
	projects, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].ID != "abc12345" {
		t.Errorf("ID = %q, want %q", projects[0].ID, "abc12345")
	}
	if projects[0].Path != "/home/user/myapp" {
		t.Errorf("Path = %q, want %q", projects[0].Path, "/home/user/myapp")
	}
	if projects[0].ProfileID != "prof1" {
		t.Errorf("ProfileID = %q, want %q", projects[0].ProfileID, "prof1")
	}
}

func TestStore_FindForCwd_ExactMatch(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(filepath.Join(tmpDir, "projects.json"))
	svc := NewService(store)

	_, _, err := svc.GetOrCreate("/home/user/myapp")
	if err != nil {
		t.Fatalf("GetOrCreate() error = %v", err)
	}

	p, found, err := store.FindForCwd("/home/user/myapp")
	if err != nil {
		t.Fatalf("FindForCwd() error = %v", err)
	}
	if !found {
		t.Fatal("expected found = true for exact match")
	}
	if p.Path != "/home/user/myapp" {
		t.Errorf("Path = %q, want %q", p.Path, "/home/user/myapp")
	}
}

func TestStore_FindForCwd_AncestorMatch(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(filepath.Join(tmpDir, "projects.json"))
	svc := NewService(store)

	_, _, err := svc.GetOrCreate("/home/user/myapp")
	if err != nil {
		t.Fatalf("GetOrCreate() error = %v", err)
	}

	p, found, err := store.FindForCwd("/home/user/myapp/worktree-1/sub")
	if err != nil {
		t.Fatalf("FindForCwd() error = %v", err)
	}
	if !found {
		t.Fatal("expected found = true for ancestor match")
	}
	if p.Path != "/home/user/myapp" {
		t.Errorf("Path = %q, want %q", p.Path, "/home/user/myapp")
	}
}

func TestStore_FindForCwd_PrefersDeepestAncestor(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(filepath.Join(tmpDir, "projects.json"))
	svc := NewService(store)

	_, _, err := svc.GetOrCreate("/home/user")
	if err != nil {
		t.Fatalf("GetOrCreate() error = %v", err)
	}
	_, _, err = svc.GetOrCreate("/home/user/myapp")
	if err != nil {
		t.Fatalf("GetOrCreate() error = %v", err)
	}

	p, found, err := store.FindForCwd("/home/user/myapp/worktree-1")
	if err != nil {
		t.Fatalf("FindForCwd() error = %v", err)
	}
	if !found {
		t.Fatal("expected found = true")
	}
	if p.Path != "/home/user/myapp" {
		t.Errorf("Path = %q, want %q (deepest ancestor)", p.Path, "/home/user/myapp")
	}
}

func TestStore_FindForCwd_NoMatch(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(filepath.Join(tmpDir, "projects.json"))

	_, found, err := store.FindForCwd("/some/random/path")
	if err != nil {
		t.Fatalf("FindForCwd() error = %v", err)
	}
	if found {
		t.Error("expected found = false for empty store")
	}
}

func TestStore_FindForCwd_StopsAtRoot(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(filepath.Join(tmpDir, "projects.json"))
	svc := NewService(store)

	// Project at a sibling, not on the path from cwd to /
	_, _, err := svc.GetOrCreate("/elsewhere/myapp")
	if err != nil {
		t.Fatalf("GetOrCreate() error = %v", err)
	}

	_, found, err := store.FindForCwd("/home/user/myapp")
	if err != nil {
		t.Fatalf("FindForCwd() error = %v", err)
	}
	if found {
		t.Error("expected found = false when project is not an ancestor")
	}
}

func TestService_FindOrCreateForCwd_ReturnsExistingExact(t *testing.T) {
	svc := newTestService(t)

	_, _, err := svc.GetOrCreate("/home/user/myapp")
	if err != nil {
		t.Fatalf("GetOrCreate() error = %v", err)
	}

	p, created, err := svc.FindOrCreateForCwd("/home/user/myapp")
	if err != nil {
		t.Fatalf("FindOrCreateForCwd() error = %v", err)
	}
	if created {
		t.Error("expected created = false for existing project")
	}
	if p.Path != "/home/user/myapp" {
		t.Errorf("Path = %q, want %q", p.Path, "/home/user/myapp")
	}
}

func TestService_FindOrCreateForCwd_ReturnsAncestor(t *testing.T) {
	svc := newTestService(t)

	_, _, err := svc.GetOrCreate("/home/user/myapp")
	if err != nil {
		t.Fatalf("GetOrCreate() error = %v", err)
	}

	p, created, err := svc.FindOrCreateForCwd("/home/user/myapp/worktree-1")
	if err != nil {
		t.Fatalf("FindOrCreateForCwd() error = %v", err)
	}
	if created {
		t.Error("expected created = false when ancestor exists")
	}
	if p.Path != "/home/user/myapp" {
		t.Errorf("Path = %q, want %q (ancestor)", p.Path, "/home/user/myapp")
	}
}

func TestService_FindOrCreateForCwd_CreatesWhenNoAncestor(t *testing.T) {
	svc := newTestService(t)

	p, created, err := svc.FindOrCreateForCwd("/home/user/newapp")
	if err != nil {
		t.Fatalf("FindOrCreateForCwd() error = %v", err)
	}
	if !created {
		t.Error("expected created = true for new project")
	}
	if p.Path != "/home/user/newapp" {
		t.Errorf("Path = %q, want %q", p.Path, "/home/user/newapp")
	}
}
