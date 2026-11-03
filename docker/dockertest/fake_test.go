package dockertest_test

import (
	"slices"
	"testing"

	"safeagent/docker"
	"safeagent/docker/dockertest"
)

func TestFakeBuildRecordsAndCreatesImage(t *testing.T) {
	f := dockertest.New()

	err := f.BuildImage(docker.BuildRequest{Tag: "safeagent/s1:p1", Dockerfile: "FROM x", BuildArgs: map[string]string{"NODE_VERSION": "lts"}})
	if err != nil {
		t.Fatal(err)
	}

	if len(f.Builds) != 1 || f.Builds[0].BuildArgs["NODE_VERSION"] != "lts" {
		t.Fatalf("build not recorded: %+v", f.Builds)
	}
	if ok, _ := f.ImageExists("safeagent/s1:p1"); !ok {
		t.Fatal("built image should exist")
	}
}

func TestFakeListsOnlySafeagentImagesAndRemovesThem(t *testing.T) {
	f := dockertest.New()
	f.AddImage("safeagent/s1:p1")
	f.AddImage("safeagent/s2:p1")
	f.AddImage("ubuntu:latest")

	refs, _ := f.ListImages()
	if !slices.Equal(refs, []string{"safeagent/s1:p1", "safeagent/s2:p1"}) {
		t.Fatalf("got %v", refs)
	}

	if err := f.RemoveImage("safeagent/s1:p1"); err != nil {
		t.Fatal(err)
	}
	if f.HasImage("safeagent/s1:p1") || !slices.Equal(f.Removed, []string{"safeagent/s1:p1"}) {
		t.Fatalf("removal not applied: images=%v removed=%v", f.Images(), f.Removed)
	}
	if err := f.RemoveImage("safeagent/missing:x"); err == nil {
		t.Fatal("removing a missing image should fail")
	}
}

func TestFakeListsOnlyLabelledContainers(t *testing.T) {
	f := dockertest.New()
	f.AddContainer(docker.Container{ID: "abc", Labels: map[string]string{docker.LabelProject: "/work"}})
	f.AddContainer(docker.Container{ID: "other"})

	cs, _ := f.ListContainers()
	if len(cs) != 1 || cs[0].ID != "abc" {
		t.Fatalf("got %+v", cs)
	}
}
