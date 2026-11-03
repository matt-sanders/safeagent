package cmd_test

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"safeagent/docker"
)

const exclusionsNote = "Paths are relative to the directory the session starts in."

func TestExclusionsAddThenLsShowsTableWithNote(t *testing.T) {
	env := setup(t)
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")

	if got := env.MustRun("project", "exclusions", "ls"); got != "No exclusions.\n" {
		t.Fatalf("empty ls: got %q", got)
	}

	out := env.MustRun("project", "exclusions", "add", "node_modules")
	if !strings.Contains(out, "node_modules") {
		t.Fatalf("expected confirmation naming the exclusion, got %q", out)
	}
	env.MustRun("project", "exclusions", "add", "build/")

	want := "PATH\n" +
		"node_modules\n" +
		"build\n" +
		"\n" + exclusionsNote + "\n"
	if got := env.MustRun("project", "exclusions", "ls"); got != want {
		t.Fatalf("ls:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestExclusionsRejectDuplicatesAndMissingRemovals(t *testing.T) {
	env := setup(t)
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
	env.MustRun("project", "exclusions", "add", "node_modules")

	_, err := env.Run("project", "exclusions", "add", "./node_modules")
	if err == nil || !strings.Contains(err.Error(), "already excluded") {
		t.Fatalf("expected duplicate error, got %v", err)
	}
	_, err = env.Run("project", "exclusions", "rm", "build")
	if err == nil || !strings.Contains(err.Error(), "not excluded") {
		t.Fatalf("expected missing-exclusion error, got %v", err)
	}
	for _, bad := range []string{"/abs/path", "../outside", "."} {
		if _, err := env.Run("project", "exclusions", "add", bad); err == nil {
			t.Fatalf("expected %q to be rejected", bad)
		}
	}

	out := env.MustRun("project", "exclusions", "rm", "node_modules")
	if !strings.Contains(out, "node_modules") {
		t.Fatalf("expected confirmation naming the exclusion, got %q", out)
	}
	if got := env.MustRun("project", "exclusions", "ls"); got != "No exclusions.\n" {
		t.Fatalf("after rm: got %q", got)
	}
}

func TestExclusionsTargetProjectFlagOrClosestAncestor(t *testing.T) {
	env := setup(t)
	outer := env.Cwd
	inner := mkdir(t, filepath.Join(outer, "packages", "api"))
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
	env.MustRun("project", "create", "-p", inner, "--profile", "work", "--stack", "node22")

	env.Cwd = mkdir(t, filepath.Join(inner, "src"))
	env.MustRun("project", "exclusions", "add", "dist")
	env.MustRun("project", "exclusions", "add", "node_modules", "-p", outer)

	if got := env.MustRun("project", "exclusions", "ls"); !strings.Contains(got, "dist") || strings.Contains(got, "node_modules") {
		t.Fatalf("closest-ancestor ls: got %q", got)
	}
	if got := env.MustRun("project", "exclusions", "ls", "-p", outer); !strings.Contains(got, "node_modules") || strings.Contains(got, "dist") {
		t.Fatalf("-p ls: got %q", got)
	}
	env.MustRun("project", "exclusions", "rm", "node_modules", "-p", outer)
	if got := env.MustRun("project", "exclusions", "ls", "-p", outer); got != "No exclusions.\n" {
		t.Fatalf("-p rm: got %q", got)
	}

	env.Cwd = t.TempDir()
	for _, args := range [][]string{
		{"project", "exclusions", "ls"},
		{"project", "exclusions", "add", "x"},
		{"project", "exclusions", "rm", "x"},
	} {
		if _, err := env.Run(args...); err == nil || !strings.Contains(err.Error(), "no project found") {
			t.Fatalf("%v: expected no-project error, got %v", args, err)
		}
	}
}

func TestSessionFromSubdirectoryGetsAnonymousVolumesForExclusions(t *testing.T) {
	env := setup(t)
	root := env.Cwd
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
	env.MustRun("project", "exclusions", "add", "node_modules")
	env.MustRun("project", "exclusions", "add", "build/out")

	sub := mkdir(t, filepath.Join(root, "packages", "web"))
	env.Cwd = sub
	env.MustRun()

	run := env.Docker.Runs[0]
	for _, want := range []docker.Mount{
		{Source: "", Target: filepath.Join(sub, "node_modules")},
		{Source: "", Target: filepath.Join(sub, "build", "out")},
	} {
		if !slices.Contains(run.Mounts, want) {
			t.Fatalf("expected anonymous volume %+v in %+v", want, run.Mounts)
		}
	}
	for _, m := range run.Mounts {
		if m.Target == filepath.Join(root, "node_modules") {
			t.Fatalf("exclusion applied relative to the project root, not cwd: %+v", run.Mounts)
		}
	}
}

func TestProjectInspectShowsExclusions(t *testing.T) {
	env := setup(t)
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")

	if got := env.MustRun("project", "inspect"); !strings.HasSuffix(got, "\nExclusions:\nNo exclusions.\n") {
		t.Fatalf("expected an empty Exclusions section, got:\n%s", got)
	}

	env.MustRun("project", "exclusions", "add", "node_modules")
	env.MustRun("project", "exclusions", "add", "dist")
	want := "\nExclusions:\nPATH\nnode_modules\ndist\n\n" + exclusionsNote + "\n"
	if got := env.MustRun("project", "inspect"); !strings.HasSuffix(got, want) {
		t.Fatalf("inspect:\ngot:\n%s\nwant suffix:\n%s", got, want)
	}
}
