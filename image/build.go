package image

import (
	"errors"
	"os"

	"safeagent/docker"
	"safeagent/profile"
	"safeagent/stack"
)

// NodeVersionArg is the build arg that carries a Stack's node version.
const NodeVersionArg = "NODE_VERSION"

// Dockerfile composes the full Dockerfile for the Image of st and p: the base
// Dockerfile with the Stack's fragment spliced in, then the Profile's. A
// missing fragment file counts as empty.
func Dockerfile(configRoot string, st stack.Stack, p profile.Profile) (string, error) {
	stackExtra, err := readFragment(stack.NewStore(configRoot).DockerfilePath(st.ID))
	if err != nil {
		return "", err
	}
	profileExtra, err := readFragment(profile.NewStore(configRoot).DockerfilePath(p.ID))
	if err != nil {
		return "", err
	}
	return docker.ComposeDockerfile(stackExtra, profileExtra), nil
}

// Build builds (or rebuilds) the Image of st and p, tagged Ref(st.ID, p.ID),
// with the Stack's node version as a build arg. Nothing secret is ever part of
// the build: credentials are mounted at run time instead.
func Build(d docker.Docker, configRoot string, st stack.Stack, p profile.Profile) error {
	dockerfile, err := Dockerfile(configRoot, st, p)
	if err != nil {
		return err
	}
	return d.BuildImage(docker.BuildRequest{
		Tag:        Ref(st.ID, p.ID),
		Dockerfile: dockerfile,
		BuildArgs:  map[string]string{NodeVersionArg: st.Node},
	})
}

func readFragment(path string) (string, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	return string(data), err
}
