package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func newTestService(t *testing.T) (*Service, string) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "auth")
	return NewService(root), root
}

func TestService_CreateAndExists(t *testing.T) {
	svc, root := newTestService(t)

	if err := svc.Create("work"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	info, err := os.Stat(filepath.Join(root, "work", ".claude"))
	if err != nil {
		t.Fatalf("identity .claude dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("identity .claude path is not a directory")
	}

	exists, err := svc.Exists("work")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Exists(\"work\") = false, want true")
	}
}

func TestService_Create_RejectsInvalidName(t *testing.T) {
	svc, _ := newTestService(t)
	if err := svc.Create("Bad Name"); err == nil {
		t.Error("Create() expected error for invalid name, got nil")
	}
}

func TestService_Create_RejectsDuplicate(t *testing.T) {
	svc, _ := newTestService(t)
	if err := svc.Create("work"); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}
	if err := svc.Create("work"); err == nil {
		t.Error("second Create() expected error for duplicate, got nil")
	}
}

func TestService_List_SortedWithLoginState(t *testing.T) {
	svc, root := newTestService(t)
	if err := svc.Create("work"); err != nil {
		t.Fatalf("Create(work) error = %v", err)
	}
	if err := svc.Create("default"); err != nil {
		t.Fatalf("Create(default) error = %v", err)
	}
	// Mark "work" as logged in by dropping a credentials file.
	credPath := filepath.Join(root, "work", ".claude", ".credentials.json")
	if err := os.WriteFile(credPath, []byte("{}"), 0600); err != nil {
		t.Fatalf("write creds error = %v", err)
	}

	ids, err := svc.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("List() len = %d, want 2", len(ids))
	}
	if ids[0].Name != "default" || ids[1].Name != "work" {
		t.Errorf("List() order = [%q, %q], want [default, work]", ids[0].Name, ids[1].Name)
	}
	if ids[0].LoggedIn {
		t.Error("default LoggedIn = true, want false")
	}
	if !ids[1].LoggedIn {
		t.Error("work LoggedIn = false, want true")
	}
}

func TestService_List_EmptyWhenRootMissing(t *testing.T) {
	svc, _ := newTestService(t)
	ids, err := svc.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("List() len = %d, want 0", len(ids))
	}
}

func TestService_Remove(t *testing.T) {
	svc, _ := newTestService(t)
	if err := svc.Create("work"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := svc.Remove("work"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	exists, _ := svc.Exists("work")
	if exists {
		t.Error("identity still exists after Remove()")
	}
}

func TestService_Remove_RefusesDefault(t *testing.T) {
	svc, _ := newTestService(t)
	if err := svc.Create("default"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := svc.Remove("default"); err == nil {
		t.Error("Remove(\"default\") expected error, got nil")
	}
}

func TestService_Remove_ErrorsForMissing(t *testing.T) {
	svc, _ := newTestService(t)
	if err := svc.Remove("ghost"); err == nil {
		t.Error("Remove() expected error for missing identity, got nil")
	}
}

func TestService_EnsureDir(t *testing.T) {
	svc, root := newTestService(t)
	if err := svc.EnsureDir("work"); err != nil {
		t.Fatalf("EnsureDir() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "work", ".claude")); err != nil {
		t.Fatalf("EnsureDir did not create dir: %v", err)
	}
	// Idempotent.
	if err := svc.EnsureDir("work"); err != nil {
		t.Fatalf("second EnsureDir() error = %v", err)
	}
}

func TestService_DirFor(t *testing.T) {
	svc, root := newTestService(t)
	want := filepath.Join(root, "work", ".claude")
	if got := svc.DirFor("work"); got != want {
		t.Errorf("DirFor() = %q, want %q", got, want)
	}
}
