package cmd

import "fmt"

// removeProjectContainer stops and removes the container for a project, clearing its ID.
// Does nothing if the project has no container.
func removeProjectContainer(cwd string, containerID string) error {
	if containerID == "" {
		return nil
	}

	exists, err := dockerService.ContainerExists(containerID)
	if err != nil {
		return err
	}
	if exists {
		fmt.Printf("Removing container %s...\n", containerID)
		dockerService.StopContainer(containerID)
		if err := dockerService.RemoveContainer(containerID); err != nil {
			return fmt.Errorf("failed to remove container: %w", err)
		}
	}

	return projectService.SetContainerID(cwd, "")
}
