package auth

import (
	"testing"
)

func TestService_LoadConfig_EmptyWhenMissing(t *testing.T) {
	svc, _ := newTestService(t)
	if err := svc.Create("work"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	cfg, err := svc.LoadConfig("work")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if len(cfg.Mounts) != 0 {
		t.Errorf("LoadConfig() mounts = %v, want empty", cfg.Mounts)
	}
}

func TestService_AddMount(t *testing.T) {
	svc, _ := newTestService(t)
	if err := svc.Create("work"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := svc.AddMount("work", "/host/gh", "/home/developer/.config/gh"); err != nil {
		t.Fatalf("AddMount() error = %v", err)
	}

	cfg, err := svc.LoadConfig("work")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if len(cfg.Mounts) != 1 {
		t.Fatalf("LoadConfig() mounts len = %d, want 1", len(cfg.Mounts))
	}
	if cfg.Mounts[0].Source != "/host/gh" || cfg.Mounts[0].Target != "/home/developer/.config/gh" {
		t.Errorf("mount = %+v, want {/host/gh /home/developer/.config/gh}", cfg.Mounts[0])
	}
}

func TestService_AddMount_RejectsDuplicateTarget(t *testing.T) {
	svc, _ := newTestService(t)
	if err := svc.Create("work"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := svc.AddMount("work", "/host/a", "/container/x"); err != nil {
		t.Fatalf("first AddMount() error = %v", err)
	}
	if err := svc.AddMount("work", "/host/b", "/container/x"); err == nil {
		t.Error("second AddMount() expected error for duplicate target, got nil")
	}
}

func TestService_RemoveMount(t *testing.T) {
	svc, _ := newTestService(t)
	if err := svc.Create("work"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := svc.AddMount("work", "/host/gh", "/container/gh"); err != nil {
		t.Fatalf("AddMount() error = %v", err)
	}
	if err := svc.RemoveMount("work", "/container/gh"); err != nil {
		t.Fatalf("RemoveMount() error = %v", err)
	}
	cfg, err := svc.LoadConfig("work")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if len(cfg.Mounts) != 0 {
		t.Errorf("after RemoveMount, mounts = %v, want empty", cfg.Mounts)
	}
}

func TestService_RemoveMount_ErrorsForMissing(t *testing.T) {
	svc, _ := newTestService(t)
	if err := svc.Create("work"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := svc.RemoveMount("work", "/nonexistent"); err == nil {
		t.Error("RemoveMount() expected error for missing target, got nil")
	}
}
