package docker

import (
	_ "embed"
	"strings"
)

// BaseDockerfile is the embedded base Dockerfile every Image is built from.
//
//go:embed Dockerfile
var BaseDockerfile string

const (
	stackMarker   = "# {{STACK_EXTENSIONS}}\n"
	profileMarker = "# {{PROFILE_EXTENSIONS}}\n"
)

// ComposeDockerfile splices a Stack's and a Profile's Dockerfile fragments into
// the base Dockerfile, Stack first, then Profile, both before ENTRYPOINT. An
// empty fragment removes its marker line cleanly.
func ComposeDockerfile(stackExtra, profileExtra string) string {
	out := strings.Replace(BaseDockerfile, stackMarker, fragment(stackExtra), 1)
	return strings.Replace(out, profileMarker, fragment(profileExtra), 1)
}

func fragment(extra string) string {
	trimmed := strings.TrimSpace(extra)
	if trimmed == "" {
		return ""
	}
	return trimmed + "\n"
}
