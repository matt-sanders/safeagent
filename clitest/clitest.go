// Package clitest runs the safeagent command tree in-process for tests, with a
// temp-dir config root and a recording fake Docker.
package clitest

import (
	"bytes"
	"testing"

	"safeagent/cmd"
	"safeagent/docker/dockertest"
)

// Env is one test's safeagent installation. The same Env can run many
// commands; state persists across runs via ConfigRoot and Docker.
type Env struct {
	t          *testing.T
	ConfigRoot string
	Docker     *dockertest.Fake
}

// New returns an Env with an empty temp config root and an empty fake Docker.
func New(t *testing.T) *Env {
	t.Helper()
	return &Env{t: t, ConfigRoot: t.TempDir(), Docker: dockertest.New()}
}

// Run executes `safeagent <args...>` against a freshly built command tree and
// returns everything written to stdout and the command's error. Pass every
// value as a flag or argument so no interactive prompt runs.
func (e *Env) Run(args ...string) (string, error) {
	e.t.Helper()
	root := cmd.NewRootCommand(cmd.Deps{ConfigRoot: e.ConfigRoot, Docker: e.Docker})
	var out, errOut bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetIn(&bytes.Buffer{})
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

// MustRun is Run that fails the test on error.
func (e *Env) MustRun(args ...string) string {
	e.t.Helper()
	out, err := e.Run(args...)
	if err != nil {
		e.t.Fatalf("safeagent %v: %v\noutput:\n%s", args, err, out)
	}
	return out
}
