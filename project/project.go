package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// Project represents a directory-based project configuration.
type Project struct {
	Path      string `json:"path"`
	ProfileID string `json:"profile_id,omitempty"`
}

// Store handles project persistence to a JSON file.
type Store struct {
	filePath string
}

// NewStore creates a Store that reads/writes projects at filePath.
func NewStore(filePath string) *Store {
	return &Store{filePath: filePath}
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

	p := Project{Path: path}
	projects = append(projects, p)

	if err := s.save(projects); err != nil {
		return Project{}, false, err
	}

	return p, true, nil
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
