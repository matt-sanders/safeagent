package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"safeagent/auth"
)

var authRmCmd = &cobra.Command{
	Use:   "rm <name>",
	Short: "Remove an auth identity",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if name == auth.DefaultName {
			return fmt.Errorf("cannot remove the %q identity", auth.DefaultName)
		}

		// Refuse to remove the identity the current project is using.
		if cwd, err := os.Getwd(); err == nil {
			if proj, found, err := projectService.FindForCwd(cwd); err == nil && found {
				if auth.Resolve(proj.AuthName) == name {
					return fmt.Errorf("identity %q is in use by the current project; switch with 'safeagent use-auth' first", name)
				}
			}
		}

		if err := authService.Remove(name); err != nil {
			return err
		}
		fmt.Printf("Identity %q removed.\n", name)
		return nil
	},
}

func init() {
	authCmd.AddCommand(authRmCmd)
}
