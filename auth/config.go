package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Mount is a host-to-container bind mount stored in an identity config.
type Mount struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// Config holds per-identity settings persisted to config.json.
type Config struct {
	Mounts []Mount `json:"mounts,omitempty"`
}

func (s *Service) configPath(name string) string {
	return filepath.Join(s.root, name, "config.json")
}

// LoadConfig returns the config for the named identity.
// Returns an empty Config if none exists yet.
func (s *Service) LoadConfig(name string) (Config, error) {
	data, err := os.ReadFile(s.configPath(name))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("failed to read config for identity %q: %w", name, err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse config for identity %q: %w", name, err)
	}
	return cfg, nil
}

func (s *Service) saveConfig(name string, cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config for identity %q: %w", name, err)
	}
	if err := os.WriteFile(s.configPath(name), data, 0600); err != nil {
		return fmt.Errorf("failed to write config for identity %q: %w", name, err)
	}
	return nil
}

// AddMount appends a bind mount to the named identity's config.
// Returns an error if a mount with the same target already exists.
func (s *Service) AddMount(name, source, target string) error {
	cfg, err := s.LoadConfig(name)
	if err != nil {
		return err
	}
	for _, m := range cfg.Mounts {
		if m.Target == target {
			return fmt.Errorf("mount with target %q already exists", target)
		}
	}
	cfg.Mounts = append(cfg.Mounts, Mount{Source: source, Target: target})
	return s.saveConfig(name, cfg)
}

// RemoveMount removes the mount with the given target from the named identity's config.
// Returns an error if no such mount exists.
func (s *Service) RemoveMount(name, target string) error {
	cfg, err := s.LoadConfig(name)
	if err != nil {
		return err
	}
	for i, m := range cfg.Mounts {
		if m.Target == target {
			cfg.Mounts = append(cfg.Mounts[:i], cfg.Mounts[i+1:]...)
			return s.saveConfig(name, cfg)
		}
	}
	return fmt.Errorf("mount with target %q not found", target)
}
