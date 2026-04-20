package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// ContainerExists checks if a Docker container exists (running or stopped).
func (s *Service) ContainerExists(containerID string) (bool, error) {
	cmd := exec.Command("docker", "container", "inspect", containerID)
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check container %s: %w", containerID, err)
	}
	return true, nil
}

// CreateContainer creates a new Docker container that stays alive for exec.
func (s *Service) CreateContainer(name string, imageName string, mounts []Mount) error {
	args := []string{"create", "--name", name, "-it"}
	for _, m := range mounts {
		if m.Source == "" {
			// Anonymous volume — shadows the bind mount at this path
			args = append(args, "-v", m.Target)
		} else {
			args = append(args, "-v", fmt.Sprintf("%s:%s", m.Source, m.Target))
		}
	}
	args = append(args, imageName, "sleep", "infinity")

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker create failed: %w", err)
	}

	return nil
}

// StopContainer stops a running container.
// Does nothing if the container is already stopped.
func (s *Service) StopContainer(containerID string) error {
	cmd := exec.Command("docker", "stop", containerID)
	cmd.Stdout = nil
	cmd.Stderr = nil
	// Ignore error if container is already stopped
	cmd.Run()
	return nil
}

// RemoveContainer removes a container. Returns an error if removal fails.
func (s *Service) RemoveContainer(containerID string) error {
	cmd := exec.Command("docker", "rm", containerID)
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove container %s: %w", containerID, err)
	}
	return nil
}

// StartContainer starts a container in detached mode.
// Does nothing if the container is already running.
func (s *Service) StartContainer(containerID string) error {
	cmd := exec.Command("docker", "start", containerID)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker start failed: %w", err)
	}
	return nil
}

// Exec runs a command inside a running container, attached to the user's terminal.
func (s *Service) Exec(containerID string, command []string) error {
	args := []string{"exec", "-it", "-w", "/workspace", containerID, "/bin/zsh", "-ic"}
	args = append(args, command...)

	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
