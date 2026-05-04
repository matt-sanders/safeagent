package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"safeagent/docker"
)

func startSession() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not determine working directory: %w", err)
	}

	// Ensure project exists
	proj, created, err := projectService.FindOrCreateForCwd(cwd)
	if err != nil {
		return err
	}
	if created {
		fmt.Printf("Created new project for %s\n", proj.Path)
	}

	// Check if profile is set
	if proj.ProfileID == "" {
		selected, profileID, err := promptAndSetProfile(proj.Path)
		if err != nil {
			return err
		}
		if !selected {
			return nil
		}
		proj.ProfileID = profileID
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

	relPath, err := filepath.Rel(proj.Path, cwd)
	if err != nil {
		return fmt.Errorf("could not determine relative path from project root %q to %q: %w", proj.Path, cwd, err)
	}
	if !filepath.IsLocal(relPath) && relPath != "." {
		return fmt.Errorf("cwd %q is not within project root %q", cwd, proj.Path)
	}

	workdir := filepath.Join("/workspace", relPath)
	mounts := buildMounts(proj.Path, relPath, proj.Exclusions)

	fmt.Printf("Starting Claude in %s with profile %q (node %s)...\n", cwd, prof.Name, prof.NodeVersion)
	return dockerService.Run(imageName, workdir, mounts, []string{"claude"})
}

func buildMounts(projectPath, cwdSubpath string, exclusions []string) []docker.Mount {
	mounts := []docker.Mount{
		{Source: projectPath, Target: "/workspace"},
		{Source: filepath.Join(configDir, ".claude"), Target: "/claude"},
	}

	// Excluded paths get anonymous volumes that shadow the bind mount,
	// rooted at the cwd subpath inside the workspace.
	for _, exc := range exclusions {
		mounts = append(mounts, docker.Mount{
			Target: filepath.Join("/workspace", cwdSubpath, exc),
		})
	}

	return mounts
}
