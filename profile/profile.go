package profile

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Profile represents a named environment configuration.
type Profile struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	NodeVersion string `json:"node_version"`
	BunVersion  string `json:"bun_version,omitempty"`
}

// Store handles profile persistence to a JSON file.
type Store struct {
	filePath string
}

// NewStore creates a Store that reads/writes profiles at filePath.
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

// Load reads all profiles from the file.
// Returns an empty slice if the file does not exist.
func (s *Store) Load() ([]Profile, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Profile{}, nil
		}
		return nil, fmt.Errorf("failed to read profiles file: %w", err)
	}

	var profiles []Profile
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, fmt.Errorf("failed to parse profiles file: %w", err)
	}

	return profiles, nil
}

// Add appends a profile, generating an ID automatically.
// Returns an error if a profile with the same name already exists.
func (s *Store) Add(p Profile) (Profile, error) {
	profiles, err := s.Load()
	if err != nil {
		return Profile{}, err
	}

	for _, existing := range profiles {
		if existing.Name == p.Name {
			return Profile{}, fmt.Errorf("profile %q already exists", p.Name)
		}
	}

	p.ID = generateID()
	profiles = append(profiles, p)

	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return Profile{}, fmt.Errorf("failed to marshal profiles: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0600); err != nil {
		return Profile{}, fmt.Errorf("failed to write profiles file: %w", err)
	}

	return p, nil
}

// Remove deletes the profile with the given ID from the store and removes its
// data directory (~/.safeagent/profiles/{id}/).
// Returns an error if no profile with that ID exists.
func (s *Store) Remove(id string) error {
	profiles, err := s.Load()
	if err != nil {
		return err
	}

	idx := -1
	var target Profile
	for i, p := range profiles {
		if p.ID == id {
			idx = i
			target = p
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("profile %q not found", id)
	}

	profiles = append(profiles[:idx], profiles[idx+1:]...)

	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profiles: %w", err)
	}
	if err := os.WriteFile(s.filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write profiles file: %w", err)
	}

	dir := filepath.Join(filepath.Dir(s.filePath), "profiles", target.ID)
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to remove profile directory: %w", err)
	}

	return nil
}

// DockerfileExtraPath returns the path to the extra Dockerfile fragment for
// the given profile ID (~/.safeagent/profile-{id}.dockerfile).
func (s *Store) DockerfileExtraPath(profileID string) string {
	return filepath.Join(filepath.Dir(s.filePath), "profiles", profileID, "dockerfile")
}

// LoadDockerfileExtra reads the extra Dockerfile fragment for a profile.
// Returns an empty string if no file exists yet.
func (s *Store) LoadDockerfileExtra(profileID string) (string, error) {
	data, err := os.ReadFile(s.DockerfileExtraPath(profileID))
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to read extra dockerfile: %w", err)
	}
	return string(data), nil
}
