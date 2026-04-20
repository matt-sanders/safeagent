package cmd

import (
	"fmt"

	"safeagent/docker"

	"github.com/spf13/cobra"
)

var profilesRebuildCmd = &cobra.Command{
	Use:   "rebuild [profile-id]",
	Short: "Rebuild the Docker image for a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileID := args[0]

		// Validate profile exists
		prof, err := profileService.GetByID(profileID)
		if err != nil {
			return fmt.Errorf("profile %q not found", profileID)
		}

		imageName := fmt.Sprintf("safeagent-profile-%s:latest", prof.ID)

		// Find all projects using this profile and remove their containers
		projects, err := projectService.ListByProfileID(prof.ID)
		if err != nil {
			return err
		}

		for _, proj := range projects {
			if err := removeProjectContainer(proj.Path, proj.ContainerID); err != nil {
				return err
			}
		}

		// Rebuild the image (replaces the existing tag)
		fmt.Printf("Rebuilding image for profile %q (node %s)...\n", prof.Name, prof.NodeVersion)
		buildArgs := map[string]string{
			"NODE_VERSION": prof.NodeVersion,
		}
		if err := dockerService.BuildImage(imageName, docker.Dockerfile, buildArgs); err != nil {
			return err
		}

		fmt.Printf("Profile %q image rebuilt successfully.\n", prof.Name)
		return nil
	},
}

func init() {
	profilesCmd.AddCommand(profilesRebuildCmd)
}
