// Package dockertest provides a recording, in-memory fake of docker.Docker.
package dockertest

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"safeagent/docker"
)

// Fake records every call and keeps an in-memory set of images and running
// containers. Preload state with AddImage / AddContainer; inspect what
// safeagent asked for via Builds, Runs and Removed.
type Fake struct {
	images     map[string]bool
	containers []docker.Container

	Builds  []docker.BuildRequest
	Runs    []docker.RunRequest
	Removed []string

	// Errors injected per operation, returned instead of acting.
	BuildErr  error
	RunErr    error
	RemoveErr error
}

var _ docker.Docker = (*Fake)(nil)

// New returns an empty Fake.
func New() *Fake {
	return &Fake{images: map[string]bool{}}
}

// AddImage preloads an image reference.
func (f *Fake) AddImage(ref string) {
	f.images[ref] = true
}

// AddContainer preloads a running container.
func (f *Fake) AddContainer(c docker.Container) {
	f.containers = append(f.containers, c)
}

// HasImage reports whether the image currently exists in the fake.
func (f *Fake) HasImage(ref string) bool {
	return f.images[ref]
}

// Images returns every image reference currently in the fake, sorted.
func (f *Fake) Images() []string {
	return slices.Sorted(maps.Keys(f.images))
}

// BuildImage records the build and adds the tag to the image set.
func (f *Fake) BuildImage(req docker.BuildRequest) error {
	req.BuildArgs = maps.Clone(req.BuildArgs)
	f.Builds = append(f.Builds, req)
	if f.BuildErr != nil {
		return f.BuildErr
	}
	f.images[req.Tag] = true
	return nil
}

// Run records the run.
func (f *Fake) Run(req docker.RunRequest) error {
	req.Mounts = slices.Clone(req.Mounts)
	req.Labels = maps.Clone(req.Labels)
	req.Command = slices.Clone(req.Command)
	f.Runs = append(f.Runs, req)
	return f.RunErr
}

// RemoveImage records the removal and drops the image. Removing a missing
// image is an error, as with `docker rmi`.
func (f *Fake) RemoveImage(ref string) error {
	f.Removed = append(f.Removed, ref)
	if f.RemoveErr != nil {
		return f.RemoveErr
	}
	if !f.images[ref] {
		return fmt.Errorf("no such image: %s", ref)
	}
	delete(f.images, ref)
	return nil
}

// ListImages returns the images in the safeagent repository, sorted.
func (f *Fake) ListImages() ([]string, error) {
	var refs []string
	for _, ref := range f.Images() {
		if strings.HasPrefix(ref, docker.Repository+"/") {
			refs = append(refs, ref)
		}
	}
	return refs, nil
}

// ListContainers returns preloaded containers carrying the v2 Project label.
func (f *Fake) ListContainers() ([]docker.Container, error) {
	var out []docker.Container
	for _, c := range f.containers {
		if _, ok := c.Labels[docker.LabelProject]; ok {
			out = append(out, c)
		}
	}
	return out, nil
}

// ImageExists reports whether the image is in the fake.
func (f *Fake) ImageExists(ref string) (bool, error) {
	return f.images[ref], nil
}
