package cmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func TestProfileRmRefusesWhileProjectsUseIt(t *testing.T) {
	env := setup(t)
	a := mkdir(t, filepath.Join(t.TempDir(), "a"))
	b := mkdir(t, filepath.Join(t.TempDir(), "b"))
	env.MustRun("project", "create", "-p", a, "--profile", "work", "--stack", "node22")
	env.MustRun("project", "create", "-p", b, "--profile", "work", "--stack", "node22")
	id := profileID(t, env, "work")
	ref := "safeagent/" + stackID(t, env, "node22") + ":" + id
	env.Docker.AddImage(ref)

	_, err := env.Run("profile", "rm", "work")
	if err == nil {
		t.Fatal("expected profile rm to refuse while Projects use it")
	}
	for _, p := range []string{a, b} {
		if !strings.Contains(err.Error(), p) {
			t.Errorf("expected error to list %s, got: %v", p, err)
		}
	}
	if !strings.Contains(env.MustRun("profile", "ls"), "work") {
		t.Error("expected Profile to remain")
	}
	if !dirExists(filepath.Join(env.ConfigRoot, "profiles", id)) {
		t.Error("expected Profile folder to remain")
	}
	if !env.Docker.HasImage(ref) || len(env.Docker.Removed) != 0 {
		t.Errorf("expected no images removed, removed: %v", env.Docker.Removed)
	}
}

func TestProfileRmRemovesUnusedProfileFolderAndItsImages(t *testing.T) {
	env := setup(t)
	env.MustRun("profile", "create", "--name", "home")
	env.MustRun("stack", "create", "--name", "py", "--node", "20")
	work := profileID(t, env, "work")
	home := profileID(t, env, "home")
	s1, s2 := stackID(t, env, "node22"), stackID(t, env, "py")
	env.Docker.AddImage("safeagent/" + s1 + ":" + work)
	env.Docker.AddImage("safeagent/" + s2 + ":" + work)
	env.Docker.AddImage("safeagent/" + s1 + ":" + home)

	out := env.MustRun("profile", "rm", "work")

	for _, name := range []string{"node22:work", "py:work"} {
		if !strings.Contains(out, "Removed image "+name+".\n") {
			t.Errorf("expected removal of %s reported by name, got:\n%s", name, out)
		}
	}

	if strings.Contains(env.MustRun("profile", "ls"), "work") {
		t.Error("expected Profile record removed")
	}
	if dirExists(filepath.Join(env.ConfigRoot, "profiles", work)) {
		t.Error("expected Profile folder removed")
	}
	want := "safeagent/" + s1 + ":" + home
	if got := env.Docker.Images(); len(got) != 1 || got[0] != want {
		t.Errorf("images left: %v, want [%s]", got, want)
	}
	if !dirExists(filepath.Join(env.ConfigRoot, "profiles", home)) {
		t.Error("expected other Profile folder kept")
	}
}

func TestProfileRmUnknownNameErrors(t *testing.T) {
	env := setup(t)

	_, err := env.Run("profile", "rm", "nope")
	if err == nil || !strings.Contains(err.Error(), `no profile named "nope"`) {
		t.Fatalf("expected unknown-name error, got %v", err)
	}
}

func TestStackRmRefusesWhileProjectsUseIt(t *testing.T) {
	env := setup(t)
	a := mkdir(t, filepath.Join(t.TempDir(), "a"))
	env.MustRun("project", "create", "-p", a, "--profile", "work", "--stack", "node22")
	id := stackID(t, env, "node22")
	ref := "safeagent/" + id + ":" + profileID(t, env, "work")
	env.Docker.AddImage(ref)

	_, err := env.Run("stack", "rm", "node22")
	if err == nil || !strings.Contains(err.Error(), a) {
		t.Fatalf("expected refusal listing %s, got %v", a, err)
	}
	if !strings.Contains(env.MustRun("stack", "ls"), "node22") {
		t.Error("expected Stack to remain")
	}
	if !dirExists(filepath.Join(env.ConfigRoot, "stacks", id)) {
		t.Error("expected Stack folder to remain")
	}
	if !env.Docker.HasImage(ref) || len(env.Docker.Removed) != 0 {
		t.Errorf("expected no images removed, removed: %v", env.Docker.Removed)
	}
}

func TestStackRmRemovesUnusedStackFolderAndItsImages(t *testing.T) {
	env := setup(t)
	env.MustRun("profile", "create", "--name", "home")
	env.MustRun("stack", "create", "--name", "py", "--node", "20")
	work, home := profileID(t, env, "work"), profileID(t, env, "home")
	node22, py := stackID(t, env, "node22"), stackID(t, env, "py")
	env.Docker.AddImage("safeagent/" + node22 + ":" + work)
	env.Docker.AddImage("safeagent/" + node22 + ":" + home)
	env.Docker.AddImage("safeagent/" + py + ":" + work)

	env.MustRun("stack", "rm", "node22")

	if strings.Contains(env.MustRun("stack", "ls"), "node22") {
		t.Error("expected Stack record removed")
	}
	if dirExists(filepath.Join(env.ConfigRoot, "stacks", node22)) {
		t.Error("expected Stack folder removed")
	}
	want := "safeagent/" + py + ":" + work
	if got := env.Docker.Images(); len(got) != 1 || got[0] != want {
		t.Errorf("images left: %v, want [%s]", got, want)
	}
}

func TestStackRmUnknownNameErrors(t *testing.T) {
	env := setup(t)

	_, err := env.Run("stack", "rm", "nope")
	if err == nil || !strings.Contains(err.Error(), `no stack named "nope"`) {
		t.Fatalf("expected unknown-name error, got %v", err)
	}
}
