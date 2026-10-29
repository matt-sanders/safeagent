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

// ClaudeMountPath is where every Session sees the Profile's Claude config
// folder; no Mount may use it.
const ClaudeMountPath = "/claude"

// AddMount validates m and appends it to the named Profile's Mounts. The host
// path must exist; the container path must be absolute, not /claude, and not
// already used by another of the Profile's Mounts. Paths are cleaned.
func (s *Store) AddMount(name string, m Mount) (Mount, error) {
	profiles, err := s.List()
	if err != nil {
		return Mount{}, err
	}
	i := indexByName(profiles, name)
	if i < 0 {
		return Mount{}, fmt.Errorf("no profile named %q", name)
	}
	if m.HostPath == "" {
		return Mount{}, errors.New("host path is required")
	}
	host, err := filepath.Abs(m.HostPath)
	if err != nil {
		return Mount{}, err
	}
	if _, err := os.Stat(host); err != nil {
		return Mount{}, fmt.Errorf("host path %s does not exist", host)
	}
	if !filepath.IsAbs(m.ContainerPath) {
		return Mount{}, fmt.Errorf("container path %q must be absolute", m.ContainerPath)
	}
	target := filepath.Clean(m.ContainerPath)
	if target == ClaudeMountPath {
		return Mount{}, fmt.Errorf("container path %s is reserved for the Profile's Claude config", ClaudeMountPath)
	}
	for _, existing := range profiles[i].Mounts {
		if existing.ContainerPath == target {
			return Mount{}, fmt.Errorf("profile %q already has a mount at %s", name, target)
		}
	}
	added := Mount{HostPath: host, ContainerPath: target}
	profiles[i].Mounts = append(profiles[i].Mounts, added)
	return added, s.save(profiles)
}

// RemoveMount removes the named Profile's Mount at containerPath.
func (s *Store) RemoveMount(name, containerPath string) error {
	profiles, err := s.List()
	if err != nil {
		return err
	}
	i := indexByName(profiles, name)
	if i < 0 {
		return fmt.Errorf("no profile named %q", name)
	}
	target := filepath.Clean(containerPath)
	mounts := profiles[i].Mounts
	for j, m := range mounts {
		if m.ContainerPath == target {
			profiles[i].Mounts = append(mounts[:j:j], mounts[j+1:]...)
			return s.save(profiles)
		}
	}
	return fmt.Errorf("profile %q has no mount at %s", name, target)
}

func indexByName(profiles []Profile, name string) int {
	for i, p := range profiles {
		if p.Name == name {
			return i
		}
	}
	return -1
}
