package cmd_test

import (
	"testing"

	"safeagent/docker"
)

func sessionContainer(id, projectPath, profileID, stackID, workdir, runningFor string) docker.Container {
	return docker.Container{
		ID:         id,
		Image:      "safeagent/" + stackID + ":" + profileID,
		Workdir:    workdir,
		RunningFor: runningFor,
		Labels: map[string]string{
			docker.LabelProject: projectPath,
			docker.LabelProfile: profileID,
			docker.LabelStack:   stackID,
		},
	}
}

func TestSessionLsShowsLabelledContainersWithNames(t *testing.T) {
	env := setup(t)
	pid, sid := profileID(t, env, "work"), stackID(t, env, "node22")
	env.Docker.AddContainer(sessionContainer("abc123", "/code/app", pid, sid, "/code/app/web", "5 minutes"))

	out := env.MustRun("session", "ls")

	want := "CONTAINER  PROJECT    PROFILE  STACK   DIRECTORY      RUNNING FOR\n" +
		"abc123     /code/app  work     node22  /code/app/web  5 minutes\n"
	if out != want {
		t.Fatalf("got:\n%q\nwant:\n%q", out, want)
	}
}

func TestSessionLsExcludesNonSafeagentContainers(t *testing.T) {
	env := setup(t)
	pid, sid := profileID(t, env, "work"), stackID(t, env, "node22")
	env.Docker.AddContainer(docker.Container{ID: "postgres1", Image: "postgres:16", Workdir: "/", RunningFor: "2 days"})
	env.Docker.AddContainer(docker.Container{ID: "v1sess", Image: "safeagent-profile-x:latest", Labels: map[string]string{"other": "label"}})
	env.Docker.AddContainer(sessionContainer("abc123", "/code/app", pid, sid, "/code/app", "1 hour"))

	out := env.MustRun("session", "ls")

	want := "CONTAINER  PROJECT    PROFILE  STACK   DIRECTORY  RUNNING FOR\n" +
		"abc123     /code/app  work     node22  /code/app  1 hour\n"
	if out != want {
		t.Fatalf("got:\n%q\nwant:\n%q", out, want)
	}
}

func TestSessionLsShowsRawIDsWhenProfileOrStackIsGone(t *testing.T) {
	env := setup(t)
	env.Docker.AddContainer(sessionContainer("abc123", "/code/app", "deadbeef", "cafef00d", "/code/app", "3 seconds"))

	out := env.MustRun("session", "ls")

	want := "CONTAINER  PROJECT    PROFILE   STACK     DIRECTORY  RUNNING FOR\n" +
		"abc123     /code/app  deadbeef  cafef00d  /code/app  3 seconds\n"
	if out != want {
		t.Fatalf("got:\n%q\nwant:\n%q", out, want)
	}
}

func TestSessionLsWithNothingRunningPrintsEmptyState(t *testing.T) {
	env := setup(t)
	env.Docker.AddContainer(docker.Container{ID: "postgres1", Image: "postgres:16"})

	out := env.MustRun("session", "ls")

	if out != "No running sessions.\n" {
		t.Fatalf("got %q", out)
	}
}
