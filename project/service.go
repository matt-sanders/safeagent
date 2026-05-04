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

// ListByProfileID returns all projects using the given profile ID.
func (s *Service) ListByProfileID(profileID string) ([]Project, error) {
	projects, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	var result []Project
	for _, p := range projects {
		if p.ProfileID == profileID {
			result = append(result, p)
		}
	}
	return result, nil
}

// FindForCwd returns the closest ancestor project for cwd (or cwd itself).
// Returns ok=false if no project matches.
func (s *Service) FindForCwd(cwd string) (Project, bool, error) {
	return s.store.FindForCwd(cwd)
}

// FindOrCreateForCwd resolves a project for the given cwd: returns the closest
// ancestor project if one exists; otherwise creates a new project rooted at cwd.
// The bool return value is true only when a new project was created.
func (s *Service) FindOrCreateForCwd(cwd string) (Project, bool, error) {
	proj, found, err := s.store.FindForCwd(cwd)
	if err != nil {
		return Project{}, false, err
	}
	if found {
		return proj, false, nil
	}
	return s.store.GetOrCreate(cwd)
}
