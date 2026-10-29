// Package profile stores Profiles: the "who" of a Session. Records live in
// profiles.json under the v2 config root, and each Profile owns a folder at
// profiles/<id>/ holding its Claude config folder and Dockerfile fragment.
package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"safeagent/ident"
)

// Mount is a host path a Profile makes available inside every Session.
type Mount struct {
	HostPath      string `json:"host_path"`
	ContainerPath string `json:"container_path"`
}

// Profile is one stored Profile record.
type Profile struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Mounts []Mount `json:"mounts"`
}

// Store reads and writes Profiles under a config root.
type Store struct {
	root string
}

// NewStore returns a Store rooted at the v2 config root.
func NewStore(root string) *Store {
	return &Store{root: root}
}

func (s *Store) recordsPath() string {
	return filepath.Join(s.root, "profiles.json")
}

// Dir is the Profile's own folder on the host.
func (s *Store) Dir(id string) string {
	return filepath.Join(s.root, "profiles", id)
}

// ClaudeDir is the Profile's Claude config folder (mounted at /claude).
func (s *Store) ClaudeDir(id string) string {
	return filepath.Join(s.Dir(id), "claude")
}

// DockerfilePath is the Profile's Dockerfile fragment.
func (s *Store) DockerfilePath(id string) string {
	return filepath.Join(s.Dir(id), "Dockerfile")
}

// List returns every Profile in creation order.
func (s *Store) List() ([]Profile, error) {
	data, err := os.ReadFile(s.recordsPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading profiles: %w", err)
	}
	var profiles []Profile
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, fmt.Errorf("parsing profiles: %w", err)
	}
	return profiles, nil
}

// FindByName returns the Profile with the given Name, or an error naming it
// if there is none.
func (s *Store) FindByName(name string) (Profile, error) {
	profiles, err := s.List()
	if err != nil {
		return Profile{}, err
	}
	for _, p := range profiles {
		if p.Name == name {
			return p, nil
		}
	}
	return Profile{}, fmt.Errorf("no profile named %q", name)
}

// Create validates name, then stores a new Profile with a fresh ID and
// creates its folder, Claude config folder and empty Dockerfile fragment.
func (s *Store) Create(name string) (Profile, error) {
	if err := ident.ValidateName(name); err != nil {
		return Profile{}, err
	}
	profiles, err := s.List()
	if err != nil {
		return Profile{}, err
	}
	for _, p := range profiles {
		if p.Name == name {
			return Profile{}, fmt.Errorf("a profile named %q already exists", name)
		}
	}

	p := Profile{ID: ident.NewID(), Name: name, Mounts: []Mount{}}
	if err := os.MkdirAll(s.ClaudeDir(p.ID), 0o700); err != nil {
		return Profile{}, fmt.Errorf("creating profile folder: %w", err)
	}
	if err := os.WriteFile(s.DockerfilePath(p.ID), nil, 0o644); err != nil {
		return Profile{}, fmt.Errorf("creating profile Dockerfile: %w", err)
	}
	if err := s.save(append(profiles, p)); err != nil {
		return Profile{}, err
	}
	return p, nil
}

func (s *Store) save(profiles []Profile) error {
	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(s.root, 0o700); err != nil {
		return fmt.Errorf("creating config root: %w", err)
	}
	if err := os.WriteFile(s.recordsPath(), data, 0o600); err != nil {
		return fmt.Errorf("writing profiles: %w", err)
	}
	return nil
}

// FindByID returns the Profile with the given ID, or an error if no Profile
// has it (e.g. it was removed after a Project chose it).
func (s *Store) FindByID(id string) (Profile, error) {
	profiles, err := s.List()
	if err != nil {
		return Profile{}, err
	}
	for _, p := range profiles {
		if p.ID == id {
			return p, nil
		}
	}
	return Profile{}, fmt.Errorf("no profile with ID %q", id)
}

// Remove deletes the Profile record with the given ID and its folder.
func (s *Store) Remove(id string) error {
	profiles, err := s.List()
	if err != nil {
		return err
	}
	kept := make([]Profile, 0, len(profiles))
	for _, p := range profiles {
		if p.ID != id {
			kept = append(kept, p)
		}
	}
	if len(kept) == len(profiles) {
		return fmt.Errorf("no profile with ID %q", id)
	}
	if err := s.save(kept); err != nil {
		return err
	}
	if err := os.RemoveAll(s.Dir(id)); err != nil {
		return fmt.Errorf("removing profile folder: %w", err)
	}
	return nil
}
