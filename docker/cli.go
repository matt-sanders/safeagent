package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// CLI implements Docker by shelling out to the `docker` binary.
type CLI struct{}

var _ Docker = CLI{}

// ImageExists checks if a Docker image exists locally.
func (CLI) ImageExists(ref string) (bool, error) {
	cmd := exec.Command("docker", "image", "inspect", ref)
	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check image %s: %w", ref, err)
	}
	return true, nil
}

// BuildImage builds a Docker image from an in-memory Dockerfile.
func (CLI) BuildImage(req BuildRequest) error {
	tmpDir, err := os.MkdirTemp("", "safeagent-build-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(req.Dockerfile), 0600); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	args := []string{"build", "-t", req.Tag, "-f", dockerfilePath}
	for _, k := range sortedKeys(req.BuildArgs) {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, req.BuildArgs[k]))
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

// Run starts an ephemeral container attached to the user's terminal. The
// container is removed when the command exits.
func (CLI) Run(req RunRequest) error {
	args := []string{"run", "--rm", "-it", "-w", req.Workdir}
	for _, k := range sortedKeys(req.Labels) {
		args = append(args, "--label", fmt.Sprintf("%s=%s", k, req.Labels[k]))
	}
	for _, m := range req.Mounts {
		if m.Source == "" {
			// Anonymous volume: shadows the bind mount at this path.
			args = append(args, "-v", m.Target)
		} else {
			args = append(args, "-v", fmt.Sprintf("%s:%s", m.Source, m.Target))
		}
	}
	args = append(args, req.Image)
	if len(req.Command) > 0 {
		args = append(args, "-ic", strings.Join(req.Command, " "))
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RemoveImage removes a Docker image.
func (CLI) RemoveImage(ref string) error {
	cmd := exec.Command("docker", "rmi", ref)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove image %s: %w", ref, err)
	}
	return nil
}

// ListImages returns every image reference in the safeagent repository.
func (CLI) ListImages() ([]string, error) {
	out, err := exec.Command("docker", "image", "ls",
		"--filter", "reference="+Repository+"/*",
		"--format", "{{.Repository}}:{{.Tag}}").Output()
	if err != nil {
		return nil, fmt.Errorf("docker image ls failed: %w", err)
	}
	var refs []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" || strings.HasSuffix(line, ":<none>") {
			continue
		}
		refs = append(refs, line)
	}
	sort.Strings(refs)
	return refs, nil
}

// ListContainers returns running containers labelled as safeagent v2 Sessions.
func (CLI) ListContainers() ([]Container, error) {
	out, err := exec.Command("docker", "ps",
		"--filter", "label="+LabelProject,
		"--format", "{{.ID}}\t{{.Image}}\t{{.RunningFor}}").Output()
	if err != nil {
		return nil, fmt.Errorf("docker ps failed: %w", err)
	}

	var containers []Container
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			continue
		}
		c := Container{ID: parts[0], Image: parts[1], RunningFor: parts[2]}

		cfgOut, err := exec.Command("docker", "inspect", "--format", "{{json .Config}}", c.ID).Output()
		if err != nil {
			// The container may have exited between ps and inspect.
			continue
		}
		var cfg struct {
			WorkingDir string
			Labels     map[string]string
		}
		if err := json.Unmarshal(cfgOut, &cfg); err != nil {
			continue
		}
		c.Workdir = cfg.WorkingDir
		c.Labels = cfg.Labels
		containers = append(containers, c)
	}
	return containers, nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
