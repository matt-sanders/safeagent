package cmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"safeagent/clitest"
)

func TestProfileCreateIsListed(t *testing.T) {
	env := clitest.New(t)

	env.MustRun("profile", "create", "--name", "work")

	out := env.MustRun("profile", "ls")
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected header and one row, got:\n%s", out)
	}
	if got := strings.Join(strings.Fields(lines[0]), " "); got != "NAME ID MOUNTS PROJECTS" {
		t.Fatalf("unexpected header %q", lines[0])
	}
	row := strings.Fields(lines[1])
	if len(row) != 4 || row[0] != "work" || row[2] != "0" || row[3] != "0" {
		t.Fatalf("unexpected row %q", lines[1])
	}
	if len(row[1]) != 8 || strings.Trim(row[1], "0123456789abcdef") != "" {
		t.Fatalf("expected 8-char hex ID, got %q", row[1])
	}
}

func TestProfileCreateMakesFolderWithClaudeConfigAndEmptyDockerfile(t *testing.T) {
	env := clitest.New(t)

	env.MustRun("profile", "create", "--name", "work")
	id := profileID(t, env, "work")

	dir := filepath.Join(env.ConfigRoot, "profiles", id)
	if info, err := os.Stat(filepath.Join(dir, "claude")); err != nil || !info.IsDir() {
		t.Fatalf("expected Claude config folder in %s: %v", dir, err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if err != nil {
		t.Fatalf("expected Dockerfile fragment: %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("expected empty Dockerfile fragment, got %q", data)
	}
}

func TestProfileCreateRejectsInvalidNames(t *testing.T) {
	env := clitest.New(t)

	for _, name := range []string{"", "Work", "my profile", "a.b", "über", "x/y", "-x", "_x"} {
		_, err := env.Run("profile", "create", "--name", name)
		if err == nil {
			t.Errorf("expected name %q to be rejected", name)
			continue
		}
		if !strings.Contains(err.Error(), "name") {
			t.Errorf("error for %q should mention the name: %v", name, err)
		}
	}
	if out := env.MustRun("profile", "ls"); out != "No profiles.\n" {
		t.Fatalf("no profile should have been created:\n%s", out)
	}
}

func TestProfileCreateAcceptsDigitsDashesAndUnderscores(t *testing.T) {
	env := clitest.New(t)

	env.MustRun("profile", "create", "--name", "my_work-2")
	profileID(t, env, "my_work-2")
}

func TestProfileCreateRejectsDuplicateName(t *testing.T) {
	env := clitest.New(t)
	env.MustRun("profile", "create", "--name", "work")

	_, err := env.Run("profile", "create", "--name", "work")
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected duplicate-name error, got %v", err)
	}
	out := env.MustRun("profile", "ls")
	if n := strings.Count(out, "\nwork "); n != 1 {
		t.Fatalf("expected exactly one work profile, got:\n%s", out)
	}
}

func TestProfileMayShareNameWithStack(t *testing.T) {
	env := clitest.New(t)
	env.MustRun("stack", "create", "--name", "work", "--node", "22")

	env.MustRun("profile", "create", "--name", "work")
	profileID(t, env, "work")
}

func TestProfileLsEmptyState(t *testing.T) {
	env := clitest.New(t)

	if out := env.MustRun("profile", "ls"); out != "No profiles.\n" {
		t.Fatalf("unexpected empty-state output %q", out)
	}
}

func TestProfileLsListsEveryProfileInPaddedColumns(t *testing.T) {
	env := clitest.New(t)
	env.MustRun("profile", "create", "--name", "work")
	env.MustRun("profile", "create", "--name", "personal-long-name")

	out := env.MustRun("profile", "ls")
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected header and two rows, got:\n%s", out)
	}
	// Padded: the ID column starts at the same offset on every line.
	col := strings.Index(lines[0], "ID")
	for _, l := range lines[1:] {
		if l[col-2:col] != "  " || l[col] == ' ' {
			t.Fatalf("columns not aligned:\n%s", out)
		}
	}
	if profileID(t, env, "work") == profileID(t, env, "personal-long-name") {
		t.Fatalf("profiles share an ID:\n%s", out)
	}
}

// profileID returns the ID column of the `profile ls` row for name.
func profileID(t *testing.T, env *clitest.Env, name string) string {
	t.Helper()
	lines := strings.Split(strings.TrimRight(env.MustRun("profile", "ls"), "\n"), "\n")
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) == 4 && fields[0] == name {
			return fields[1]
		}
	}
	t.Fatalf("profile %q not listed", name)
	return ""
}
