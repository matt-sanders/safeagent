package cmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeEditor sets $EDITOR to a script that appends a marker line to the file
// it is given, standing in for a developer saving an edit.
func fakeEditor(t *testing.T) {
	t.Helper()
	script := filepath.Join(t.TempDir(), "edit.sh")
	writeFile(t, script, "#!/bin/sh\necho 'RUN echo edited' >> \"$1\"\n")
	if err := os.Chmod(script, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EDITOR", script)
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestDockerfileEditsTheNamedFragmentAndPointsAtRebuild(t *testing.T) {
	env := setup(t)
	env.MustRun("profile", "create", "--name", "home")
	fakeEditor(t)
	work, home := profileID(t, env, "work"), profileID(t, env, "home")
	node22 := stackID(t, env, "node22")

	out := env.MustRun("profile", "dockerfile", "work")
	env.MustRun("stack", "dockerfile", "node22")

	if got := readFile(t, filepath.Join(env.ConfigRoot, "profiles", work, "Dockerfile")); got != "RUN echo edited\n" {
		t.Fatalf("work fragment = %q", got)
	}
	if got := readFile(t, filepath.Join(env.ConfigRoot, "profiles", home, "Dockerfile")); got != "" {
		t.Fatalf("home fragment should be untouched, got %q", got)
	}
	if got := readFile(t, filepath.Join(env.ConfigRoot, "stacks", node22, "Dockerfile")); got != "RUN echo edited\n" {
		t.Fatalf("node22 fragment = %q", got)
	}
	if !strings.Contains(out, "safeagent profile rebuild work") {
		t.Fatalf("expected a rebuild hint, got:\n%s", out)
	}
	if len(env.Docker.Builds) != 0 {
		t.Fatalf("without a terminal nothing should be rebuilt, got %v", buildTags(env.Docker.Builds))
	}
}

func TestDockerfileFallsBackToVim(t *testing.T) {
	env := setup(t)
	t.Setenv("EDITOR", "")
	t.Setenv("PATH", t.TempDir())

	_, err := env.Run("profile", "dockerfile", "work")

	if err == nil || !strings.Contains(err.Error(), "vim") {
		t.Fatalf("expected the editor to be vim, got %v", err)
	}
}

func TestEditedFragmentIsIncludedInRebuild(t *testing.T) {
	env := setup(t)
	fakeEditor(t)
	env.Docker.AddImage("safeagent/" + stackID(t, env, "node22") + ":" + profileID(t, env, "work"))

	env.MustRun("profile", "dockerfile", "work")
	env.MustRun("profile", "rebuild", "work")

	if len(env.Docker.Builds) != 1 || !strings.Contains(env.Docker.Builds[0].Dockerfile, "RUN echo edited") {
		t.Fatalf("expected one build with the edited fragment, got %+v", env.Docker.Builds)
	}
}
