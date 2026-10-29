// Package ident validates the Names users give Profiles and Stacks and
// generates the hidden IDs that back them.
package ident

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
)

var namePattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

// ValidateName reports whether name is a valid Name: non-empty, using only
// lowercase letters, digits, '-' and '_'.
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("name must not be empty")
	}
	if !namePattern.MatchString(name) {
		return fmt.Errorf("invalid name %q: use only lowercase letters, digits, '-' and '_'", name)
	}
	return nil
}

// NewID returns a random 8-character hex ID.
func NewID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	return hex.EncodeToString(b)
}
