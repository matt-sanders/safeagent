package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Mount represents a bind mount from host to container.
type Mount struct {
	Source string
	Target string
}

// Service wraps Docker CLI commands.
type Service struct{}

// NewService creates a new Docker Service.
func NewService() *Service {
	return &Service{}
}

// ImageExists checks if a Docker image exists locally.
func (s *Service) ImageExists(imageName string) (bool, error) {
	cmd := exec.Command("docker", "image", "inspect", imageName)
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check image %s: %w", imageName, err)
	}
	return true, nil
}

// BuildImage builds a Docker image from a dockerfile string.
func (s *Service) BuildImage(imageName string, dockerfile string, buildArgs map[string]string) error {
	tmpDir, err := os.MkdirTemp("", "safeagent-build-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0600); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	args := []string{"build", "-t", imageName, "-f", dockerfilePath}
	for k, v := range buildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, v))
	}
	args = append(args, tmpDir)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	return nil
}

// Run starts an ephemeral container that runs the given command interactively
// attached to the user's terminal, with the container's working directory set
// to workdir. The container is removed when the command exits.
func (s *Service) Run(imageName, workdir string, mounts []Mount, command []string) error {
	args := []string{"run", "--rm", "-it", "-w", workdir}
	for _, m := range mounts {
		if m.Source == "" {
			// Anonymous volume — shadows the bind mount at this path
			args = append(args, "-v", m.Target)
		} else {
			args = append(args, "-v", fmt.Sprintf("%s:%s", m.Source, m.Target))
		}
	}
	args = append(args, imageName)
	if len(command) > 0 {
		args = append(args, "-ic", strings.Join(command, " "))
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// RunningSession holds information about a running safeagent container.
type RunningSession struct {
	ContainerID string
	ImageName   string
	WorkingDir  string
	RunningFor  string
}

// ListRunningSessions returns all running containers started by safeagent.
func (s *Service) ListRunningSessions() ([]RunningSession, error) {
	out, err := exec.Command("docker", "ps", "--format", "{{.ID}}\t{{.Image}}\t{{.RunningFor}}").Output()
	if err != nil {
		return nil, fmt.Errorf("docker ps failed: %w", err)
	}

	var sessions []RunningSession
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			continue
		}
		id, image, runningFor := parts[0], parts[1], parts[2]
		if !strings.HasPrefix(image, "safeagent-profile-") {
			continue
		}

		wdOut, err := exec.Command("docker", "inspect", "--format", "{{.Config.WorkingDir}}", id).Output()
		if err != nil {
			continue
		}
		sessions = append(sessions, RunningSession{
			ContainerID: id,
			ImageName:   image,
			WorkingDir:  strings.TrimSpace(string(wdOut)),
			RunningFor:  runningFor,
		})
	}
	return sessions, nil
}

// RemoveImage removes a Docker image. Returns an error if removal fails.
func (s *Service) RemoveImage(imageName string) error {
	cmd := exec.Command("docker", "rmi", imageName)
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove image %s: %w", imageName, err)
	}
	return nil
}

