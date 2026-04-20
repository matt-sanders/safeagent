package cmd

import (
	"fmt"
	"os"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
)

var (
	useProfileDetailStyle = lipgloss.NewStyle().Faint(true)
)

var useProfileCmd = &cobra.Command{
	Use:   "use-profile",
	Short: "Change or set the profile for the current project",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not determine working directory: %w", err)
		}

		// Ensure project exists
		proj, _, err := projectService.GetOrCreate(cwd)
		if err != nil {
			return err
		}

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
			prefix := "  "
			if p.ID == proj.ProfileID {
				prefix = "* "
			}
			label := fmt.Sprintf("%s%s %s", prefix, p.Name, useProfileDetailStyle.Render("node "+p.NodeVersion))
			options[i] = huh.NewOption(label, p.ID)
		}

		var selectedID string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select a profile").
					Options(options...).
					Value(&selectedID),
			),
		)

		if err := form.Run(); err != nil {
			return nil
		}

		// Remove existing container if there is one
		if proj.ContainerID != "" {
			exists, err := dockerService.ContainerExists(proj.ContainerID)
			if err != nil {
				return err
			}
			if exists {
				fmt.Printf("Removing container %s...\n", proj.ContainerID)
				dockerService.StopContainer(proj.ContainerID)
				if err := dockerService.RemoveContainer(proj.ContainerID); err != nil {
					return fmt.Errorf("failed to remove container: %w", err)
				}
			}
			if err := projectService.SetContainerID(cwd, ""); err != nil {
				return err
			}
		}

		if err := projectService.SetProfile(cwd, selectedID); err != nil {
			return err
		}

		p, err := profileService.GetByID(selectedID)
		if err != nil {
			return err
		}

		fmt.Printf("Now using profile %q (node %s)\n", p.Name, p.NodeVersion)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(useProfileCmd)
}
