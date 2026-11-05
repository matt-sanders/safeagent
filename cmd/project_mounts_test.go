package cmd_test

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"safeagent/docker"
)

func TestProjectMountsAddThenLsShowsTheMount(t *testing.T) {
	env := setup(t)
	host := t.TempDir()
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")

	if got := env.MustRun("project", "mounts", "ls"); got != "No mounts.\n" {
		t.Fatalf("empty ls: got %q", got)
	}

	out := env.MustRun("project", "mounts", "add", "--host", host, "--container", "/data")

	if !strings.Contains(out, "/data") {
		t.Fatalf("expected add to confirm the mount, got:\n%s", out)
	}
	ls := env.MustRun("project", "mounts", "ls")
	lines := strings.Split(strings.TrimSpace(ls), "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[0], "HOST PATH") {
		t.Fatalf("expected a header and one row, got:\n%s", ls)
	}
	if f := strings.Fields(lines[1]); len(f) != 2 || f[0] != host || f[1] != "/data" {
		t.Fatalf("unexpected row %q", lines[1])
	}
}

func TestProjectMountsAddRejectsInvalidMounts(t *testing.T) {
	env := setup(t)
	host := t.TempDir()
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
	env.MustRun("project", "mounts", "add", "--host", host, "--container", "/data")

	cases := []struct {
		name, host, container, want string
	}{
		{"missing host path", filepath.Join(host, "nope"), "/other", "does not exist"},
		{"relative container path", host, "other", "must be absolute"},
		{"claude config path", host, "/claude", "reserved"},
		{"project path", host, env.Cwd, "project path"},
		{"duplicate container path", host, "/data/", "already has a mount"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := env.Run("project", "mounts", "add", "--host", tc.host, "--container", tc.container)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected an error containing %q, got %v", tc.want, err)
			}
		})
	}
	if got := strings.Count(env.MustRun("project", "mounts", "ls"), "\n"); got != 2 {
		t.Fatalf("expected only the first mount to be stored, got %d lines", got)
	}
}

func TestProjectMountsAddResolvesRelativeHostPathAgainstCurrentDirectory(t *testing.T) {
	env := setup(t)
	host := mkdir(t, filepath.Join(env.Cwd, "data"))
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")

	env.MustRun("project", "mounts", "add", "--host", "data", "--container", "/data")

	if ls := env.MustRun("project", "mounts", "ls"); !strings.Contains(ls, host+" ") {
		t.Fatalf("expected host path %s, got:\n%s", host, ls)
	}
}

func TestProjectMountsRmRemovesByContainerPath(t *testing.T) {
	env := setup(t)
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
	env.MustRun("project", "mounts", "add", "--host", t.TempDir(), "--container", "/data")

	env.MustRun("project", "mounts", "rm", "/data/")

	if got := env.MustRun("project", "mounts", "ls"); got != "No mounts.\n" {
		t.Fatalf("expected no mounts, got:\n%s", got)
	}
	_, err := env.Run("project", "mounts", "rm", "/data")
	if err == nil || !strings.Contains(err.Error(), "no mount at") {
		t.Fatalf("expected an unknown-mount error, got %v", err)
	}
}

func TestProjectMountsTargetProjectFlagOrClosestAncestor(t *testing.T) {
	env := setup(t)
	outer := env.Cwd
	inner := mkdir(t, filepath.Join(outer, "packages", "api"))
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
	env.MustRun("project", "create", "-p", inner, "--profile", "work", "--stack", "node22")

	env.Cwd = mkdir(t, filepath.Join(inner, "src"))
	env.MustRun("project", "mounts", "add", "--host", t.TempDir(), "--container", "/inner")
	env.MustRun("project", "mounts", "add", "-p", outer, "--host", t.TempDir(), "--container", "/outer")

	if got := env.MustRun("project", "mounts", "ls"); !strings.Contains(got, "/inner") || strings.Contains(got, "/outer") {
		t.Fatalf("closest-ancestor ls: got %q", got)
	}
	if got := env.MustRun("project", "mounts", "ls", "-p", outer); !strings.Contains(got, "/outer") || strings.Contains(got, "/inner") {
		t.Fatalf("-p ls: got %q", got)
	}
	env.MustRun("project", "mounts", "rm", "/outer", "-p", outer)
	if got := env.MustRun("project", "mounts", "ls", "-p", outer); got != "No mounts.\n" {
		t.Fatalf("-p rm: got %q", got)
	}

	env.Cwd = t.TempDir()
	for _, args := range [][]string{
		{"project", "mounts", "ls"},
		{"project", "mounts", "add", "--host", outer, "--container", "/x"},
		{"project", "mounts", "rm", "/x"},
	} {
		if _, err := env.Run(args...); err == nil || !strings.Contains(err.Error(), "no project found") {
			t.Fatalf("%v: expected no-project error, got %v", args, err)
		}
	}
}

func TestStartIncludesOnlyThatProjectsMounts(t *testing.T) {
	env := setup(t)
	root := env.Cwd
	host := t.TempDir()
	other := mkdir(t, filepath.Join(root, "other"))
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
	env.MustRun("project", "create", "-p", other, "--profile", "work", "--stack", "node22")
	env.MustRun("project", "mounts", "add", "--host", host, "--container", "/data")

	env.MustRun()
	env.Cwd = other
	env.MustRun()

	want := docker.Mount{Source: host, Target: "/data"}
	if !slices.Contains(env.Docker.Runs[0].Mounts, want) {
		t.Fatalf("expected %+v among mounts %+v", want, env.Docker.Runs[0].Mounts)
	}
	if slices.Contains(env.Docker.Runs[1].Mounts, want) {
		t.Fatalf("expected no %+v in another project's mounts %+v", want, env.Docker.Runs[1].Mounts)
	}
}

func TestProjectMountReplacesProfileMountAtTheSamePath(t *testing.T) {
	env := setup(t)
	profileHost, projectHost := t.TempDir(), t.TempDir()
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
	env.MustRun("profile", "mounts", "add", "work", "--host", profileHost, "--container", "/data")
	env.MustRun("project", "mounts", "add", "--host", projectHost, "--container", "/data")

	env.MustRun()

	var sources []string
	for _, m := range env.Docker.Runs[0].Mounts {
		if m.Target == "/data" {
			sources = append(sources, m.Source)
		}
	}
	if !slices.Equal(sources, []string{projectHost}) {
		t.Fatalf("expected only the project's mount at /data, got %v", sources)
	}
	if inspect := env.MustRun("project", "inspect"); strings.Contains(inspect, profileHost) || !strings.Contains(inspect, projectHost) {
		t.Fatalf("expected inspect to show only the project's mount:\n%s", inspect)
	}
}

func TestProjectMountsSurviveSwitchingProfile(t *testing.T) {
	env := setup(t)
	host := t.TempDir()
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
	env.MustRun("project", "mounts", "add", "--host", host, "--container", "/data")
	env.MustRun("profile", "create", "--name", "personal")

	env.MustRun("project", "use-profile", "personal")

	if ls := env.MustRun("project", "mounts", "ls"); !strings.Contains(ls, host) {
		t.Fatalf("expected the mount to survive the switch, got:\n%s", ls)
	}
}

func TestProjectInspectShowsProjectMounts(t *testing.T) {
	env := setup(t)
	host := t.TempDir()
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
	env.MustRun("project", "mounts", "add", "--host", host, "--container", "/data")

	out := env.MustRun("project", "inspect")

	found := false
	for _, line := range strings.Split(out, "\n") {
		if f := strings.Fields(line); len(f) == 2 && f[0] == host && f[1] == "/data" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected the project mount in inspect output:\n%s", out)
	}
}
