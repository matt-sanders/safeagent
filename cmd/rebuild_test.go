package cmd_test

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"safeagent/docker"
)

func buildTags(builds []docker.BuildRequest) []string {
	var tags []string
	for _, b := range builds {
		tags = append(tags, b.Tag)
	}
	slices.Sort(tags)
	return tags
}

func TestProfileRebuildRebuildsOnlyExistingImagesWithThatProfile(t *testing.T) {
	env := setup(t)
	env.MustRun("profile", "create", "--name", "home")
	env.MustRun("stack", "create", "--name", "py", "--node", "20")
	work, home := profileID(t, env, "work"), profileID(t, env, "home")
	node22, py := stackID(t, env, "node22"), stackID(t, env, "py")
	env.Docker.AddImage("safeagent/" + node22 + ":" + work)
	env.Docker.AddImage("safeagent/" + py + ":" + work)
	env.Docker.AddImage("safeagent/" + node22 + ":" + home)
	writeFile(t, filepath.Join(env.ConfigRoot, "profiles", work, "Dockerfile"), "RUN echo edited-profile\n")

	out := env.MustRun("profile", "rebuild", "work")

	want := []string{"safeagent/" + node22 + ":" + work, "safeagent/" + py + ":" + work}
	slices.Sort(want)
	if got := buildTags(env.Docker.Builds); !slices.Equal(got, want) {
		t.Fatalf("rebuilt %v, want %v", got, want)
	}
	for _, b := range env.Docker.Builds {
		if !strings.Contains(b.Dockerfile, "edited-profile") {
			t.Fatalf("build %s is missing the edited fragment:\n%s", b.Tag, b.Dockerfile)
		}
	}
	nodes := map[string]string{}
	for _, b := range env.Docker.Builds {
		nodes[b.Tag] = b.BuildArgs["NODE_VERSION"]
	}
	if nodes["safeagent/"+py+":"+work] != "20" || nodes["safeagent/"+node22+":"+work] != "22" {
		t.Fatalf("each build should use its Stack's node version, got %v", nodes)
	}
	if !strings.Contains(out, "node22:work") || !strings.Contains(out, "py:work") {
		t.Fatalf("expected rebuilt Images named in output, got:\n%s", out)
	}
}

func TestStackRebuildRebuildsOnlyExistingImagesWithThatStack(t *testing.T) {
	env := setup(t)
	env.MustRun("profile", "create", "--name", "home")
	env.MustRun("profile", "create", "--name", "unused")
	env.MustRun("stack", "create", "--name", "py", "--node", "20")
	work, home := profileID(t, env, "work"), profileID(t, env, "home")
	node22, py := stackID(t, env, "node22"), stackID(t, env, "py")
	env.Docker.AddImage("safeagent/" + node22 + ":" + work)
	env.Docker.AddImage("safeagent/" + node22 + ":" + home)
	env.Docker.AddImage("safeagent/" + py + ":" + work)
	writeFile(t, filepath.Join(env.ConfigRoot, "stacks", node22, "Dockerfile"), "RUN echo edited-stack\n")

	out := env.MustRun("stack", "rebuild", "node22")

	want := []string{"safeagent/" + node22 + ":" + work, "safeagent/" + node22 + ":" + home}
	slices.Sort(want)
	if got := buildTags(env.Docker.Builds); !slices.Equal(got, want) {
		t.Fatalf("rebuilt %v, want %v", got, want)
	}
	for _, b := range env.Docker.Builds {
		if !strings.Contains(b.Dockerfile, "edited-stack") || b.BuildArgs["NODE_VERSION"] != "22" {
			t.Fatalf("build %s missing edited fragment or node version: %v\n%s", b.Tag, b.BuildArgs, b.Dockerfile)
		}
	}
	if !strings.Contains(out, "node22:work") || !strings.Contains(out, "node22:home") {
		t.Fatalf("expected rebuilt Images named in output, got:\n%s", out)
	}
}

func TestRebuildWithNoMatchingImagesBuildsNothing(t *testing.T) {
	for _, args := range [][]string{{"profile", "rebuild", "home"}, {"stack", "rebuild", "py"}} {
		env := setup(t)
		env.MustRun("profile", "create", "--name", "home")
		env.MustRun("stack", "create", "--name", "py", "--node", "20")
		// Only node22:work exists; neither home nor py has an Image yet.
		env.Docker.AddImage("safeagent/" + stackID(t, env, "node22") + ":" + profileID(t, env, "work"))

		out := env.MustRun(args...)

		if len(env.Docker.Builds) != 0 {
			t.Fatalf("%v: expected no builds, got %v", args, buildTags(env.Docker.Builds))
		}
		if !strings.Contains(out, "No existing images") {
			t.Fatalf("%v: expected a nothing-to-rebuild message, got:\n%s", args, out)
		}
	}
}

func TestRebuildAndDockerfileRejectUnknownNames(t *testing.T) {
	env := setup(t)
	t.Setenv("EDITOR", "true")
	cases := map[string][]string{
		`no profile named "nope"`: {"profile", "rebuild", "nope"},
		`no stack named "nope"`:   {"stack", "rebuild", "nope"},
	}
	cases2 := map[string][]string{
		`no profile named "nope"`: {"profile", "dockerfile", "nope"},
		`no stack named "nope"`:   {"stack", "dockerfile", "nope"},
	}
	for _, cs := range []map[string][]string{cases, cases2} {
		for want, args := range cs {
			_, err := env.Run(args...)
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("%v: expected error containing %q, got %v", args, want, err)
			}
		}
	}
	if len(env.Docker.Builds) != 0 {
		t.Fatalf("expected no builds, got %v", buildTags(env.Docker.Builds))
	}
}
