package project

// Service provides project operations using a Store for persistence.
type Service struct {
	store *Store
}

// NewService creates a Service backed by the given Store.
func NewService(store *Store) *Service {
	return &Service{store: store}
}

// GetOrCreate returns the project for the given path, creating it if needed.
func (s *Service) GetOrCreate(path string) (Project, bool, error) {
	return s.store.GetOrCreate(path)
}

// SetProfile sets the profile ID for the project at the given path.
func (s *Service) SetProfile(path string, profileID string) error {
	return s.store.SetProfile(path, profileID)
}
