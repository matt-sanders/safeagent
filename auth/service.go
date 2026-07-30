package auth

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Service manages auth identities stored as directories under a root
// (typically ~/.safeagent/auth).
type Service struct {
	root string
}

// NewService creates a Service rooted at the given auth directory.
func NewService(root string) *Service {
	return &Service{root: root}
}

// DirFor returns the host .claude directory for an identity, which is what
// gets mounted to /claude in the container.
func (s *Service) DirFor(name string) string {
	return filepath.Join(s.root, name, ".claude")
}

func (s *Service) isLoggedIn(name string) bool {
	_, err := os.Stat(filepath.Join(s.DirFor(name), ".credentials.json"))
	return err == nil
}

// List returns all identities sorted by name. Returns an empty slice if the
// auth root does not exist yet.
func (s *Service) List() ([]Identity, error) {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		if os.IsNotExist(err) {
			return []Identity{}, nil
		}
		return nil, fmt.Errorf("failed to read auth directory: %w", err)
	}

	var ids []Identity
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		ids = append(ids, Identity{Name: e.Name(), LoggedIn: s.isLoggedIn(e.Name())})
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i].Name < ids[j].Name })
	return ids, nil
}

// Exists reports whether an identity directory exists.
func (s *Service) Exists(name string) (bool, error) {
	info, err := os.Stat(filepath.Join(s.root, name))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to inspect identity %q: %w", name, err)
	}
	return info.IsDir(), nil
}

// Create makes a new identity directory. Errors if the name is invalid or the
// identity already exists.
func (s *Service) Create(name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}
	exists, err := s.Exists(name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("identity %q already exists", name)
	}
	if err := os.MkdirAll(s.DirFor(name), 0700); err != nil {
		return fmt.Errorf("failed to create identity %q: %w", name, err)
	}
	return nil
}

// EnsureDir makes sure an identity's .claude directory exists. It is used at
// session start so a selected-but-not-yet-created identity self-heals.
func (s *Service) EnsureDir(name string) error {
	if err := os.MkdirAll(s.DirFor(name), 0700); err != nil {
		return fmt.Errorf("failed to ensure identity %q: %w", name, err)
	}
	return nil
}

// Remove deletes an identity directory. It refuses to remove the default
// identity and errors if the identity does not exist.
func (s *Service) Remove(name string) error {
	if name == DefaultName {
		return fmt.Errorf("cannot remove the %q identity", DefaultName)
	}
	exists, err := s.Exists(name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("identity %q not found", name)
	}
	if err := os.RemoveAll(filepath.Join(s.root, name)); err != nil {
		return fmt.Errorf("failed to remove identity %q: %w", name, err)
	}
	return nil
}
