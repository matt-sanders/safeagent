package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	tmpDir := t.TempDir()
	store := NewStore(filepath.Join(tmpDir, "profiles.json"))
	return NewService(store)
}

func TestService_Create(t *testing.T) {
	svc := newTestService(t)

	p, err := svc.Create("default", "22", "")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if p.Name != "default" {
		t.Errorf("Name = %q, want %q", p.Name, "default")
	}
	if p.NodeVersion != "22" {
		t.Errorf("NodeVersion = %q, want %q", p.NodeVersion, "22")
	}
	if p.ID == "" {
		t.Error("ID is empty, want non-empty")
	}
	if len(p.ID) != 8 {
		t.Errorf("ID length = %d, want 8", len(p.ID))
	}
}

func TestService_Create_RejectsDuplicate(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.Create("dupe", "20", "")
	if err != nil {
		t.Fatalf("Create() first error = %v", err)
	}
	_, err = svc.Create("dupe", "22", "")
	if err == nil {
		t.Error("Create() expected error for duplicate name, got nil")
	}
}

func TestService_List(t *testing.T) {
	svc := newTestService(t)

	profiles, err := svc.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(profiles))
	}

	_, _ = svc.Create("first", "20", "")
	_, _ = svc.Create("second", "22", "")

	profiles, err = svc.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(profiles))
	}
}

func TestService_GetByID(t *testing.T) {
	svc := newTestService(t)

	created, _ := svc.Create("default", "22", "")

	p, err := svc.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if p.Name != "default" {
		t.Errorf("Name = %q, want %q", p.Name, "default")
	}
}

func TestService_GetByID_NotFound(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.GetByID("nonexistent")
	if err == nil {
		t.Error("GetByID() expected error for missing ID, got nil")
	}
}

func TestService_List_ErrorForCorruptFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "profiles.json")
	os.WriteFile(filePath, []byte("not json"), 0600)

	svc := NewService(NewStore(filePath))
	_, err := svc.List()
	if err == nil {
		t.Error("List() expected error for corrupt file, got nil")
	}
}

func TestService_Create_UniqueIDs(t *testing.T) {
	svc := newTestService(t)

	names := []string{"a", "b", "c", "d", "e"}
	for _, name := range names {
		if _, err := svc.Create(name, "22", ""); err != nil {
			t.Fatalf("Create() error for %q: %v", name, err)
		}
	}

	profiles, _ := svc.List()
	seen := make(map[string]bool)
	for _, p := range profiles {
		if seen[p.ID] {
			t.Errorf("duplicate ID %q found", p.ID)
		}
		seen[p.ID] = true
	}
}
