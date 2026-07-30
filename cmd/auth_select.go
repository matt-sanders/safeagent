package cmd

import (
	"fmt"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"safeagent/auth"
)

var useAuthDetailStyle = lipgloss.NewStyle().Faint(true)

// promptAndSetAuth asks the user to pick an auth identity for the project at
// projectPath and persists the selection. The default identity is preselected.
// Returns the chosen identity name.
func promptAndSetAuth(projectPath string) (string, error) {
	identities, err := authService.List()
	if err != nil {
		return "", err
	}

	options := make([]huh.Option[string], len(identities))
	for i, id := range identities {
		label := id.Name
		if !id.LoggedIn {
			label = fmt.Sprintf("%s %s", id.Name, useAuthDetailStyle.Render("(not logged in)"))
		}
		options[i] = huh.NewOption(label, id.Name)
	}

	selected := auth.DefaultName
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select an auth identity for this project").
				Options(options...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return "", fmt.Errorf("identity selection cancelled: %w", err)
	}

	if err := projectService.SetAuth(projectPath, selected); err != nil {
		return "", err
	}

	fmt.Printf("Identity set for this project.\n")
	return selected, nil
}
