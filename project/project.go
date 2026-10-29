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
)

// ErrNotFound is returned (wrapped) when no Project matches a lookup.
var ErrNotFound = errors.New("no project found")

// Project is one stored Project record.
type Project struct {
	Path       string   `json:"path"`
	ProfileID  string   `json:"profile_id"`
	StackID    string   `json:"stack_id"`
	Exclusions []string `json:"exclusions"`
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
