package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"safeagent/docker"

	"charm.land/huh/v2"
)

func startSession() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not determine working directory: %w", err)
	}

	// Ensure project exists
	proj, created, err := projectService.GetOrCreate(cwd)
	if err != nil {
		return err
	}
	if created {
		fmt.Printf("Created new project for %s\n", cwd)
	}

	// Check if profile is set
	if proj.ProfileID == "" {
		profiles, err := profileService.List()
		if err != nil {
			return err
		}

		if len(profiles) == 0 {
			fmt.Println("No profiles available. Create one first with: safeagent profiles create")
			return nil
		}

		options := make([]huh.Option[string], len(profiles))
		for i, p := range profiles {
			options[i] = huh.NewOption(profileService.FormatOption(p), p.ID)
		}

		var selectedID string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select a profile for this project").
					Options(options...).
					Value(&selectedID),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("profile selection cancelled: %w", err)
		}

		if err := projectService.SetProfile(cwd, selectedID); err != nil {
			return err
		}

		proj.ProfileID = selectedID
		fmt.Printf("Profile set for this project.\n")
	}

	// Get profile details
	prof, err := profileService.GetByID(proj.ProfileID)
	if err != nil {
		return fmt.Errorf("profile %s not found: %w", proj.ProfileID, err)
	}

	// Ensure Docker image exists for this profile
	imageName := fmt.Sprintf("safeagent-profile-%s:latest", prof.ID)
	imageExists, err := dockerService.ImageExists(imageName)
	if err != nil {
		return err
	}
	if !imageExists {
		fmt.Printf("Building Docker image for profile %q (node %s)...\n", prof.Name, prof.NodeVersion)
		buildArgs := map[string]string{
			"NODE_VERSION": prof.NodeVersion,
		}
		if err := dockerService.BuildImage(imageName, docker.Dockerfile, buildArgs); err != nil {
			return err
		}
		fmt.Println("Image built successfully.")
	}

	// Ensure container exists for this project
	if proj.ContainerID == "" {
		containerName := fmt.Sprintf("safeagent-%s", proj.ID)
		fmt.Printf("Creating container %s...\n", containerName)

		claudeConfigDir := filepath.Join(configDir, ".claude")
		mounts := []docker.Mount{
			{Source: cwd, Target: "/workspace"},
			{Source: claudeConfigDir, Target: "/claude"},
		}

		if err := dockerService.CreateContainer(containerName, imageName, mounts); err != nil {
			return err
		}

		if err := projectService.SetContainerID(cwd, containerName); err != nil {
			return err
		}

		proj.ContainerID = containerName
		fmt.Println("Container created.")
	} else {
		// Check if the saved container still exists
		exists, err := dockerService.ContainerExists(proj.ContainerID)
		if err != nil {
			return err
		}
		if !exists {
			// Container was deleted externally, recreate
			containerName := fmt.Sprintf("safeagent-%s", proj.ID)
			fmt.Printf("Container %s no longer exists, recreating...\n", proj.ContainerID)

			claudeConfigDir := filepath.Join(configDir, ".claude")
			mounts := []docker.Mount{
				{Source: cwd, Target: "/workspace"},
				{Source: claudeConfigDir, Target: "/claude"},
			}

			if err := dockerService.CreateContainer(containerName, imageName, mounts); err != nil {
				return err
			}

			if err := projectService.SetContainerID(cwd, containerName); err != nil {
				return err
			}

			proj.ContainerID = containerName
			fmt.Println("Container recreated.")
		}
	}

	// Start the container and exec claude
	if err := dockerService.StartContainer(proj.ContainerID); err != nil {
		return err
	}

	fmt.Printf("Starting Claude in %s with profile %q (node %s)...\n", cwd, prof.Name, prof.NodeVersion)
	return dockerService.Exec(proj.ContainerID, []string{"claude"})
}
