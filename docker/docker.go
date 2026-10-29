// Package docker is safeagent's narrow port onto Docker. Commands depend on the
// Docker interface; the binary wires in CLI (which shells out to `docker`) and
// tests wire in dockertest.Fake.
package docker

// Repository is the Docker repository every safeagent v2 Image lives in. An
// Image's reference is "safeagent/<stack-id>:<profile-id>".
const Repository = "safeagent"

// Labels set on every v2 Session container so `session ls` can identify them
// and fill in its columns without parsing image names.
const (
	LabelProject = "safeagent.v2.project"
	LabelProfile = "safeagent.v2.profile"
	LabelStack   = "safeagent.v2.stack"
)

// Docker is everything safeagent asks of Docker.
type Docker interface {
	// BuildImage builds an image from an in-memory Dockerfile.
	BuildImage(req BuildRequest) error
	// Run starts an interactive, ephemeral container attached to the terminal.
	Run(req RunRequest) error
	// RemoveImage removes an image by reference.
	RemoveImage(ref string) error
	// ListImages returns the references of every image in the safeagent
	// repository, e.g. "safeagent/<stack-id>:<profile-id>".
	ListImages() ([]string, error)
	// ListContainers returns running containers that carry the safeagent v2
	// Project label.
	ListContainers() ([]Container, error)
	// ImageExists reports whether an image with the reference exists locally.
	ImageExists(ref string) (bool, error)
}

// BuildRequest describes an image build.
type BuildRequest struct {
	Tag        string
	Dockerfile string
	BuildArgs  map[string]string
}

// Mount is a bind mount from host to container. An empty Source means an
// anonymous volume at Target (used to hide Exclusions).
type Mount struct {
	Source string
	Target string
}

// RunRequest describes a container run.
type RunRequest struct {
	Image   string
	Workdir string
	Mounts  []Mount
	Labels  map[string]string
	// Command is passed to the image's shell entrypoint as `-ic "<command>"`.
	// Empty means an interactive shell.
	Command []string
}

// Container is a running safeagent container.
type Container struct {
	ID         string
	Image      string
	Workdir    string
	Labels     map[string]string
	RunningFor string
}
