package docker_test

import (
	"strings"
	"testing"

	"safeagent/docker"
)

func TestComposeDockerfilePutsStackBeforeProfileBeforeEntrypoint(t *testing.T) {
	got := docker.ComposeDockerfile("RUN echo stack", "RUN echo profile")

	stack := strings.Index(got, "RUN echo stack\n")
	profile := strings.Index(got, "RUN echo profile\n")
	entry := strings.Index(got, "ENTRYPOINT")
	if stack < 0 || profile < 0 || entry < 0 {
		t.Fatalf("missing stack, profile or ENTRYPOINT in:\n%s", got)
	}
	if !(stack < profile && profile < entry) {
		t.Fatalf("want stack < profile < ENTRYPOINT, got %d, %d, %d", stack, profile, entry)
	}
}

func TestComposeDockerfileRemovesMarkersWhenExtensionsEmpty(t *testing.T) {
	got := docker.ComposeDockerfile("  \n", "")

	if strings.Contains(got, "{{") {
		t.Fatalf("markers left in Dockerfile:\n%s", got)
	}
	if !strings.Contains(got, "ENTRYPOINT") {
		t.Fatal("ENTRYPOINT missing")
	}
}
