package cmd_test

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"safeagent/docker"
)

func TestStartWithoutProjectCreatesOneAndStartsClaude(t *testing.T) {
	env := setup(t)

	env.MustRun("--profile", "work", "--stack", "node22")

	ps := readProjects(t, env)
	if len(ps) != 1 || ps[0].Path != env.Cwd {
		t.Fatalf("expected a project at %s, got %+v", env.Cwd, ps)
	}
	if len(env.Docker.Runs) != 1 {
		t.Fatalf("expected one run, got %+v", env.Docker.Runs)
	}
	run := env.Docker.Runs[0]
	wantImage := "safeagent/" + stackID(t, env, "node22") + ":" + profileID(t, env, "work")
	if run.Image != wantImage {
		t.Fatalf("ran image %q, want %q", run.Image, wantImage)
	}
	if !slices.Equal(run.Command, []string{"claude", "--dangerously-skip-permissions"}) {
		t.Fatalf("unexpected command %v", run.Command)
	}
}

func TestStartFromSubdirectoryUsesInnermostProject(t *testing.T) {
	env := setup(t)
	env.MustRun("profile", "create", "--name", "home")
	outer := env.Cwd
	inner := mkdir(t, filepath.Join(outer, "packages", "api"))
	deep := mkdir(t, filepath.Join(inner, "src"))
	env.MustRun("project", "create", "-p", outer, "--profile", "work", "--stack", "node22")
	env.MustRun("project", "create", "-p", inner, "--profile", "home", "--stack", "node22")

	env.Cwd = deep
	env.MustRun()

	if got := len(readProjects(t, env)); got != 2 {
		t.Fatalf("starting a session should not create a project, got %d", got)
	}
	run := env.Docker.Runs[0]
	if run.Labels[docker.LabelProject] != inner || run.Labels[docker.LabelProfile] != profileID(t, env, "home") {
		t.Fatalf("expected the inner project with profile home, got labels %v", run.Labels)
	}
}

func TestStartBuildsMissingImageWithStackThenProfileFragmentOnce(t *testing.T) {
	env := setup(t)
	sid, pid := stackID(t, env, "node22"), profileID(t, env, "work")
	writeFile(t, filepath.Join(env.ConfigRoot, "stacks", sid, "Dockerfile"), "RUN echo stack-fragment\n")
	writeFile(t, filepath.Join(env.ConfigRoot, "profiles", pid, "Dockerfile"), "RUN echo profile-fragment\n")

	env.MustRun("--profile", "work", "--stack", "node22")
	env.MustRun()

	if len(env.Docker.Builds) != 1 {
		t.Fatalf("expected exactly one build, got %d", len(env.Docker.Builds))
	}
	b := env.Docker.Builds[0]
	if b.Tag != "safeagent/"+sid+":"+pid {
		t.Fatalf("unexpected tag %q", b.Tag)
	}
	if b.BuildArgs["NODE_VERSION"] != "22" {
		t.Fatalf("expected NODE_VERSION=22, got %v", b.BuildArgs)
	}
	s := strings.Index(b.Dockerfile, "stack-fragment")
	p := strings.Index(b.Dockerfile, "profile-fragment")
	e := strings.Index(b.Dockerfile, "ENTRYPOINT")
	if s < 0 || p < 0 || s > p || p > e {
		t.Fatalf("expected stack fragment, then profile fragment, before ENTRYPOINT:\n%s", b.Dockerfile)
	}
	if len(env.Docker.Runs) != 2 {
		t.Fatalf("expected two runs, got %d", len(env.Docker.Runs))
	}
}

func TestStartDoesNotRebuildExistingImage(t *testing.T) {
	env := setup(t)
	env.Docker.AddImage("safeagent/" + stackID(t, env, "node22") + ":" + profileID(t, env, "work"))

	env.MustRun("--profile", "work", "--stack", "node22")

	if len(env.Docker.Builds) != 0 {
		t.Fatalf("expected no build, got %+v", env.Docker.Builds)
	}
	if len(env.Docker.Runs) != 1 {
		t.Fatalf("expected one run, got %d", len(env.Docker.Runs))
	}
}

func TestStartRunsWithProjectMountClaudeMountWorkdirAndLabels(t *testing.T) {
	env := setup(t)
	root := env.Cwd
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
	sub := mkdir(t, filepath.Join(root, "web"))

	env.Cwd = sub
	env.MustRun()

	run := env.Docker.Runs[0]
	pid, sid := profileID(t, env, "work"), stackID(t, env, "node22")
	if run.Workdir != sub {
		t.Fatalf("workdir %q, want %q", run.Workdir, sub)
	}
	wantMounts := []docker.Mount{
		{Source: root, Target: root},
		{Source: filepath.Join(env.ConfigRoot, "profiles", pid, "claude"), Target: "/claude"},
	}
	if !slices.Equal(run.Mounts, wantMounts) {
		t.Fatalf("mounts %+v, want %+v", run.Mounts, wantMounts)
	}
	wantLabels := map[string]string{
		"safeagent.v2.project": root,
		"safeagent.v2.profile": pid,
		"safeagent.v2.stack":   sid,
	}
	if !maps.Equal(run.Labels, wantLabels) {
		t.Fatalf("labels %v, want %v", run.Labels, wantLabels)
	}
}

func TestStartWithShellRunsAShellInsteadOfClaude(t *testing.T) {
	env := setup(t)

	env.MustRun("--shell", "--profile", "work", "--stack", "node22")

	if cmd := env.Docker.Runs[0].Command; len(cmd) != 0 {
		t.Fatalf("expected the image's default shell (no command), got %v", cmd)
	}
}

func TestStartErrorsWhenProjectsProfileOrStackIsGone(t *testing.T) {
	for _, kind := range []string{"profile", "stack"} {
		t.Run(kind, func(t *testing.T) {
			env := setup(t)
			env.MustRun("project", "create", "--profile", "work", "--stack", "node22")
			writeFile(t, filepath.Join(env.ConfigRoot, kind+"s.json"), "[]")

			_, err := env.Run()

			if err == nil || !strings.Contains(err.Error(), "no longer exists") || !strings.Contains(err.Error(), "project use-"+kind) {
				t.Fatalf("expected a clear missing-%s error, got %v", kind, err)
			}
			if len(env.Docker.Runs) != 0 || len(env.Docker.Builds) != 0 {
				t.Fatalf("expected no build or run, got builds=%d runs=%d", len(env.Docker.Builds), len(env.Docker.Runs))
			}
		})
	}
}

func TestStartRejectsProfileFlagForExistingProject(t *testing.T) {
	env := setup(t)
	env.MustRun("project", "create", "--profile", "work", "--stack", "node22")

	_, err := env.Run("--profile", "work")

	if err == nil || !strings.Contains(err.Error(), "use-profile") {
		t.Fatalf("expected an error pointing at use-profile, got %v", err)
	}
	if len(env.Docker.Runs) != 0 {
		t.Fatal("expected no run")
	}
}

func TestStartWithUnknownProfileCreatesNothing(t *testing.T) {
	env := setup(t)

	_, err := env.Run("--profile", "nope", "--stack", "node22")

	if err == nil || !strings.Contains(err.Error(), `no profile named "nope"`) {
		t.Fatalf("expected unknown profile error, got %v", err)
	}
	if len(readProjects(t, env)) != 0 || len(env.Docker.Runs) != 0 {
		t.Fatal("expected no project and no run")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
