package cmd_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"safeagent/clitest"
)

// storedProject is the on-disk shape of a Project record, read to check that
// Projects reference Profiles and Stacks by ID rather than by Name.
type storedProject struct {
	Path      string `json:"path"`
	ProfileID string `json:"profile_id"`
	StackID   string `json:"stack_id"`
}

func readProjects(t *testing.T, env *clitest.Env) []storedProject {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(env.ConfigRoot, "projects.json"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var ps []storedProject
	if err := json.Unmarshal(data, &ps); err != nil {
		t.Fatal(err)
	}
	return ps
}

// setup creates the "work" Profile and "node22" Stack most tests use.
func setup(t *testing.T) *clitest.Env {
	t.Helper()
	env := clitest.New(t)
	env.MustRun("profile", "create", "--name", "work")
	env.MustRun("stack", "create", "--name", "node22", "--node", "22")
	return env
}

func mkdir(t *testing.T, path string) string {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestProjectCreateReferencesProfileAndStackByID(t *testing.T) {
	env := setup(t)
	dir := mkdir(t, filepath.Join(t.TempDir(), "app"))

	env.MustRun("project", "create", "-p", dir, "--profile", "work", "--stack", "node22")

	ps := readProjects(t, env)
	if len(ps) != 1 {
		t.Fatalf("expected one project, got %+v", ps)
	}
	want := storedProject{Path: dir, ProfileID: profileID(t, env, "work"), StackID: stackID(t, env, "node22")}
	if ps[0] != want {
		t.Fatalf("got %+v, want %+v", ps[0], want)
	}
}

func TestStackAndProfileLsCountProjects(t *testing.T) {
	env := setup(t)
	env.MustRun("profile", "create", "--name", "home")
	env.MustRun("stack", "create", "--name", "py", "--node", "20")
	root := t.TempDir()
	env.MustRun("project", "create", "-p", mkdir(t, filepath.Join(root, "a")), "--profile", "work", "--stack", "node22")
	env.MustRun("project", "create", "-p", mkdir(t, filepath.Join(root, "b")), "--profile", "work", "--stack", "py")

	wantStacks := "NAME    ID        NODE  PROJECTS\n" +
		"node22  " + stackID(t, env, "node22") + "  22    1\n" +
		"py      " + stackID(t, env, "py") + "  20    1\n"
	if got := env.MustRun("stack", "ls"); got != wantStacks {
		t.Fatalf("stack ls:\ngot:\n%s\nwant:\n%s", got, wantStacks)
	}
	wantProfiles := "NAME  ID        MOUNTS  PROJECTS\n" +
		"work  " + profileID(t, env, "work") + "  0       2\n" +
		"home  " + profileID(t, env, "home") + "  0       0\n"
	if got := env.MustRun("profile", "ls"); got != wantProfiles {
		t.Fatalf("profile ls:\ngot:\n%s\nwant:\n%s", got, wantProfiles)
	}
}

func TestProjectCreateDefaultsToCurrentDirectory(t *testing.T) {
	env := setup(t)

	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")

	if ps := readProjects(t, env); len(ps) != 1 || ps[0].Path != env.Cwd {
		t.Fatalf("expected a project at %s, got %+v", env.Cwd, ps)
	}
}

func TestProjectCreateResolvesRelativePathAgainstCurrentDirectory(t *testing.T) {
	env := setup(t)
	mkdir(t, filepath.Join(env.Cwd, "sub", "app"))

	env.MustRun("project", "create", "-p", "sub/../sub/app/", "--profile", "work", "--stack", "node22")

	want := filepath.Join(env.Cwd, "sub", "app")
	if ps := readProjects(t, env); len(ps) != 1 || ps[0].Path != want {
		t.Fatalf("expected a project at %s, got %+v", want, ps)
	}
}

func TestProjectCreateRefusesExistingPathButAllowsNested(t *testing.T) {
	env := setup(t)
	outer := env.Cwd
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")

	_, err := env.Run("project", "create", "-p", outer, "--profile", "work", "--stack", "node22")
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected duplicate-path error, got %v", err)
	}

	inner := mkdir(t, filepath.Join(outer, "packages", "api"))
	env.MustRun("project", "create", "-p", inner, "--profile", "work", "--stack", "node22")
	if ps := readProjects(t, env); len(ps) != 2 {
		t.Fatalf("expected outer and nested project, got %+v", ps)
	}
}

func TestProjectCreateRejectsUnknownNames(t *testing.T) {
	env := setup(t)

	_, err := env.Run("project", "create", "--profile", "nope", "--stack", "node22")
	if err == nil || !strings.Contains(err.Error(), `no profile named "nope"`) {
		t.Fatalf("expected unknown profile error, got %v", err)
	}
	_, err = env.Run("project", "create", "--profile", "work", "--stack", "nope")
	if err == nil || !strings.Contains(err.Error(), `no stack named "nope"`) {
		t.Fatalf("expected unknown stack error, got %v", err)
	}
	if ps := readProjects(t, env); len(ps) != 0 {
		t.Fatalf("expected no projects, got %+v", ps)
	}
}

func TestProjectCreateWithoutAnyProfileOrStackErrorsInsteadOfPrompting(t *testing.T) {
	env := clitest.New(t)

	_, err := env.Run("project", "create", "--stack", "node22")
	if err == nil || !strings.Contains(err.Error(), "safeagent profile create") {
		t.Fatalf("expected a hint to create a profile, got %v", err)
	}

	env.MustRun("profile", "create", "--name", "work")
	_, err = env.Run("project", "create", "--profile", "work")
	if err == nil || !strings.Contains(err.Error(), "safeagent stack create") {
		t.Fatalf("expected a hint to create a stack, got %v", err)
	}
}

func TestProjectRmRemovesInnermostAncestorAndLeavesFiles(t *testing.T) {
	env := setup(t)
	outer := env.Cwd
	inner := mkdir(t, filepath.Join(outer, "packages", "api"))
	deep := mkdir(t, filepath.Join(inner, "src", "lib"))
	file := filepath.Join(inner, "keep.txt")
	if err := os.WriteFile(file, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
	env.MustRun("project", "create", "-p", inner, "--profile", "work", "--stack", "node22")

	env.Cwd = deep
	env.MustRun("project", "rm")

	ps := readProjects(t, env)
	if len(ps) != 1 || ps[0].Path != outer {
		t.Fatalf("expected only the outer project to remain, got %+v", ps)
	}
	if data, err := os.ReadFile(file); err != nil || string(data) != "keep" {
		t.Fatalf("files on disk must be untouched: %v", err)
	}
	if _, err := os.Stat(deep); err != nil {
		t.Fatalf("project directories must be untouched: %v", err)
	}
}

func TestProjectRmWithRelativePathRemovesExactProject(t *testing.T) {
	env := setup(t)
	inner := mkdir(t, filepath.Join(env.Cwd, "app"))
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
	env.MustRun("project", "create", "-p", "app", "--profile", "work", "--stack", "node22")

	env.MustRun("project", "rm", "-p", "app")

	ps := readProjects(t, env)
	if len(ps) != 1 || ps[0].Path == inner {
		t.Fatalf("expected the app project to be removed, got %+v", ps)
	}
}

func TestProjectRmErrorsWhenNoProjectFound(t *testing.T) {
	env := setup(t)

	_, err := env.Run("project", "rm")
	if err == nil || !strings.Contains(err.Error(), "no project found") || !strings.Contains(err.Error(), "safeagent project create") {
		t.Fatalf("expected a clear no-project error, got %v", err)
	}

	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
	_, err = env.Run("project", "rm", "-p", "elsewhere")
	if err == nil || !strings.Contains(err.Error(), "no project found") {
		t.Fatalf("expected a clear no-project error for -p, got %v", err)
	}
}
