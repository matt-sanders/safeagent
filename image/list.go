package image

import (
	"strings"

	"safeagent/docker"
)

// ListForStack returns every existing safeagent Image built for the Stack
// with the given ID (any Profile).
func ListForStack(d docker.Docker, stackID string) ([]string, error) {
	return listMatching(d, func(s, _ string) bool { return s == stackID })
}

// ListForProfile returns every existing safeagent Image built for the Profile
// with the given ID (any Stack).
func ListForProfile(d docker.Docker, profileID string) ([]string, error) {
	return listMatching(d, func(_, p string) bool { return p == profileID })
}

func listMatching(d docker.Docker, match func(stackID, profileID string) bool) ([]string, error) {
	refs, err := d.ListImages()
	if err != nil {
		return nil, err
	}
	var out []string
	for _, ref := range refs {
		rest, ok := strings.CutPrefix(ref, docker.Repository+"/")
		if !ok {
			continue
		}
		stackID, profileID, ok := strings.Cut(rest, ":")
		if ok && match(stackID, profileID) {
			out = append(out, ref)
		}
	}
	return out, nil
}
