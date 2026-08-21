package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var profilesDockerfileCmd = &cobra.Command{
	Use:   "dockerfile <profile-id>",
	Short: "Edit the extra Dockerfile fragment for a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileID := args[0]

		prof, err := profileService.GetByID(profileID)
		if err != nil {
			return fmt.Errorf("profile %q not found", profileID)
		}

		path := profileService.DockerfileExtraPath(prof.ID)

		// Seed the file with a starter comment if it doesn't exist yet.
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				return fmt.Errorf("failed to create profile directory: %w", err)
			}
			seed := "# Extra Dockerfile instructions for profile " + prof.Name + "\n" +
				"# These lines are inserted into the image build just before ENTRYPOINT.\n" +
				"# The current user is 'developer'. Example:\n" +
				"#   RUN sudo apt-get install -y ruby\n"
			if err := os.WriteFile(path, []byte(seed), 0600); err != nil {
				return fmt.Errorf("failed to create dockerfile fragment: %w", err)
			}
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}

		editorCmd := exec.Command(editor, path)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr
		if err := editorCmd.Run(); err != nil {
			return fmt.Errorf("editor exited with error: %w", err)
		}

		fmt.Printf("Saved. Run `safeagent profiles rebuild %s` to apply changes.\n", prof.ID)
		return nil
	},
}

func init() {
	profilesCmd.AddCommand(profilesDockerfileCmd)
}
