package auth

import (
	"fmt"
	"regexp"
)

// DefaultName is the reserved identity that always exists and cannot be removed.
const DefaultName = "default"

var namePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// Identity is a named Claude Code auth identity.
type Identity struct {
	Name     string
	LoggedIn bool
}

// ValidateName reports whether name is a legal identity name.
func ValidateName(name string) error {
	if !namePattern.MatchString(name) {
		return fmt.Errorf("invalid identity name %q: use lowercase letters, digits, and dashes (must start with a letter or digit)", name)
	}
	return nil
}

// Resolve returns name, or DefaultName if name is empty.
func Resolve(name string) string {
	if name == "" {
		return DefaultName
	}
	return name
}
