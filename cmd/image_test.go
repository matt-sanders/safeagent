package cmd_test

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"safeagent/clitest"
)

func imageRef(t *testing.T, env *clitest.Env, stack, profile string) string {
	t.Helper()
	return "safeagent/" + stackID(t, env, stack) + ":" + profileID(t, env, profile)
}

func builtTags(env *clitest.Env) []string {
	var tags []string
	for _, b := range env.Docker.Builds {
		tags = append(tags, b.Tag)
	}
	slices.Sort(tags)
	return tags
}

func TestImageLsWithNoImagesPrintsEmptyState(t *testing.T) {
	env := clitest.New(t)
	env.Docker.AddImage("other/thing:latest")

	out := env.MustRun("image", "ls")

	if out != "No images.\n" {
		t.Fatalf("got %q", out)
	}
}

func TestImageLsShowsNamesProjectCountsAndUnresolvedIDs(t *testing.T) {
	env := setup(t)
	env.MustRun("profile", "create", "--name", "home")
	root := t.TempDir()
	for _, d := range []string{"a", "b"} {
		env.MustRun("project", "create", "-p", mkdir(t, filepath.Join(root, d)), "--profile", "work", "--stack", "node22")
	}
	env.Docker.AddImage(imageRef(t, env, "node22", "work"))
	env.Docker.AddImage(imageRef(t, env, "node22", "home"))
	env.Docker.AddImage("safeagent/deadbeef:" + profileID(t, env, "work"))

	out := env.MustRun("image", "ls")

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected header + 3 rows, got:\n%s", out)
	}
	if got := strings.Join(strings.Fields(lines[0]), " "); got != "IMAGE STACK PROFILE PROJECTS" {
		t.Fatalf("header = %q", lines[0])
	}
	want := map[string]string{
		"node22:work":   "node22:work node22 work 2",
		"node22:home":   "node22:home node22 home 0",
		"deadbeef:work": "deadbeef:work deadbeef work 0",
	}
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if want[fields[0]] != strings.Join(fields, " ") {
			t.Errorf("unexpected row %q", line)
		}
		delete(want, fields[0])
	}
	if len(want) != 0 {
		t.Fatalf("missing rows %v in:\n%s", want, out)
	}
}

func TestImageRmRemovesImageAndListsProjectsThatWillRebuildIt(t *testing.T) {
	env := setup(t)
	root := t.TempDir()
	a := mkdir(t, filepath.Join(root, "a"))
	b := mkdir(t, filepath.Join(root, "b"))
	env.MustRun("project", "create", "-p", a, "--profile", "work", "--stack", "node22")
	env.MustRun("project", "create", "-p", b, "--profile", "work", "--stack", "node22")
	ref := imageRef(t, env, "node22", "work")
	env.Docker.AddImage(ref)

	out := env.MustRun("image", "rm", "node22:work")

	if env.Docker.HasImage(ref) {
		t.Fatal("image still present")
	}
	want := "Removed image node22:work\n" +
		"These Projects will rebuild it at their next session:\n" +
		"  " + a + "\n" +
		"  " + b + "\n"
	if out != want {
		t.Fatalf("got:\n%s\nwant:\n%s", out, want)
	}
}

func TestImageRmWithNoProjectsJustRemoves(t *testing.T) {
	env := setup(t)
	ref := imageRef(t, env, "node22", "work")
	env.Docker.AddImage(ref)

	out := env.MustRun("image", "rm", "node22:work")

	if out != "Removed image node22:work\n" || env.Docker.HasImage(ref) {
		t.Fatalf("got %q, image present: %v", out, env.Docker.HasImage(ref))
	}
}

func TestImageRmErrors(t *testing.T) {
	env := setup(t)
	cases := map[string]string{
		"node22":      "expected <stack>:<profile>",
		"node22:nope": `no profile named "nope"`,
		"nope:work":   `no stack named "nope"`,
		"node22:work": "image node22:work is not built",
		"Node22:work": "invalid image",
	}
	for arg, wantErr := range cases {
		_, err := env.Run("image", "rm", arg)
		if err == nil || !strings.Contains(err.Error(), wantErr) {
			t.Errorf("image rm %s: err = %v, want containing %q", arg, err, wantErr)
		}
	}
	if len(env.Docker.Removed) != 0 {
		t.Fatalf("removed %v", env.Docker.Removed)
	}
}

func TestImageRebuildOneRebuildsJustThatImage(t *testing.T) {
	env := setup(t)
	env.MustRun("profile", "create", "--name", "home")
	work := imageRef(t, env, "node22", "work")
	env.Docker.AddImage(work)
	env.Docker.AddImage(imageRef(t, env, "node22", "home"))

	out := env.MustRun("image", "rebuild", "node22:work")

	if got := builtTags(env); !slices.Equal(got, []string{work}) {
		t.Fatalf("builds = %v", got)
	}
	if env.Docker.Builds[0].BuildArgs["NODE_VERSION"] != "22" {
		t.Fatalf("build args = %v", env.Docker.Builds[0].BuildArgs)
	}
	if out != "Rebuilt image node22:work\n" {
		t.Fatalf("got %q", out)
	}
}

func TestImageRebuildAllRebuildsEveryExistingImage(t *testing.T) {
	env := setup(t)
	env.MustRun("profile", "create", "--name", "home")
	env.MustRun("profile", "create", "--name", "spare")
	work := imageRef(t, env, "node22", "work")
	home := imageRef(t, env, "node22", "home")
	env.Docker.AddImage(work)
	env.Docker.AddImage(home)

	out := env.MustRun("image", "rebuild", "--all")

	want := []string{work, home}
	slices.Sort(want)
	if got := builtTags(env); !slices.Equal(got, want) {
		t.Fatalf("builds = %v, want %v", got, want)
	}
	if !strings.Contains(out, "Rebuilt image node22:work\n") || !strings.Contains(out, "Rebuilt image node22:home\n") {
		t.Fatalf("got %q", out)
	}
}

func TestImageRebuildAllSkipsImagesWhoseIDsNoLongerResolve(t *testing.T) {
	env := setup(t)
	work := imageRef(t, env, "node22", "work")
	env.Docker.AddImage(work)
	env.Docker.AddImage("safeagent/deadbeef:" + profileID(t, env, "work"))

	out := env.MustRun("image", "rebuild", "--all")

	if got := builtTags(env); !slices.Equal(got, []string{work}) {
		t.Fatalf("builds = %v", got)
	}
	if !strings.Contains(out, "Skipped deadbeef:work") {
		t.Fatalf("got %q", out)
	}
}

func TestImageRebuildAllWithNoImages(t *testing.T) {
	env := setup(t)

	out := env.MustRun("image", "rebuild", "--all")

	if out != "No images to rebuild.\n" || len(env.Docker.Builds) != 0 {
		t.Fatalf("got %q, builds %v", out, env.Docker.Builds)
	}
}

func TestImageRebuildErrors(t *testing.T) {
	env := setup(t)
	env.Docker.AddImage(imageRef(t, env, "node22", "work"))
	cases := []struct {
		args    []string
		wantErr string
	}{
		{[]string{"image", "rebuild"}, "pass an image or --all"},
		{[]string{"image", "rebuild", "node22:work", "--all"}, "pass an image or --all, not both"},
		{[]string{"image", "rebuild", "node22-work"}, "expected <stack>:<profile>"},
		{[]string{"image", "rebuild", "node22:nope"}, `no profile named "nope"`},
		{[]string{"image", "rebuild", "nope:work"}, `no stack named "nope"`},
	}
	for _, c := range cases {
		_, err := env.Run(c.args...)
		if err == nil || !strings.Contains(err.Error(), c.wantErr) {
			t.Errorf("%v: err = %v, want containing %q", c.args, err, c.wantErr)
		}
	}
	if len(env.Docker.Builds) != 0 {
		t.Fatalf("builds = %v", env.Docker.Builds)
	}
}
