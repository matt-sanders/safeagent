// Package stack manages Stacks: the "what" of a session. Records live in
// stacks.json under the v2 root, and each Stack has a stacks/<id>/ folder
// holding its Dockerfile fragment.
package stack

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"safeagent/ident"
)

// DefaultNode is the node version offered when creating a Stack.
const DefaultNode = "lts"

// Stack is one Stack record. Node is fixed once created.
type Stack struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Node string `json:"node"`
}

// Store is the file-backed list of Stacks under a v2 root.
type Store struct {
	root string
}

// NewStore returns a Store rooted at the v2 config root.
func NewStore(root string) *Store {
	return &Store{root: root}
}

func (s *Store) recordsPath() string {
	return filepath.Join(s.root, "stacks.json")
}

// Dir is the folder for the Stack with the given ID.
func (s *Store) Dir(id string) string {
	return filepath.Join(s.root, "stacks", id)
}

// DockerfilePath is the Dockerfile fragment for the Stack with the given ID.
func (s *Store) DockerfilePath(id string) string {
	return filepath.Join(s.Dir(id), "Dockerfile")
}

// List returns every Stack in creation order.
func (s *Store) List() ([]Stack, error) {
	data, err := os.ReadFile(s.recordsPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading stacks: %w", err)
	}
	var stacks []Stack
	if err := json.Unmarshal(data, &stacks); err != nil {
		return nil, fmt.Errorf("parsing stacks: %w", err)
	}
	return stacks, nil
}

// FindByName returns the Stack with the given Name, or an error naming it if
// there is none.
func (s *Store) FindByName(name string) (Stack, error) {
	stacks, err := s.List()
	if err != nil {
		return Stack{}, err
	}
	for _, st := range stacks {
		if st.Name == name {
			return st, nil
		}
	}
	return Stack{}, fmt.Errorf("no stack named %q", name)
}

// Create validates name, adds a new Stack with a generated ID and creates its
// folder with an empty Dockerfile fragment. An empty node means DefaultNode.
func (s *Store) Create(name, node string) (Stack, error) {
	if err := ident.ValidateName(name); err != nil {
		return Stack{}, err
	}
	if node == "" {
		node = DefaultNode
	}
	stacks, err := s.List()
	if err != nil {
		return Stack{}, err
	}
	for _, st := range stacks {
		if st.Name == name {
			return Stack{}, fmt.Errorf("a stack named %q already exists", name)
		}
	}

	st := Stack{ID: ident.NewID(), Name: name, Node: node}
	if err := os.MkdirAll(s.Dir(st.ID), 0o700); err != nil {
		return Stack{}, fmt.Errorf("creating stack folder: %w", err)
	}
	if err := os.WriteFile(s.DockerfilePath(st.ID), nil, 0o600); err != nil {
		return Stack{}, fmt.Errorf("creating stack Dockerfile: %w", err)
	}
	if err := s.save(append(stacks, st)); err != nil {
		return Stack{}, err
	}
	return st, nil
}

func (s *Store) save(stacks []Stack) error {
	data, err := json.MarshalIndent(stacks, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(s.root, 0o700); err != nil {
		return fmt.Errorf("creating config root: %w", err)
	}
	if err := os.WriteFile(s.recordsPath(), data, 0o600); err != nil {
		return fmt.Errorf("writing stacks: %w", err)
	}
	return nil
}

// FindByID returns the Stack with the given ID, or an error if no Stack has
// it (e.g. it was removed after a Project chose it).
func (s *Store) FindByID(id string) (Stack, error) {
	stacks, err := s.List()
	if err != nil {
		return Stack{}, err
	}
	for _, st := range stacks {
		if st.ID == id {
			return st, nil
		}
	}
	return Stack{}, fmt.Errorf("no stack with ID %q", id)
}

// Remove deletes the Stack record with the given ID and its folder.
func (s *Store) Remove(id string) error {
	stacks, err := s.List()
	if err != nil {
		return err
	}
	kept := make([]Stack, 0, len(stacks))
	for _, st := range stacks {
		if st.ID != id {
			kept = append(kept, st)
		}
	}
	if len(kept) == len(stacks) {
		return fmt.Errorf("no stack with ID %q", id)
	}
	if err := s.save(kept); err != nil {
		return err
	}
	if err := os.RemoveAll(s.Dir(id)); err != nil {
		return fmt.Errorf("removing stack folder: %w", err)
	}
	return nil
}
