// Package image derives Image identities. An Image is never stored: it is the
// pair (Stack, Profile). Docker knows it as "safeagent/<stack-id>:<profile-id>";
// the CLI writes it as "<stack-name>:<profile-name>".
package image

import (
	"fmt"
	"strings"

	"safeagent/docker"
	"safeagent/ident"
)

// Ref returns the Docker reference of the Image for a Stack and Profile ID.
func Ref(stackID, profileID string) string {
	return docker.Repository + "/" + stackID + ":" + profileID
}

// Name returns the CLI form of an Image: "<stack-name>:<profile-name>".
func Name(stackName, profileName string) string {
	return stackName + ":" + profileName
}

// ParseName splits the CLI form "<stack-name>:<profile-name>" into its Stack
// and Profile Names, validating both.
func ParseName(s string) (stackName, profileName string, err error) {
	stackName, profileName, ok := strings.Cut(s, ":")
	if !ok {
		return "", "", fmt.Errorf("invalid image %q: expected <stack>:<profile>", s)
	}
	if err := ident.ValidateName(stackName); err != nil {
		return "", "", fmt.Errorf("invalid image %q: stack: %w", s, err)
	}
	if err := ident.ValidateName(profileName); err != nil {
		return "", "", fmt.Errorf("invalid image %q: profile: %w", s, err)
	}
	return stackName, profileName, nil
}
