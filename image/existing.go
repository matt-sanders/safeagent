package image

import (
	"strings"

	"safeagent/docker"
)

// Pair is an existing Image's Stack and Profile IDs.
type Pair struct {
	StackID   string
	ProfileID string
}

// ExistingWithProfile returns the existing Images (in Docker) that contain the
// Profile ID, in ListImages order.
func ExistingWithProfile(d docker.Docker, profileID string) ([]Pair, error) {
	return existing(d, func(p Pair) bool { return p.ProfileID == profileID })
}

// ExistingWithStack returns the existing Images (in Docker) that contain the
// Stack ID, in ListImages order.
func ExistingWithStack(d docker.Docker, stackID string) ([]Pair, error) {
	return existing(d, func(p Pair) bool { return p.StackID == stackID })
}

func existing(d docker.Docker, keep func(Pair) bool) ([]Pair, error) {
	refs, err := d.ListImages()
	if err != nil {
		return nil, err
	}
	var pairs []Pair
	for _, ref := range refs {
		p, ok := splitRef(ref)
		if ok && keep(p) {
			pairs = append(pairs, p)
		}
	}
	return pairs, nil
}

// splitRef is the inverse of Ref; ok is false for anything else.
func splitRef(ref string) (Pair, bool) {
	rest, ok := strings.CutPrefix(ref, docker.Repository+"/")
	if !ok {
		return Pair{}, false
	}
	stackID, profileID, ok := strings.Cut(rest, ":")
	if !ok || stackID == "" || profileID == "" || strings.ContainsAny(profileID, ":/") {
		return Pair{}, false
	}
	return Pair{StackID: stackID, ProfileID: profileID}, true
}
