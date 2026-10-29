package cmd_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"safeagent/clitest"
)

// stackID pulls the ID for the named Stack out of `stack ls` output.
func stackID(t *testing.T, env *clitest.Env, name string) string {
	t.Helper()
	for _, line := range strings.Split(env.MustRun("stack", "ls"), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == name {
			return fields[1]
		}
	}
	t.Fatalf("stack %q not listed", name)
	return ""
}

func TestStackCreateWithFlagsCreatesStackAndEmptyFragment(t *testing.T) {
	env := clitest.New(t)

	env.MustRun("stack", "create", "--name", "node22", "--node", "22")

	id := stackID(t, env, "node22")
	if !regexp.MustCompile(`^[0-9a-f]{8}$`).MatchString(id) {
		t.Fatalf("expected 8-char hex ID, got %q", id)
	}
	data, err := os.ReadFile(filepath.Join(env.ConfigRoot, "stacks", id, "Dockerfile"))
	if err != nil {
		t.Fatalf("expected Dockerfile fragment: %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("expected empty fragment, got %q", data)
	}
}

func TestStackCreateRejectsInvalidNames(t *testing.T) {
	env := clitest.New(t)

	for _, name := range []string{"Node22", "my stack", "node.22", "node/22"} {
		_, err := env.Run("stack", "create", "--name", name, "--node", "22")
		if err == nil {
			t.Errorf("expected %q to be rejected", name)
			continue
		}
		if !strings.Contains(err.Error(), "invalid name") || !strings.Contains(err.Error(), "lowercase") {
			t.Errorf("unclear error for %q: %v", name, err)
		}
	}
	if out := env.MustRun("stack", "ls"); !strings.Contains(out, "No stacks.") {
		t.Fatalf("expected no stacks to be created, got:\n%s", out)
	}
}

func TestStackCreateRejectsEmptyName(t *testing.T) {
	env := clitest.New(t)

	if _, err := env.Run("stack", "create", "--name", "", "--node", "22"); err == nil {
		t.Fatal("expected empty name to be rejected")
	}
}

func TestStackCreateRejectsDuplicateName(t *testing.T) {
	env := clitest.New(t)
	env.MustRun("stack", "create", "--name", "node22", "--node", "22")

	_, err := env.Run("stack", "create", "--name", "node22", "--node", "20")
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected duplicate name error, got %v", err)
	}
	ls := env.MustRun("stack", "ls")
	if strings.Count(ls, "node22") != 1 {
		t.Fatalf("expected exactly one node22 stack, got:\n%s", ls)
	}
}

func TestStackLsPrintsPaddedTable(t *testing.T) {
	env := clitest.New(t)
	env.MustRun("stack", "create", "--name", "a", "--node", "22")
	env.MustRun("stack", "create", "--name", "longer-name", "--node", "lts")
	idA := stackID(t, env, "a")
	idB := stackID(t, env, "longer-name")

	want := "NAME         ID        NODE  PROJECTS\n" +
		"a            " + idA + "  22    0\n" +
		"longer-name  " + idB + "  lts   0\n"
	if got := env.MustRun("stack", "ls"); got != want {
		t.Fatalf("unexpected table:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestStackLsEmptyState(t *testing.T) {
	env := clitest.New(t)

	if got := env.MustRun("stack", "ls"); got != "No stacks.\n" {
		t.Fatalf("unexpected empty output: %q", got)
	}
}
