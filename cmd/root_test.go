package cmd_test

import (
	"strings"
	"testing"

	"safeagent/clitest"
)

func TestHelpDescribesSafeagent(t *testing.T) {
	env := clitest.New(t)

	out, err := env.Run("--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "safeagent") || !strings.Contains(out, "Run agents safely inside Docker") {
		t.Fatalf("help output missing description:\n%s", out)
	}
}

func TestV1CommandsAreGone(t *testing.T) {
	env := clitest.New(t)

	for _, name := range []string{"auth", "profiles", "sessions", "exclude", "use-auth", "use-profile"} {
		if _, err := env.Run(name); err == nil {
			t.Errorf("expected v1 command %q to be unknown", name)
		}
	}
}
