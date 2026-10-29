package cmd_test

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectUseProfileSwitchesClosestAncestorProject(t *testing.T) {
	env := setup(t)
	env.MustRun("profile", "create", "--name", "home")
	outer := env.Cwd
	inner := mkdir(t, filepath.Join(outer, "packages", "api"))
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
	env.MustRun("project", "create", "-p", inner, "--profile", "work", "--stack", "node22")

	env.Cwd = mkdir(t, filepath.Join(inner, "src"))
	out := env.MustRun("project", "use-profile", "home")

	if !strings.Contains(out, "home") {
		t.Fatalf("expected confirmation naming the profile, got %q", out)
	}
	want := map[string]string{outer: profileID(t, env, "work"), inner: profileID(t, env, "home")}
	for _, p := range readProjects(t, env) {
		if p.ProfileID != want[p.Path] {
			t.Fatalf("project %s: profile ID %s, want %s", p.Path, p.ProfileID, want[p.Path])
		}
	}
	wantProfiles := "NAME  ID        MOUNTS  PROJECTS\n" +
		"work  " + profileID(t, env, "work") + "  0       1\n" +
		"home  " + profileID(t, env, "home") + "  0       1\n"
	if got := env.MustRun("profile", "ls"); got != wantProfiles {
		t.Fatalf("profile ls:\ngot:\n%s\nwant:\n%s", got, wantProfiles)
	}
}

func TestProjectUseStackWithProjectFlagSwitchesExactProject(t *testing.T) {
	env := setup(t)
	env.MustRun("stack", "create", "--name", "py", "--node", "20")
	outer := env.Cwd
	inner := mkdir(t, filepath.Join(outer, "app"))
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
	env.MustRun("project", "create", "-p", inner, "--profile", "work", "--stack", "node22")

	out := env.MustRun("project", "use-stack", "py", "-p", outer)

	if !strings.Contains(out, "py") {
		t.Fatalf("expected confirmation naming the stack, got %q", out)
	}
	want := map[string]string{outer: stackID(t, env, "py"), inner: stackID(t, env, "node22")}
	for _, p := range readProjects(t, env) {
		if p.StackID != want[p.Path] {
			t.Fatalf("project %s: stack ID %s, want %s", p.Path, p.StackID, want[p.Path])
		}
	}
	wantStacks := "NAME    ID        NODE  PROJECTS\n" +
		"node22  " + stackID(t, env, "node22") + "  22    1\n" +
		"py      " + stackID(t, env, "py") + "  20    1\n"
	if got := env.MustRun("stack", "ls"); got != wantStacks {
		t.Fatalf("stack ls:\ngot:\n%s\nwant:\n%s", got, wantStacks)
	}
}

func TestProjectUseRejectsUnknownNamesAndLeavesProjectUnchanged(t *testing.T) {
	env := setup(t)
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
	before := readProjects(t, env)

	_, err := env.Run("project", "use-profile", "nope")
	if err == nil || !strings.Contains(err.Error(), `no profile named "nope"`) {
		t.Fatalf("expected unknown profile error, got %v", err)
	}
	_, err = env.Run("project", "use-stack", "nope", "-p", env.Cwd)
	if err == nil || !strings.Contains(err.Error(), `no stack named "nope"`) {
		t.Fatalf("expected unknown stack error, got %v", err)
	}

	after := readProjects(t, env)
	if len(after) != 1 || after[0] != before[0] {
		t.Fatalf("project changed: before %+v, after %+v", before, after)
	}
}

func TestProjectUseErrorsWhenNoProjectFound(t *testing.T) {
	env := setup(t)

	for _, args := range [][]string{
		{"project", "use-profile", "work"},
		{"project", "use-stack", "node22"},
	} {
		_, err := env.Run(args...)
		if err == nil || !strings.Contains(err.Error(), "no project found") || !strings.Contains(err.Error(), "safeagent project create") {
			t.Fatalf("%v: expected a clear no-project error, got %v", args, err)
		}
	}

	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
	for _, args := range [][]string{
		{"project", "use-profile", "work", "-p", "elsewhere"},
		{"project", "use-stack", "node22", "-p", "elsewhere"},
	} {
		_, err := env.Run(args...)
		if err == nil || !strings.Contains(err.Error(), "no project found") {
			t.Fatalf("%v: expected a clear no-project error for -p, got %v", args, err)
		}
	}
}
