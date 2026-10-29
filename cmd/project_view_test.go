package cmd_test

import (
	"path/filepath"
	"strings"
	"testing"

	"safeagent/clitest"
)

func TestProjectLsWithNoProjectsPrintsEmptyState(t *testing.T) {
	env := clitest.New(t)

	out := env.MustRun("project", "ls")

	if out != "No projects.\n" {
		t.Fatalf("got %q", out)
	}
}

func TestProjectLsShowsNamesAndImageState(t *testing.T) {
	env := setup(t)
	env.MustRun("profile", "create", "--name", "home")
	root := t.TempDir()
	a := mkdir(t, filepath.Join(root, "a"))
	bb := mkdir(t, filepath.Join(root, "bb"))
	env.MustRun("project", "create", "-p", a, "--profile", "work", "--stack", "node22")
	env.MustRun("project", "create", "-p", bb, "--profile", "home", "--stack", "node22")
	env.Docker.AddImage("safeagent/" + stackID(t, env, "node22") + ":" + profileID(t, env, "work"))

	out := env.MustRun("project", "ls")

	want := "PATH" + strings.Repeat(" ", len(bb)-4) + "  PROFILE  STACK   IMAGE\n" +
		a + " " + "  work     node22  built\n" +
		bb + "  home     node22  not built\n"
	if out != want {
		t.Fatalf("got:\n%s\nwant:\n%s", out, want)
	}
}

func TestProjectInspectWithFlagShowsDetailsAndBaseMounts(t *testing.T) {
	env := setup(t)
	dir := mkdir(t, filepath.Join(t.TempDir(), "app"))
	env.MustRun("project", "create", "-p", dir, "--profile", "work", "--stack", "node22")
	claude := filepath.Join(env.ConfigRoot, "profiles", profileID(t, env, "work"), "claude")

	out := env.MustRun("project", "inspect", "-p", dir)

	want := "Path:     " + dir + "\n" +
		"Profile:  work\n" +
		"Stack:    node22\n" +
		"Image:    node22:work (not built)\n" +
		"\nMounts:\n" +
		"HOST" + strings.Repeat(" ", len(claude)-4) + "  CONTAINER\n" +
		dir + strings.Repeat(" ", len(claude)-len(dir)) + "  " + dir + "\n" +
		claude + "  /claude\n" +
		"\nExclusions:\n" +
		"No exclusions.\n"
	if out != want {
		t.Fatalf("got:\n%s\nwant:\n%s", out, want)
	}
}

func TestProjectInspectUsesInnermostAncestorAndShowsBuiltImage(t *testing.T) {
	env := setup(t)
	env.MustRun("profile", "create", "--name", "home")
	outer := mkdir(t, filepath.Join(t.TempDir(), "outer"))
	inner := mkdir(t, filepath.Join(outer, "inner"))
	env.MustRun("project", "create", "-p", outer, "--profile", "work", "--stack", "node22")
	env.MustRun("project", "create", "-p", inner, "--profile", "home", "--stack", "node22")
	env.Docker.AddImage("safeagent/" + stackID(t, env, "node22") + ":" + profileID(t, env, "home"))
	env.Cwd = mkdir(t, filepath.Join(inner, "src"))

	out := env.MustRun("project", "inspect")

	if !strings.HasPrefix(out, "Path:     "+inner+"\nProfile:  home\n") {
		t.Fatalf("expected inner project, got:\n%s", out)
	}
	if !strings.Contains(out, "Image:    node22:home (built)\n") {
		t.Fatalf("expected built image, got:\n%s", out)
	}
}

func TestProjectInspectErrorsWhenNoProjectFound(t *testing.T) {
	env := setup(t)

	_, err := env.Run("project", "inspect")
	if err == nil || !strings.Contains(err.Error(), "no project found") {
		t.Fatalf("expected no-project error, got %v", err)
	}

	_, err = env.Run("project", "inspect", "-p", t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "no project found") {
		t.Fatalf("expected no-project error for -p, got %v", err)
	}
}
