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

// FormatOption returns a styled string for displaying a profile.
func (s *Service) FormatOption(p Profile) string {
	name := nameStyle.Render(p.Name)
	details := detailStyle.Render(fmt.Sprintf("  id: %s  node: %s", p.ID, p.NodeVersion))
	return name + "\n" + details
}
