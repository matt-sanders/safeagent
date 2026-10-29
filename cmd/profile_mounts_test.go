package cmd_test

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"safeagent/docker"
)

func TestProfileMountsAddThenLsShowsTheMount(t *testing.T) {
	env := setup(t)
	host := t.TempDir()

	out := env.MustRun("profile", "mounts", "add", "work", "--host", host, "--container", "/home/developer/.config/gh")

	if !strings.Contains(out, "/home/developer/.config/gh") {
		t.Fatalf("expected add to confirm the mount, got:\n%s", out)
	}
	ls := env.MustRun("profile", "mounts", "ls", "work")
	lines := strings.Split(strings.TrimSpace(ls), "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[0], "HOST PATH") || !strings.Contains(lines[0], "CONTAINER PATH") {
		t.Fatalf("expected a header and one row, got:\n%s", ls)
	}
	if f := strings.Fields(lines[1]); len(f) != 2 || f[0] != host || f[1] != "/home/developer/.config/gh" {
		t.Fatalf("unexpected row %q", lines[1])
	}
}

func TestProfileMountsAddRejectsInvalidMounts(t *testing.T) {
	env := setup(t)
	host := t.TempDir()
	env.MustRun("profile", "mounts", "add", "work", "--host", host, "--container", "/home/developer/.ssh")

	cases := []struct {
		name, host, container, want string
	}{
		{"missing host path", filepath.Join(host, "nope"), "/home/developer/.config/gh", "does not exist"},
		{"relative container path", host, "home/developer/.config/gh", "must be absolute"},
		{"claude config path", host, "/claude", "reserved"},
		{"duplicate container path", host, "/home/developer/.ssh/", "already has a mount"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := env.Run("profile", "mounts", "add", "work", "--host", tc.host, "--container", tc.container)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected an error containing %q, got %v", tc.want, err)
			}
		})
	}
	if got := strings.Count(env.MustRun("profile", "mounts", "ls", "work"), "\n"); got != 2 {
		t.Fatalf("expected only the first mount to be stored, got %d lines", got)
	}
}

func TestProfileMountsAddShowsTheProfileFolder(t *testing.T) {
	env := setup(t)

	out := env.MustRun("profile", "mounts", "add", "work", "--host", t.TempDir(), "--container", "/x")

	want := filepath.Join(env.ConfigRoot, "profiles", profileID(t, env, "work"))
	if !strings.Contains(out, want) {
		t.Fatalf("expected the profile folder %s in the output, got:\n%s", want, out)
	}
}

func TestProfileMountsRmRemovesByContainerPath(t *testing.T) {
	env := setup(t)
	env.MustRun("profile", "mounts", "add", "work", "--host", t.TempDir(), "--container", "/home/developer/.config/gh")

	env.MustRun("profile", "mounts", "rm", "work", "/home/developer/.config/gh")

	if out := env.MustRun("profile", "mounts", "ls", "work"); strings.TrimSpace(out) != "No mounts." {
		t.Fatalf("expected no mounts, got:\n%s", out)
	}
	_, err := env.Run("profile", "mounts", "rm", "work", "/home/developer/.config/gh")
	if err == nil || !strings.Contains(err.Error(), "no mount at") {
		t.Fatalf("expected an unknown-mount error, got %v", err)
	}
}

func TestProfileLsCountsMounts(t *testing.T) {
	env := setup(t)
	env.MustRun("profile", "mounts", "add", "work", "--host", t.TempDir(), "--container", "/a")
	env.MustRun("profile", "mounts", "add", "work", "--host", t.TempDir(), "--container", "/b")

	out := env.MustRun("profile", "ls")

	for _, line := range strings.Split(out, "\n") {
		if f := strings.Fields(line); len(f) == 4 && f[0] == "work" {
			if f[2] != "2" {
				t.Fatalf("expected MOUNTS 2, got %q", line)
			}
			return
		}
	}
	t.Fatalf("no work row in:\n%s", out)
}

func TestStartIncludesProfileMounts(t *testing.T) {
	env := setup(t)
	host := t.TempDir()
	env.MustRun("profile", "mounts", "add", "work", "--host", host, "--container", "/home/developer/.config/gh")

	env.MustRun("--profile", "work", "--stack", "node22")

	run := env.Docker.Runs[0]
	want := docker.Mount{Source: host, Target: "/home/developer/.config/gh"}
	if !slices.Contains(run.Mounts, want) {
		t.Fatalf("expected %+v among mounts %+v", want, run.Mounts)
	}
}

func TestStartFailsWhenAMountTargetsTheProjectPath(t *testing.T) {
	env := setup(t)
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
	env.MustRun("profile", "mounts", "add", "work", "--host", t.TempDir(), "--container", env.Cwd)

	_, err := env.Run()

	if err == nil || !strings.Contains(err.Error(), env.Cwd) || !strings.Contains(err.Error(), "project path") {
		t.Fatalf("expected a clear project-path clash error, got %v", err)
	}
	if len(env.Docker.Runs) != 0 {
		t.Fatal("expected no run")
	}
}

func TestProjectInspectShowsProfileMounts(t *testing.T) {
	env := setup(t)
	host := t.TempDir()
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
	env.MustRun("profile", "mounts", "add", "work", "--host", host, "--container", "/home/developer/.config/gh")

	out := env.MustRun("project", "inspect")

	found := false
	for _, line := range strings.Split(out, "\n") {
		if f := strings.Fields(line); len(f) == 2 && f[0] == host && f[1] == "/home/developer/.config/gh" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected the profile mount in inspect output:\n%s", out)
	}
}
