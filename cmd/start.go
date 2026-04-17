package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"safe-claude/project"

	"charm.land/huh/v2"
)

func startSession() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not determine working directory: %w", err)
	}

	projectsPath := filepath.Join(homeDir, ".safe-claude", "projects.json")

	// Ensure project exists
	proj, created, err := project.GetOrCreate(projectsPath, cwd)
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
			fmt.Println("No profiles available. Create one first with: safe-claude profiles create")
			return nil
		}

		// Build select options using shared display formatting
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

		if err := project.SetProfile(projectsPath, cwd, selectedID); err != nil {
			return err
		}

		proj.ProfileID = selectedID
		fmt.Printf("Profile set for this project.\n")
	}

	// Find the profile name for display
	p, err := profileService.GetByID(proj.ProfileID)
	if err != nil {
		fmt.Printf("Starting session in %s (profile %s not found)\n", cwd, proj.ProfileID)
		return nil
	}

	fmt.Printf("Starting session in %s with profile %q (node %s)\n", cwd, p.Name, p.NodeVersion)
	return nil
}
