package image

import (
	"strings"

	"safeagent/docker"
)

// Existing is an Image that has actually been built: a reference in the
// safeagent repository, split back into its Stack and Profile IDs.
type Existing struct {
	Ref       string
	StackID   string
	ProfileID string
}

// ParseRef splits a Docker reference "safeagent/<stack-id>:<profile-id>" into
// its IDs. ok is false for anything that isn't a safeagent Image reference.
func ParseRef(ref string) (stackID, profileID string, ok bool) {
	rest, ok := strings.CutPrefix(ref, docker.Repository+"/")
	if !ok {
		return "", "", false
	}
	stackID, profileID, ok = strings.Cut(rest, ":")
	if !ok || stackID == "" || profileID == "" || strings.ContainsAny(profileID, ":/") || strings.Contains(stackID, "/") {
		return "", "", false
	}
	return stackID, profileID, true
}

// ListExisting returns every built safeagent Image, in the order Docker lists
// them. References that don't parse as safeagent Images are skipped.
func ListExisting(d docker.Docker) ([]Existing, error) {
	refs, err := d.ListImages()
	if err != nil {
		return nil, err
	}
	var out []Existing
	for _, ref := range refs {
		stackID, profileID, ok := ParseRef(ref)
		if !ok {
			continue
		}
		out = append(out, Existing{Ref: ref, StackID: stackID, ProfileID: profileID})
	}
	return out, nil
}
