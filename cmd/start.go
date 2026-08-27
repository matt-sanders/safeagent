package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"safeagent/auth"
	"safeagent/docker"
)

func startSession(shellMode bool) error {
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

	// Resolve the auth identity: prompt only for brand-new projects; existing
	// projects with no saved identity resolve to the default silently.
	if created {
		name, err := promptAndSetAuth(proj.Path)
		if err != nil {
			return err
		}
		proj.AuthName = name
	}
	authName := auth.Resolve(proj.AuthName)
	if err := authService.EnsureDir(authName); err != nil {
		return err
	}
	claudeDir := authService.DirFor(authName)

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
			"BUN_VERSION":  prof.BunVersion,
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

	authCfg, err := authService.LoadConfig(authName)
	if err != nil {
		return err
	}

	mounts := buildMounts(proj.Path, cwd, claudeDir, proj.Exclusions, authCfg.Mounts)

	var command []string
	if shellMode {
		fmt.Printf("Opening shell in %s with profile %q (node %s) as %q...\n", cwd, prof.Name, prof.NodeVersion, authName)
	} else {
		fmt.Printf("Starting Claude in %s with profile %q (node %s) as %q...\n", cwd, prof.Name, prof.NodeVersion, authName)
		command = []string{"claude", "--dangerously-skip-permissions"}
	}
	return dockerService.Run(imageName, cwd, mounts, command)
}

func buildMounts(projectPath, cwd, claudeDir string, exclusions []string, authMounts []auth.Mount) []docker.Mount {
	mounts := []docker.Mount{
		{Source: projectPath, Target: projectPath},
		{Source: claudeDir, Target: "/claude"},
	}

	for _, m := range authMounts {
		mounts = append(mounts, docker.Mount{Source: m.Source, Target: m.Target})
	}

	// Excluded paths get anonymous volumes that shadow the bind mount,
	// rooted at the cwd inside the container (which matches the host cwd).
	for _, exc := range exclusions {
		mounts = append(mounts, docker.Mount{
			Target: filepath.Join(cwd, exc),
		})
	}

	return mounts
}
