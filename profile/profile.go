package profile

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
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
