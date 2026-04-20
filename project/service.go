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

// GetExclusions returns the exclusions for the project at the given path.
func (s *Service) GetExclusions(path string) ([]string, error) {
	proj, _, err := s.store.GetOrCreate(path)
	if err != nil {
		return nil, err
	}
	if proj.Exclusions == nil {
		return []string{}, nil
	}
	return proj.Exclusions, nil
}

// AddExclusion adds an exclusion to the project at the given path.
func (s *Service) AddExclusion(path string, exclusion string) error {
	return s.store.AddExclusion(path, exclusion)
}

// RemoveExclusion removes an exclusion from the project at the given path.
func (s *Service) RemoveExclusion(path string, exclusion string) error {
	return s.store.RemoveExclusion(path, exclusion)
}

// SetContainerID sets the container ID for the project at the given path.
func (s *Service) SetContainerID(path string, containerID string) error {
	return s.store.SetContainerID(path, containerID)
}
