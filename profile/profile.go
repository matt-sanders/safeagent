package profile

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Profile struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	NodeVersion string `json:"node_version"`
}

// generateID returns an 8-character random hex string.
func generateID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("failed to generate random ID: %v", err))
	}
	return hex.EncodeToString(b)
}

// Load reads profiles from the JSON file at filePath.
// Returns an empty slice if the file does not exist.
func Load(filePath string) ([]Profile, error) {
	data, err := os.ReadFile(filePath)
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

// Add appends a profile to the JSON file at filePath.
// Returns an error if a profile with the same name already exists.
func Add(filePath string, p Profile) error {
	profiles, err := Load(filePath)
	if err != nil {
		return err
	}

	for _, existing := range profiles {
		if existing.Name == p.Name {
			return fmt.Errorf("profile %q already exists", p.Name)
		}
	}

	p.ID = generateID()
	profiles = append(profiles, p)

	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profiles: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write profiles file: %w", err)
	}

	return nil
}
