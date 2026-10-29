// Package ident holds the Name rules and ID generation shared by Profiles and
// Stacks.
package ident

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
)

var namePattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

// ValidateName reports whether name is a valid Name: non-empty and made only
// of lowercase letters, digits, '-' and '_'.
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if !namePattern.MatchString(name) {
		return fmt.Errorf("invalid name %q: use only lowercase letters, digits, '-' and '_'", name)
	}
	return nil
}

// NewID returns a fresh 8-character hex ID.
func NewID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("ident: crypto/rand failed: %v", err))
	}
	return hex.EncodeToString(b)
}
