// Package project stores Projects: the "where" of a Session. Records live in
// projects.json under the v2 config root. A Project is identified by its
// path (absolute and cleaned) and references its Profile and Stack by ID.
package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"safeagent/profile"
)

// ErrNotFound is returned (wrapped) when no Project matches a lookup.
var ErrNotFound = errors.New("no project found")

// Project is one stored Project record.
type Project struct {
	Path       string          `json:"path"`
	ProfileID  string          `json:"profile_id"`
	StackID    string          `json:"stack_id"`
	Exclusions []string        `json:"exclusions"`
	Mounts     []profile.Mount `json:"mounts"`
}

// Store reads and writes Projects under a config root.
type Store struct {
	root string
}

// NewStore returns a Store rooted at the v2 config root.
func NewStore(root string) *Store {
	return &Store{root: root}
}

func (s *Store) recordsPath() string {
	return filepath.Join(s.root, "projects.json")
}

// List returns every Project in creation order.
func (s *Store) List() ([]Project, error) {
	data, err := os.ReadFile(s.recordsPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading projects: %w", err)
	}
	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("parsing projects: %w", err)
	}
	return projects, nil
}

// Get returns the Project at exactly path (which must be absolute and
// cleaned). The error wraps ErrNotFound if there is none.
func (s *Store) Get(path string) (Project, error) {
	projects, err := s.List()
	if err != nil {
		return Project{}, err
	}
	for _, p := range projects {
		if p.Path == path {
			return p, nil
		}
	}
	return Project{}, fmt.Errorf("%w at %s", ErrNotFound, path)
}

// FindForDir returns the Project whose path is dir or its closest ancestor,
// so the innermost of nested Projects wins. dir must be absolute and cleaned.
// The error wraps ErrNotFound if no Project contains dir.
func (s *Store) FindForDir(dir string) (Project, error) {
	projects, err := s.List()
	if err != nil {
		return Project{}, err
	}
	var best Project
	found := false
	for _, p := range projects {
		if contains(p.Path, dir) && (!found || len(p.Path) > len(best.Path)) {
			best, found = p, true
		}
	}
	if !found {
		return Project{}, fmt.Errorf("%w for %s or any parent directory", ErrNotFound, dir)
	}
	return best, nil
}

// contains reports whether dir is base or inside it.
func contains(base, dir string) bool {
	rel, err := filepath.Rel(base, dir)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// Create stores p. It refuses if a Project already exists at exactly p.Path;
// a Project nested inside another is allowed.
func (s *Store) Create(p Project) error {
	projects, err := s.List()
	if err != nil {
		return err
	}
	for _, existing := range projects {
		if existing.Path == p.Path {
			return fmt.Errorf("a project already exists at %s", p.Path)
		}
	}
	if p.Exclusions == nil {
		p.Exclusions = []string{}
	}
	if p.Mounts == nil {
		p.Mounts = []profile.Mount{}
	}
	return s.save(append(projects, p))
}

// Remove forgets the Project at exactly path. It never touches files on disk
// other than the records file. The error wraps ErrNotFound if there is none.
func (s *Store) Remove(path string) error {
	projects, err := s.List()
	if err != nil {
		return err
	}
	for i, p := range projects {
		if p.Path == path {
			return s.save(append(projects[:i], projects[i+1:]...))
		}
	}
	return fmt.Errorf("%w at %s", ErrNotFound, path)
}

// Update replaces the stored Project at exactly p.Path with p. The error
// wraps ErrNotFound if there is none.
func (s *Store) Update(p Project) error {
	projects, err := s.List()
	if err != nil {
		return err
	}
	for i, existing := range projects {
		if existing.Path == p.Path {
			projects[i] = p
			return s.save(projects)
		}
	}
	return fmt.Errorf("%w at %s", ErrNotFound, p.Path)
}

// CountByProfile returns how many Projects use each Profile ID.
func (s *Store) CountByProfile() (map[string]int, error) {
	return s.count(func(p Project) string { return p.ProfileID })
}

// CountByStack returns how many Projects use each Stack ID.
func (s *Store) CountByStack() (map[string]int, error) {
	return s.count(func(p Project) string { return p.StackID })
}

func (s *Store) count(key func(Project) string) (map[string]int, error) {
	projects, err := s.List()
	if err != nil {
		return nil, err
	}
	counts := map[string]int{}
	for _, p := range projects {
		counts[key(p)]++
	}
	return counts, nil
}

func (s *Store) save(projects []Project) error {
	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(s.root, 0o700); err != nil {
		return fmt.Errorf("creating config root: %w", err)
	}
	if err := os.WriteFile(s.recordsPath(), data, 0o600); err != nil {
		return fmt.Errorf("writing projects: %w", err)
	}
	return nil
}

// CleanExclusion validates an Exclusion and returns it in canonical form.
// Exclusions are relative to the directory a Session starts in, so absolute
// paths and paths that climb out of it are rejected.
func CleanExclusion(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("exclusion path must not be empty")
	}
	if filepath.IsAbs(path) {
		return "", fmt.Errorf("exclusion %q must be relative to the session's start directory, not absolute", path)
	}
	clean := filepath.Clean(path)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("exclusion %q must name a path inside the session's start directory", path)
	}
	return clean, nil
}

// AddExclusion adds an Exclusion (cleaned with CleanExclusion) to the Project
// at exactly projectPath, rejecting one that is already there. It returns the
// stored form.
func (s *Store) AddExclusion(projectPath, exclusion string) (string, error) {
	clean, err := CleanExclusion(exclusion)
	if err != nil {
		return "", err
	}
	p, err := s.Get(projectPath)
	if err != nil {
		return "", err
	}
	for _, e := range p.Exclusions {
		if e == clean {
			return "", fmt.Errorf("%s is already excluded in project %s", clean, p.Path)
		}
	}
	p.Exclusions = append(p.Exclusions, clean)
	return clean, s.Update(p)
}

// RemoveExclusion removes an Exclusion from the Project at exactly
// projectPath, erroring if it isn't there. It returns the stored form.
func (s *Store) RemoveExclusion(projectPath, exclusion string) (string, error) {
	clean, err := CleanExclusion(exclusion)
	if err != nil {
		return "", err
	}
	p, err := s.Get(projectPath)
	if err != nil {
		return "", err
	}
	for i, e := range p.Exclusions {
		if e == clean {
			p.Exclusions = append(p.Exclusions[:i], p.Exclusions[i+1:]...)
			return clean, s.Update(p)
		}
	}
	return "", fmt.Errorf("%s is not excluded in project %s", clean, p.Path)
}

// AddMount validates m (see profile.CleanMount) and appends it to the Mounts of
// the Project at exactly projectPath, rejecting the Project's own path (where
// the Project itself is mounted) and a container path another of its Mounts
// already uses. It returns the stored form.
func (s *Store) AddMount(projectPath string, m profile.Mount) (profile.Mount, error) {
	added, err := profile.CleanMount(m)
	if err != nil {
		return profile.Mount{}, err
	}
	p, err := s.Get(projectPath)
	if err != nil {
		return profile.Mount{}, err
	}
	if added.ContainerPath == p.Path {
		return profile.Mount{}, fmt.Errorf("container path %s is the project path, where the project itself is mounted", p.Path)
	}
	for _, existing := range p.Mounts {
		if existing.ContainerPath == added.ContainerPath {
			return profile.Mount{}, fmt.Errorf("project %s already has a mount at %s", p.Path, added.ContainerPath)
		}
	}
	p.Mounts = append(p.Mounts, added)
	return added, s.Update(p)
}

// RemoveMount removes the Mount at containerPath from the Project at exactly
// projectPath, erroring if it isn't there.
func (s *Store) RemoveMount(projectPath, containerPath string) error {
	p, err := s.Get(projectPath)
	if err != nil {
		return err
	}
	target := filepath.Clean(containerPath)
	for i, m := range p.Mounts {
		if m.ContainerPath == target {
			p.Mounts = append(p.Mounts[:i], p.Mounts[i+1:]...)
			return s.Update(p)
		}
	}
	return fmt.Errorf("project %s has no mount at %s", p.Path, target)
}
