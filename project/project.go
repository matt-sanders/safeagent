package project

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Project represents a directory-based project configuration.
type Project struct {
	ID         string   `json:"id"`
	Path       string   `json:"path"`
	ProfileID  string   `json:"profile_id,omitempty"`
	AuthName   string   `json:"auth_name,omitempty"`
	Exclusions []string `json:"exclusions,omitempty"`
}

// Store handles project persistence to a JSON file.
type Store struct {
	filePath string
}

// NewStore creates a Store that reads/writes projects at filePath.
func NewStore(filePath string) *Store {
	return &Store{filePath: filePath}
}

func generateID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("failed to generate random ID: %v", err))
	}
	return hex.EncodeToString(b)
}

// Load reads all projects from the file.
// Returns an empty slice if the file does not exist.
func (s *Store) Load() ([]Project, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Project{}, nil
		}
		return nil, fmt.Errorf("failed to read projects file: %w", err)
	}

	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("failed to parse projects file: %w", err)
	}

	return projects, nil
}

func (s *Store) save(projects []Project) error {
	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal projects: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write projects file: %w", err)
	}

	return nil
}

// GetOrCreate returns the project for the given path, creating it if it doesn't exist.
// The bool return value is true if a new project was created.
func (s *Store) GetOrCreate(path string) (Project, bool, error) {
	projects, err := s.Load()
	if err != nil {
		return Project{}, false, err
	}

	for _, p := range projects {
		if p.Path == path {
			return p, false, nil
		}
	}

	p := Project{ID: generateID(), Path: path}
	projects = append(projects, p)

	if err := s.save(projects); err != nil {
		return Project{}, false, err
	}

	return p, true, nil
}

// FindForCwd returns the project whose Path is the closest ancestor of cwd
// (or cwd itself). If no project matches any ancestor, returns ok=false.
// This method does not create projects.
func (s *Store) FindForCwd(cwd string) (Project, bool, error) {
	projects, err := s.Load()
	if err != nil {
		return Project{}, false, err
	}

	cur := filepath.Clean(cwd)
	for {
		for _, p := range projects {
			if p.Path == cur {
				return p, true, nil
			}
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return Project{}, false, nil
		}
		cur = parent
	}
}

// SetProfile sets the profile ID for the project at the given path.
func (s *Store) SetProfile(path string, profileID string) error {
	projects, err := s.Load()
	if err != nil {
		return err
	}

	for i, p := range projects {
		if p.Path == path {
			projects[i].ProfileID = profileID
			return s.save(projects)
		}
	}

	return fmt.Errorf("project %q not found", path)
}

// SetAuth sets the auth identity name for the project at the given path.
func (s *Store) SetAuth(path string, name string) error {
	projects, err := s.Load()
	if err != nil {
		return err
	}

	for i, p := range projects {
		if p.Path == path {
			projects[i].AuthName = name
			return s.save(projects)
		}
	}

	return fmt.Errorf("project %q not found", path)
}

// AddExclusion adds an exclusion to the project at the given path.
// Returns an error if the exclusion already exists.
func (s *Store) AddExclusion(path string, exclusion string) error {
	projects, err := s.Load()
	if err != nil {
		return err
	}

	for i, p := range projects {
		if p.Path == path {
			for _, e := range p.Exclusions {
				if e == exclusion {
					return fmt.Errorf("exclusion %q already exists", exclusion)
				}
			}
			projects[i].Exclusions = append(projects[i].Exclusions, exclusion)
			return s.save(projects)
		}
	}

	return fmt.Errorf("project %q not found", path)
}

// RemoveExclusion removes an exclusion from the project at the given path.
// Returns an error if the exclusion does not exist.
func (s *Store) RemoveExclusion(path string, exclusion string) error {
	projects, err := s.Load()
	if err != nil {
		return err
	}

	for i, p := range projects {
		if p.Path == path {
			for j, e := range p.Exclusions {
				if e == exclusion {
					projects[i].Exclusions = append(p.Exclusions[:j], p.Exclusions[j+1:]...)
					return s.save(projects)
				}
			}
			return fmt.Errorf("exclusion %q not found", exclusion)
		}
	}

	return fmt.Errorf("project %q not found", path)
}
