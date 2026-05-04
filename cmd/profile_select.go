package cmd

import (
	"fmt"

	"charm.land/huh/v2"
)

// promptAndSetProfile asks the user to pick a profile for the project at projectPath
// and persists the selection. If no profiles exist, prints a hint and returns
// selected=false without an error.
func promptAndSetProfile(projectPath string) (selected bool, profileID string, err error) {
	profiles, err := profileService.List()
	if err != nil {
		return false, "", err
	}

	if len(profiles) == 0 {
		fmt.Println("No profiles available. Create one first with: safeagent profiles create")
		return false, "", nil
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
		return false, "", fmt.Errorf("profile selection cancelled: %w", err)
	}

	if err := projectService.SetProfile(projectPath, selectedID); err != nil {
		return false, "", err
	}

	fmt.Printf("Profile set for this project.\n")
	return true, selectedID, nil
}
