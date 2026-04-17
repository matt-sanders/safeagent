package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Project struct {
	Path      string `json:"path"`
	ProfileID string `json:"profile_id,omitempty"`
}

// Load reads projects from the JSON file at filePath.
// Returns an empty slice if the file does not exist.
func Load(filePath string) ([]Project, error) {
	data, err := os.ReadFile(filePath)
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

func save(filePath string, projects []Project) error {
	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal projects: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write projects file: %w", err)
	}

	return nil
}

// GetOrCreate returns the project for the given path, creating it if it doesn't exist.
// The bool return value is true if a new project was created.
func GetOrCreate(filePath string, path string) (Project, bool, error) {
	projects, err := Load(filePath)
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

	if err := save(filePath, projects); err != nil {
		return Project{}, false, err
	}

	return p, true, nil
}

// SetProfile sets the profile ID for the project at the given path.
func SetProfile(filePath string, path string, profileID string) error {
	projects, err := Load(filePath)
	if err != nil {
		return err
	}

	for i, p := range projects {
		if p.Path == path {
			projects[i].ProfileID = profileID
			return save(filePath, projects)
		}
	}

	return fmt.Errorf("project %q not found", path)
}
