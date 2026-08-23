package profile

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

var (
	nameStyle   = lipgloss.NewStyle().Bold(true)
	detailStyle = lipgloss.NewStyle().Faint(true)
)

// Service provides profile operations using a Store for persistence.
type Service struct {
	store *Store
}

// NewService creates a Service backed by the given Store.
func NewService(store *Store) *Service {
	return &Service{store: store}
}

// List returns all profiles.
func (s *Service) List() ([]Profile, error) {
	return s.store.Load()
}

// Create adds a new profile with the given name and node version.
func (s *Service) Create(name, nodeVersion string) (Profile, error) {
	return s.store.Add(Profile{Name: name, NodeVersion: nodeVersion})
}

// Remove deletes the profile with the given ID and its data directory.
func (s *Service) Remove(id string) error {
	return s.store.Remove(id)
}

// GetByID returns the profile with the given ID.
func (s *Service) GetByID(id string) (Profile, error) {
	profiles, err := s.store.Load()
	if err != nil {
		return Profile{}, err
	}
	for _, p := range profiles {
		if p.ID == id {
			return p, nil
		}
	}
	return Profile{}, fmt.Errorf("profile %q not found", id)
}

// DockerfileExtraPath returns the filesystem path for the profile's extra Dockerfile fragment.
func (s *Service) DockerfileExtraPath(profileID string) string {
	return s.store.DockerfileExtraPath(profileID)
}

// LoadDockerfileExtra returns the extra Dockerfile fragment for a profile,
// or an empty string if none has been written yet.
func (s *Service) LoadDockerfileExtra(profileID string) (string, error) {
	return s.store.LoadDockerfileExtra(profileID)
}

// FormatOption returns a styled string for displaying a profile.
func (s *Service) FormatOption(p Profile) string {
	name := nameStyle.Render(p.Name)
	detail := fmt.Sprintf("  id: %s  node: %s", p.ID, p.NodeVersion)
	if p.BunVersion != "" {
		detail += fmt.Sprintf("  bun: %s", p.BunVersion)
	}
	return name + "\n" + detailStyle.Render(detail)
}
